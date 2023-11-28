package main

import (
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	so "webapp/standingorder"
	u "webapp/utils"
)

func main() {
	log.Printf("%sGo worker starting..%s", u.ColorGreen, u.ColorReset)

	// Load the Temporal Cloud from env
	clientOptions, err := u.LoadClientOptions(u.SDKMetrics, "8078")
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
	workername := "StandingOrderWorker." + hostname + ":" + fmt.Sprintf("%d", os.Getpid())

	log.Println("Go worker (" + workername + ") initialising..")

	w := worker.New(c, so.StandingOrdersTaskQueueName, worker.Options{Identity: workername})

	log.Println("Go worker registering for Workflow moneytransfer.StandingOrderWorkflow:")
	w.RegisterWorkflow(so.StandingOrderWorkflow)

	// Start listening to the Task Queue.
	log.Printf("%sGo worker listening on %s task queue..%s", u.ColorGreen, so.StandingOrdersTaskQueueName, u.ColorReset)
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start StandingOrderWorkflow Worker", err)
	}

	log.Printf("%sGo worker stopped.%s", u.ColorGreen, u.ColorReset)
}
