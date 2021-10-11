package scheduler

import (
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"sync"
	"time"
)

type EntryStatus string

const (
	Stopped    EntryStatus = "Stopped"
	Running    EntryStatus = "Running"
	PendingRun EntryStatus = "Pending Run"
	Failing    EntryStatus = "Failing"
	Success    EntryStatus = "Success"
)

// Entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	// The schedule on which this job should be run.
	Schedule Schedule `json:"-"`

	// The next time the job will run. This is the zero time if Atmo has not been
	// started or this entry's schedule is unsatisfiable
	Next time.Time `json:"next"`

	// The last time this job was run. This is the zero time if the job has never
	// been run.
	Prev time.Time `json:"prev"`

	// The Task to run.
	Task Task `json:"task"`

	Notify chan bool `json:"-"`

	Status EntryStatus `json:"-"`

	History       []*TaskHistory      `json:"history"` // time | status
	Errors        map[time.Time]error `json:"errors"`
	*sync.RWMutex `json:"-"`
}

func (e Entry) MarshalToBSON() (bson.M, error) {
	var history = bson.A{}
	var errors = bson.A{}
	for _, h := range e.History {
		history = append(history, bson.M{h.ExecutionTime.String(): bson.D{
			{Key: "logs", Value: h.Logs},
			{Key: "status", Value: h.Status},
		}})
	}
	for key, er := range e.Errors {
		errors = append(errors, bson.M{key.String(): bson.D{
			{"error", er},
		}})
	}
	return bson.M{
		"history": bson.M{"$each": history},
		"errors":  bson.M{"$each": errors},
	}, nil
}

func UnmarshalBSON(data bson.M, e *Entry) error {
	e.Lock()
	defer e.Unlock()
	for key, val := range data {
		switch key {
		case "history":
			for _, a := range val.(bson.A) {
				for k, v := range a.(bson.M) {
					k = strings.TrimSpace(strings.Split(k, "m=+")[0])
					ti, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", k)
					if err != nil {
						return err
					}
					status := v.(bson.M)["status"].(string)
					logs := v.(bson.M)["logs"].(string)
					e.History = append(e.History, &TaskHistory{
						ExecutionTime: ti,
						Status:        EntryStatus(status),
						Logs:          logs,
					})
				}
			}
		case "errors":
			for _, a := range val.(bson.A) {
				for k, v := range a.(bson.M) {
					ti, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", k)
					if err != nil {
						return err
					}
					e.Errors[ti] = v.(bson.D).Map()["error"].(error)
				}
			}
		}
	}
	return nil
}

func (e *Entry) ChangeStatus(s EntryStatus) {
	e.Lock()
	defer e.Unlock()
	e.Status = s
}
