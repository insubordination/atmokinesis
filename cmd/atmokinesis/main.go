package main

import (
	"context"
	"flag"
	"github.com/insubordination/atmokinesis/cmd/atmokinesis-web"
	"github.com/insubordination/atmokinesis/cmd/atmokinesis/scheduler"
	_ "github.com/insubordination/atmokinesis/cmd/atmokinesis/scheduler"
	_ "github.com/insubordination/atmokinesis/tasks"
	"log"
	"os"
	"os/signal"
)

const (
	servePort         = 8081
	defaultDBFilename = "./atmo_db"
	defaultMongoDbUri = "mongodb://root:example@127.0.0.1:8091/"
)

func main() {
	var notify = waitForSignal()
	log.SetOutput(os.Stderr)

	flag.Int("port", servePort, "Port to serve web API.")
	flag.String("db-location", defaultDBFilename, "Atmokinesis DB File")
	log.Println(logo)
	log.Println(`The scheduler that doesn't use "DAG" and "Runs" in the same sentence.`)
	log.Println("---------------------------------------------------------------------------")
	log.Println("starting Server at :", servePort)

	log.Println("initializing store...")
	store, err := scheduler.NewMongoStore(context.TODO(), defaultMongoDbUri)
	if err != nil {
		log.Printf("failed to initialize store, {error: %v}", err)
		os.Exit(1)
	}

	if err = scheduler.InitScheduler(store); err != nil {
		log.Printf("failed to initialize scheduler, {error: %v}", err)
		os.Exit(1)
	}

	err = atmokinesis_web.StartServer()
	log.Printf("serving web UI at -> http://127.0.0.1:%d", servePort)

	log.Println("signal received (", <-notify, "), shutting down...")
	if err != nil {
		log.Println("error: ", err)
	}
	if err = scheduler.StopScheduler(store); err != nil {
		log.Println("error: ", err)
	}
}

func waitForSignal() chan os.Signal {
	notify := make(chan os.Signal, 1)
	signal.Notify(notify, os.Kill, os.Interrupt)
	return notify
}

var logo = `
  ___ ________  ________ _   _______ _   _  _____ _____ _____ _____ 
 / _ \_   _|  \/  |  _  | | / /_   _| \ | ||  ___/  ___|_   _/  ___|
/ /_\ \| | | .  . | | | | |/ /  | | |  \| || |__ \ '--.  | | \ '--.
|  _  || | | |\/| | | | |    \  | | | . ' ||  __| '--. \ | |  '--. \
| | | || | | |  | \ \_/ / |\  \_| |_| |\  || |___/\__/ /_| |_/\__/ /
\_| |_/\_/ \_|  |_/\___/\_| \_/\___/\_| \_/\____/\____/ \___/\____/

`
