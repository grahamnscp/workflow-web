package main

import (
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	mt "webapp/moneytransfer"
	u "webapp/utils"
)

func main() {
	log.Printf("%sGo worker starting..%s", u.ColorGreen, u.ColorReset)

	// Load the Temporal Cloud from env
	clientOptions, err := u.LoadClientOptions(u.SDKMetrics, "8077")
	if err != nil {
		log.Fatalf("Failed to load Temporal Cloud environment: %v", err)
	}

	log.Println("Go worker connecting to server..")

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create Temporal client.", err)
	}
	defer c.Close()

	taskqueuename := mt.MoneyTransferTaskQueueName
	hostname, _ := os.Hostname()
	workername := "MoneyTransferWorker." + hostname + ":" + fmt.Sprintf("%d", os.Getpid())

	log.Println("Go worker (" + workername + ") initialising..")

	w := worker.New(c, taskqueuename, worker.Options{Identity: workername})

	// This worker hosts both Workflow and Activity functions.
	log.Println("Go worker registering for Workflow moneytransfer.Transfer..")
	w.RegisterWorkflow(mt.TransferWorkflow)

	log.Println("Go worker registering for Activity moneytransfer.Withdraw..")
	w.RegisterActivity(mt.Withdraw)

	log.Println("Go worker registering for Activity moneytransfer.Deposit..")
	w.RegisterActivity(mt.Deposit)

	log.Println("Go worker registering for Activity moneytransfer.Refund..")
	w.RegisterActivity(mt.Refund)

	// Start listening to the Task Queue.
	log.Printf("%sGo worker listening on %s task queue..%s", u.ColorGreen, taskqueuename, u.ColorReset)
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start MoneyTransfer Worker", err)
	}

	log.Printf("%sGo worker stopped.%s", u.ColorGreen, u.ColorReset)
}
