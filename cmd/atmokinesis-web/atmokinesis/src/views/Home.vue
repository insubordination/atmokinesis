<template>
  <div class="home">
    <div class="card">
      <div class="card-body card-body-blue">
        <h2 class="card-title">Welcome to atmokenisis,</h2>
        <h3 class="card-text">the scheduler that never uses <span class="font-italic">DAG</span> and <span
            class="font-italic">Runs</span></h3>
        <h3 class="card-text">in the same sentence. (🐑💩)</h3>
      </div>
    </div>
    <div class="card">
      <div class="card-body card-body-content">
        <h4>Intuitive dashboard</h4>
        <div class="row">
          <div class="col-1"></div>
          <div class="col-11">
            <h5 style="color: #464646">Displays metrics in realtime so you always know the current schedule.</h5>
          </div>
        </div>
        <img src="../assets/img/dashboard-tasks.png" class="card-img" alt="...">
      </div>
    </div>
    <div class="card">
      <div class="card-body card-body-content">
        <h4>Individual task controls and details.</h4>
        <img src="../assets/img/dashboard-task_details.png" class="card-img" alt="...">
      </div>
    </div>
    <div class="card">
      <div class="card-body card-body-content">
        <h4>Easy to read task timeline with the latest status and logs.</h4>
        <img src="../assets/img/dashboard-timeline.png" class="card-img" alt="...">
      </div>
    </div>
    <div class="card">
      <div class="card-body card-body-content">
        <h4>Easy to write and understand Golang tasks.</h4>
        <div class="row">
          <div class="col-1"></div>
          <div class="col-11">
            <h5 style="color: #464646">No confusing documentation and not overly "customizable" it never runs the first
              time.</h5>
          </div>
        </div>
        <div class="row">
          <div class="col-1"></div>
          <div class="col-11">
            <h5 style="color: #464646">Just implement the Task interface and you've got it. We keep it simple because
              simple runs!</h5>
          </div>
        </div>
        <pre style="margin-right: 5rem; margin-left: 5rem; margin-top: 2rem"><code class="language-go">package tasks

import (
	"github.com/insubordination/atmokinesis/cmd/atmokinesis/scheduler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// Schedule your task.
func init() {
	scheduler.ScheduleTask(&ExampleTask{
		logger: logger,
	})
}

// Define your tasks struct with whatever you may need.
type ExampleTask struct {
	logger func(ctx scheduler.Context) *zap.Logger
}

// Give your task a unique ID.
func (e ExampleTask) TaskID() scheduler.ID {
	return `example-task`
}

// Run, The actual logic of the task.
func (e ExampleTask) Run(ctx scheduler.Context) error {
	var ll = e.logger(ctx)
	ll.Info("hello world", zap.String("next_run", ctx.NextRunDate().String()))
	return nil
}

// Define your schedule using a cron expression.
func (e ExampleTask) Schedule() scheduler.Cron {
	return `0 0 * * *`
}

// Set your schedule options, (ex. Start Date, End Date, Whether to stop on failure,
// Allow tasks to overlap their schedule)
func (e ExampleTask) ScheduleOptions() scheduler.ScheduleOptions {
	return scheduler.NewScheduleOptions(
		scheduler.Time(2021, 9, 30, 0, 0, 0, 0),
		scheduler.Time(9999, 1, 1, 0, 0, 0, 0),
		true,
		false,
		false,
	)
}

// Define dependent tasks and whether they can run in parallel.
func (e ExampleTask) SubTasks() (isParallel bool, tasks []scheduler.Task) {
	return
}

func logger(ctx scheduler.Context) *zap.Logger {
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(ctx.LogWriteSyncer()), zap.DebugLevel)
	return zap.New(core)
}
</code></pre>
      </div>
    </div>
  </div>
</template>

<script>
import * as Prism from 'prismjs'
import 'prismjs/components/prism-go'
import 'prismjs/themes/prism-okaidia.css'

export default {
  name: 'Home',
  mounted: () => {
    Prism.highlightAll();
  }
}
</script>

<style scoped>
.card {
  border-width: 0;
}

.card-body-blue {
  padding-top: 5rem;
  padding-bottom: 6.5rem;
  background-color: rgba(0, 223, 253, 0.83)
}

.card-body-content {
  padding-top: 3rem;
  background-color: white;
}

.card-text {
  font-size: 2.5rem;
  font-weight: 100
}
</style>
