package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"go.temporal.io/sdk/client"

	mt "webapp/moneytransfer"
	u "webapp/utils"
)

/* Main */
func main() {

	log.Println("workflow start program..")

	// Load the Temporal Cloud from env
	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("Failed to load Temporal Cloud environment: %v", err)
	}
	log.Println("connecting to temporal server..")
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create Temporal client.", err)
	}
	defer c.Close()

	// Temporal Client Start Workflow Options
	workflowID := fmt.Sprintf("go-moneytxfr-wkfl-%d", rand.Intn(99999))

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: mt.MoneyTransferTaskQueueName,
	}

	// Sample workflow data
	pmnt := &mt.PaymentDetails{
		SourceAccount: "harry",
		TargetAccount: "sally",
		ReferenceID:   "from Go Starter",
		Amount:        100,
	}
	var delay int = 5 // delay between withdraw and deposit for demo purposes (seconds)

	// ExecuteWorkflow mt.Transfer
	log.Println("Starting moneytransfer workflow on", mt.MoneyTransferTaskQueueName, "task queue")

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, mt.TransferWorkflow, *pmnt, delay)

	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Printf("%sWorkflow started:%s (WorkflowID: %s, RunID: %s)", u.ColorYellow, u.ColorReset, we.GetID(), we.GetRunID())

	// Check workflow status
	var result string

	err = we.Get(context.Background(), &result)

	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalln("Unable to format result in JSON format", err)
	}
	log.Printf("%sWorkflow result:%s %s", u.ColorYellow, u.ColorReset, string(data))

	// done
	log.Print("Start workflow client done.")
}
