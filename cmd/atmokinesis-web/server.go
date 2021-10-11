package atmokinesis_web

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/insubordination/atmokinesis/cmd/atmokinesis/scheduler"
	"net/http"
	"reflect"
	"time"
)

func StartServer() (err error) {
	var upgrader = &websocket.Upgrader{}
	mux := http.NewServeMux()
	Routes(upgrader, mux)
	go func() {
		serveErr := http.ListenAndServe(":8082", mux)
		if serveErr != nil {
			errors.As(serveErr, &err)
		}
	}()
	return
}

func Routes(upgrader *websocket.Upgrader, mux *http.ServeMux) {
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

	})

	mux.HandleFunc("/taskstatus", func(writer http.ResponseWriter, request *http.Request) {
		tic := time.NewTicker(3000 * time.Millisecond)
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		c, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		go func() {
			defer c.Close()
			for {
				mt, message, err := c.ReadMessage()
				if err != nil {
					break
				}
				switch {
				case string(message) == "get" && mt == 1:
					go func() {
						var currentTaskList []scheduler.DisplayTask
						if err = c.WriteJSON(scheduler.TaskList()); err != nil {
							return
						}
						for {
							select {
							case <-tic.C:
								var taskList = scheduler.TaskList()
								if !reflect.DeepEqual(taskList, currentTaskList) {
									currentTaskList = taskList
									if err = c.WriteJSON(scheduler.TaskList()); err != nil {
										return
									}
								}
							}
						}
					}()
				case string(message) == "stop" && mt == 1:
					tic.Stop()
					return
				default:

				}
			}
		}()
	})
}
