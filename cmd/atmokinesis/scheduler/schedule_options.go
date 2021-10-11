package scheduler

import (
	"time"
)

// ScheduleOptions
type ScheduleOptions interface {
	StartDate() time.Time
	EndDate() time.Time
	StopOnFailure() bool // Should we stop on a Failure
	AllowOverlap() bool
	Rescue() bool // Catch up on previous failed tasks
}

func NoEndDate() time.Time {
	return time.Date(2999, 1, 1, 0, 0, 0, 0, time.Local)
}

// NewScheduleOptions
func NewScheduleOptions(startDate, endDate time.Time, stopOnFailure, allowOverlap, rescue bool) *DefaultScheduleOptions {
	return &DefaultScheduleOptions{startDate: startDate, endDate: endDate, stopOnFailure: stopOnFailure, allowOverlap: allowOverlap, rescue: rescue}
}

// DefaultScheduleOptions
type DefaultScheduleOptions struct {
	startDate     time.Time
	endDate       time.Time
	stopOnFailure bool
	allowOverlap  bool
	rescue        bool
}

func (d DefaultScheduleOptions) StartDate() time.Time {
	return d.startDate
}

func (d DefaultScheduleOptions) EndDate() time.Time {
	return d.endDate
}

func (d DefaultScheduleOptions) StopOnFailure() bool {
	return d.stopOnFailure
}

func (d DefaultScheduleOptions) AllowOverlap() bool {
	return d.allowOverlap
}

func (d DefaultScheduleOptions) Rescue() bool {
	return d.rescue
}

func NewStartImmediately(stopOnFailure, allowOverlap, rescue bool) ScheduleOptions {
	return &StartImmediately{
		stopOnFailure: stopOnFailure,
		rescue:        rescue,
		allowOverlap:  allowOverlap,
	}
}

// NoScheduleOptions
type StartImmediately struct {
	stopOnFailure bool
	rescue        bool
	allowOverlap  bool
}

func (s StartImmediately) StartDate() time.Time {
	return time.Now()
}

func (s StartImmediately) EndDate() time.Time {
	return NoEndDate()
}

func (s StartImmediately) StopOnFailure() bool {
	return s.stopOnFailure
}

func (s StartImmediately) AllowOverlap() bool {
	return s.allowOverlap
}

func (s StartImmediately) Rescue() bool {
	return s.Rescue()
}
