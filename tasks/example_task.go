package tasks

import (
	"github.com/insubordination/atmokinesis/cmd/atmokinesis/scheduler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func init() {
	scheduler.ScheduleTask(&ExampleTask{
		logger: logger,
	})
}

type ExampleTask struct {
	logger func(ctx scheduler.Context) *zap.Logger
}

func (e ExampleTask) TaskID() scheduler.ID {
	return `example-task`
}

func (e ExampleTask) Run(ctx scheduler.Context) error {
	var ll = e.logger(ctx)
	ll.Info("hello world", zap.String("next_run", ctx.NextRunDate().String()))

	time.Sleep(10 * time.Second)
	return nil
}

func (e ExampleTask) Schedule() scheduler.Cron {
	return `0 0 * * *`
}

func (e ExampleTask) ScheduleOptions() scheduler.ScheduleOptions {
	return scheduler.NewScheduleOptions(
		scheduler.Time(2021, 9, 30, 0, 0, 0, 0),
		scheduler.Time(9999, 1, 1, 0, 0, 0, 0),
		true,
		false,
		false,
	)
}

func (e ExampleTask) SubTasks() (isParallel bool, tasks []scheduler.Task) {
	return
}

func logger(ctx scheduler.Context) *zap.Logger {
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(ctx.LogWriteSyncer()), zap.DebugLevel)
	return zap.New(core)
}
