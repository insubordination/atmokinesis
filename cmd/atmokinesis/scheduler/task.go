package scheduler

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"runtime"
	"time"
)

type ID string
type Cron string

func (id ID) ToString() string {
	return string(id)
}

func (c Cron) ToString() string {
	return string(c)
}

type Task interface {
	TaskID() ID
	Run(ctx Context) error
	Schedule() Cron
	ScheduleOptions() ScheduleOptions
	SubTasks() (isParallel bool, tasks []Task)
}

type Context interface {
	ExecutionDate() time.Time
	StartDate() time.Time
	NextRunDate() time.Time
	PreviousRunDate() time.Time
	LogWriteSyncer() WriteSyncer
	StreamToSubTasks(out interface{})
	NotifySubTasks()
}

type WriteSyncer interface {
	io.Writer
	Sync() error
}

func DockerRunner(ctx context.Context, containerName string, containerConfig *container.Config,
	hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig, logsConfig types.ContainerLogsOptions) (logs io.ReadCloser, containerErr chan error, err error) {
	var resp = container.ContainerCreateCreatedBody{}
	var c *client.Client
	containerErr = make(chan error, 1)

	c, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return
	}

	var platform = &v1.Platform{
		Architecture: runtime.GOARCH,
		OS:           runtime.GOOS,
		OSVersion:    "",
		OSFeatures:   nil,
		Variant:      "",
	}

	resp, err = c.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, platform, containerName)
	if err != nil {
		return
	}

	if logs, err = c.ContainerLogs(ctx, resp.ID, logsConfig); err != nil {
		return
	}

	if err = c.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return
	}

	go func() {
		for {
			var jsbody types.ContainerJSON
			if jsbody, err = c.ContainerInspect(ctx, resp.ID); err != nil {
				return
			}
			if len(jsbody.State.FinishedAt) > 0 {
				if jsbody.State.ExitCode != 0 {
					containerErr <- fmt.Errorf("pod exited with non 0 code: exit_code:%d, error:%s", jsbody.State.ExitCode, jsbody.State.Error)
					return
				} else {
					containerErr <- nil
					return
				}
			}
		}
	}()
	return
}
