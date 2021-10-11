package scheduler

import (
	"fmt"
	"log"
	"runtime"
	"sort"
	"sync"
	"time"
)

// Atmo keeps track of any number of entries, invoking the associated func as
// specified by the schedule. It may be started, stopped, and the entries may
// be inspected while running.
type Atmo struct {
	entries  []*Entry
	stop     chan struct{}
	add      chan *Entry
	snapshot chan []*Entry
	running  bool
	ErrorLog *log.Logger
	location *time.Location
}

// The Schedule describes a job's duty cycle.
type Schedule interface {
	// Return the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}

type TaskHistory struct {
	ExecutionTime time.Time   `json:"execution_time"`
	Status        EntryStatus `json:"status,omitempty"`
	Logs          string      `json:"logs,omitempty"`
}

// byTime is a wrapper for sorting the entry array by time
// (with zero time at the end).
type byTime []*Entry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

// NewCron returns a new Atmo job runner, in the Local time zone.
func NewCron() *Atmo {
	return NewWithLocation(time.Now().Location())
}

// NewWithLocation returns a new Atmo job runner.
func NewWithLocation(location *time.Location) *Atmo {
	return &Atmo{
		entries:  []*Entry{},
		add:      make(chan *Entry),
		stop:     make(chan struct{}),
		snapshot: make(chan []*Entry),
		running:  false,
		ErrorLog: nil,
		location: location,
	}
}

// AddJob adds a Task to the Atmo to be run on the given schedule.
func (c *Atmo) AddTask(cron Cron, task Task) error {
	schedule, err := Parse(cron)
	if err != nil {
		return err
	}
	c.Schedule(schedule, task)
	return nil
}

// Schedule adds a Task to the Atmo to be run on the given schedule.
func (c *Atmo) Schedule(schedule Schedule, task Task) {
	entry := &Entry{
		Schedule: schedule,
		Status:   PendingRun,
		Task:     task,
		Errors:   make(map[time.Time]error),
		RWMutex:  new(sync.RWMutex),
	}
	if !c.running {
		c.entries = append(c.entries, entry)
		return
	}

	c.add <- entry
}

// Entries returns a snapshot of the atmo entries.
func (c *Atmo) Entries() []*Entry {
	if c.running {
		c.snapshot <- nil
		x := <-c.snapshot
		return x
	}
	return c.entrySnapshot()
}

// Location gets the time zone location
func (c *Atmo) Location() *time.Location {
	return c.location
}

// Start the atmo scheduler in its own go-routine, or no-op if already started.
func (c *Atmo) Start() {
	if c.running {
		return
	}
	c.running = true
	go c.run()
}

// Run the atmo scheduler, or no-op if already running.
func (c *Atmo) Run() {
	if c.running {
		return
	}
	c.running = true
	c.run()
}

func (c *Atmo) runWithRecovery(ctx Context, e *Entry, notify chan bool, parentStream chan interface{}) {
	var stream chan interface{}
	var buffer, logWriter = NewBaseWriteSyncer()
	executionTime := time.Now()

	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			e.History = append(e.History, &TaskHistory{
				ExecutionTime: executionTime,
				Status:        Failing,
				Logs:          string(buffer.Bytes()),
			})
			e.Errors[e.Prev] = fmt.Errorf("task is panicing: %s", string(buf))
		}
	}()

	if notify != nil {
		<-notify
	}

	if e.Task.ScheduleOptions().EndDate().Before(executionTime) {
		return
	}

	if ctx == nil {
		ctx, notify, stream = NewBaseContext(executionTime, time.Now(), e.Next, e.Prev, parentStream, logWriter)
	}

	log.Printf("[%s] started", e.Task.TaskID().ToString())
	defer log.Printf("[%s] finished", e.Task.TaskID().ToString())
	e.ChangeStatus(Running)
	if err := e.Task.Run(ctx); err != nil {
		e.ChangeStatus(Failing)
		logWriter.Sync()
		e.History = append(e.History, &TaskHistory{
			ExecutionTime: executionTime,
			Status:        Success,
			Logs:          string(buffer.Bytes()),
		})
		e.Errors[executionTime] = err
	} else {
		e.ChangeStatus(PendingRun)
		logWriter.Sync()
		e.History = append(e.History, &TaskHistory{
			ExecutionTime: executionTime,
			Status:        Success,
			Logs:          string(buffer.Bytes()),
		})
	}

	if isParallel, subTasks := e.Task.SubTasks(); len(subTasks) > 0 {
		<-notify
		for _, subTask := range subTasks {
			executionTime = time.Now()
			bc, ny, st := NewBaseContext(executionTime, time.Now(), e.Next, e.Prev, stream, logWriter)
			se := c.entryByTask(subTask)
			c.runWithRecovery(bc, se, ny, st)
			if !isParallel {
				<-ny
			}
		}
	}
}

func (c *Atmo) entryByTask(task Task) *Entry {
	for i, e := range c.entries {
		if e.Task.TaskID() == task.TaskID() {
			return c.entries[i]
		}
	}
	return nil
}

// Run the scheduler. this is private just due to the need to synchronize
// access to the 'running' state variable.
func (c *Atmo) run() {
	// Figure out the next activation times for each entry.
	now := c.now()
	for _, entry := range c.entries {
		entry.Next = entry.Schedule.Next(now)
	}

	for {
		// Determine the next entry to run.
		sort.Sort(byTime(c.entries))

		var timer *time.Timer
		if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			timer = time.NewTimer(100000 * time.Hour)
		} else {
			timer = time.NewTimer(c.entries[0].Next.Sub(now))
		}

		for {
			select {
			case now = <-timer.C:
				now = now.In(c.location)
				// Run every entry whose next time was less than now
				for _, e := range c.entries {
					if e.Next.After(now) || e.Next.IsZero() {
						break
					}

					if e.Notify == nil {
						notify := make(chan bool, 1)
						e.Notify = notify
						e.Notify <- true
					}

					go func() {
						if e.Task.ScheduleOptions().AllowOverlap() {
							c.runWithRecovery(nil, e, nil, nil)
						} else {
							<-e.Notify
							c.runWithRecovery(nil, e, nil, nil)
							e.Notify <- true
						}
					}()

					e.Prev = e.Next
					e.Next = e.Schedule.Next(now)
				}

			case newEntry := <-c.add:
				timer.Stop()
				now = c.now()
				newEntry.Next = newEntry.Schedule.Next(now)
				c.entries = append(c.entries, newEntry)

			case <-c.snapshot:
				c.snapshot <- c.entrySnapshot()
				continue

			case <-c.stop:
				timer.Stop()
				return
			}

			break
		}
	}
}

// Logs an error to stderr or to the configured error log
func (c *Atmo) logf(format string, args ...interface{}) {
	if c.ErrorLog != nil {
		c.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// Stop stops the atmo scheduler if it is running; otherwise it does nothing.
func (c *Atmo) Stop() {
	if !c.running {
		return
	}
	c.stop <- struct{}{}
	c.running = false
}

// entrySnapshot returns a copy of the current atmo entry list.
func (c *Atmo) entrySnapshot() []*Entry {
	entries := []*Entry{}
	for _, e := range c.entries {
		entries = append(entries, &Entry{
			Schedule: e.Schedule,
			Status:   e.Status,
			Next:     e.Next,
			Prev:     e.Prev,
			Task:     e.Task,
			History:  e.History,
			Errors:   e.Errors,
			RWMutex:  new(sync.RWMutex),
		})
	}
	return entries
}

// now returns current time in c location
func (c *Atmo) now() time.Time {
	return time.Now().In(c.location)
}
