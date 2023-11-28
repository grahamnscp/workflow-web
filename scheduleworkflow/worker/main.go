package main

import (
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	sw "webapp/scheduleworkflow"
	u "webapp/utils"
)

func main() {
	log.Printf("%sGo worker starting..%s", u.ColorGreen, u.ColorReset)

	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("Failed to load Temporal Cloud environment: %v", err)
	}
	log.Println("Go worker connecting to server..")

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create Temporal client.", err)
	}
	defer c.Close()

	hostname, _ := os.Hostname()
	workername := "ScheduleWFWorker." + hostname + ":" + fmt.Sprintf("%d", os.Getpid())

	log.Println("Go worker (" + workername + ") initialising..")

	w := worker.New(c, sw.ScheduleWFTaskQueueName, worker.Options{Identity: workername})

	log.Println("Go worker registering for Workflow scheduleworkflow..")
	w.RegisterWorkflow(sw.ScheduleWorkflow)
	log.Println("Go worker registering for Activity ScheduleEmail..")
	w.RegisterActivity(sw.ScheduleEmail)

	log.Printf("%sGo worker listening on %s task queue..%s", u.ColorGreen, sw.ScheduleWFTaskQueueName, u.ColorReset)
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}

	log.Printf("%sGo worker stopped.%s", u.ColorGreen, u.ColorReset)
}
