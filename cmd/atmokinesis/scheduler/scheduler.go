package scheduler

import (
	"context"
	"errors"
	"log"
	"time"
)

var sch *scheduler
var schTaskBuffer = make(chan Task, 5000)

type scheduler struct {
	atmo *Atmo
}

func InitScheduler(db Store) (err error) {
	sch = &scheduler{NewCron()}
	sch.atmo.Start()

	go func() {
		var ticker = time.NewTicker(1500 * time.Millisecond)
		for {
			select {
			case t := <-schTaskBuffer:
				sched, pErr := ParseStandard(t.Schedule())
				if pErr != nil {
					errors.As(pErr, &err)
				}
				sch.atmo.Schedule(sched, t)
			case <-ticker.C:
				if len(schTaskBuffer) == 0 && len(sch.atmo.entries) > 0 {
					if updateErr := db.UpdateInMemoryEntriesFromStorage(context.TODO(), sch.atmo.entries); updateErr != nil {
						log.Printf("failed to update (entry|entries): %s", updateErr.Error())
						errors.As(updateErr, &err)
						return
					}
					return
				}
			}
		}
	}()

	return nil
}

func StopScheduler(db Store) error {
	sch.atmo.Stop()
	ctx := context.TODO()
	defer db.Close(ctx)
	return db.UpdateEntries(ctx, sch.atmo.entries)
}

func ScheduleTask(t Task) {
	if sch == nil {
		schTaskBuffer <- t
	}
}

func TaskList() []DisplayTask {
	var taskList []DisplayTask
	entries := sch.atmo.entrySnapshot()
	for _, e := range entries {
		var lastRun time.Time
		if len(e.History) > 0 {
			lastRun = e.History[len(e.History)-1].ExecutionTime
		}
		taskList = append(taskList, DisplayTask{
			ID:       string(e.Task.TaskID()),
			Status:   string(e.Status),
			Schedule: e.Task.Schedule(),
			NextRun:  e.Next,
			LastRun:  lastRun,
			History:  e.History,
		})
	}
	return taskList
}
