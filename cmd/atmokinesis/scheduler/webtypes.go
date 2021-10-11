package scheduler

import "time"

type DisplayTask struct {
	ID       string         `json:"id,omitempty"`
	Status   string         `json:"status,omitempty"`
	Schedule Cron           `json:"schedule,omitempty"`
	NextRun  time.Time      `json:"next_run,omitempty"`
	LastRun  time.Time      `json:"last_run,omitempty"`
	History  []*TaskHistory `json:"history,omitempty"`
}
