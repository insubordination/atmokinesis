package scheduler

import (
	"bufio"
	"bytes"
	"sync"
	"time"
)

type BaseContext struct {
	executionDate   time.Time
	startDate       time.Time
	nextRunDate     time.Time
	previousRunDate time.Time
	notifySubTasks  chan bool
	subTaskStream   chan interface{}
	logWriteSyncer  WriteSyncer
	*sync.RWMutex
}

type BaseWriteSyncer struct {
	*bufio.Writer
	*bytes.Buffer
	*sync.RWMutex
}

func NewBaseWriteSyncer() (*bytes.Buffer, WriteSyncer) {
	var logs []byte
	var bw = bytes.NewBuffer(logs)
	var writer = bufio.NewWriter(bw)
	return bw, &BaseWriteSyncer{
		Buffer:  bw,
		Writer:  writer,
		RWMutex: new(sync.RWMutex),
	}
}

func (bws *BaseWriteSyncer) Write(p []byte) (n int, err error) {
	bws.Lock()
	defer bws.Unlock()
	return bws.Writer.Write(p)
}

func (bws *BaseWriteSyncer) Sync() error {
	bws.Writer.Flush()
	return nil
}

func NewBaseContext(executionDate time.Time, startDate time.Time, nextRunDate time.Time, previousRunDate time.Time, subTaskStream chan interface{}, syncer WriteSyncer) (Context, chan bool, chan interface{}) {
	var notifySubTasks = make(chan bool, 1)

	if subTaskStream == nil {
		subTaskStream = make(chan interface{}, 100)
	}
	return &BaseContext{
		executionDate:   executionDate,
		startDate:       startDate,
		nextRunDate:     nextRunDate,
		previousRunDate: previousRunDate,
		notifySubTasks:  notifySubTasks,
		subTaskStream:   subTaskStream,
		logWriteSyncer:  syncer,
		RWMutex:         new(sync.RWMutex),
	}, notifySubTasks, subTaskStream
}

func (b BaseContext) ExecutionDate() time.Time {
	return b.executionDate
}

func (b BaseContext) StartDate() time.Time {
	return b.startDate
}

func (b BaseContext) NextRunDate() time.Time {
	return b.nextRunDate
}

func (b BaseContext) PreviousRunDate() time.Time {
	return b.previousRunDate
}

func (b BaseContext) NotifySubTasks() {
	b.notifySubTasks <- true
}

func (b BaseContext) StreamToSubTasks(out interface{}) {
	b.subTaskStream <- out
}

func (b BaseContext) LogWriteSyncer() WriteSyncer {
	return b.logWriteSyncer
}
