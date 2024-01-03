package main

import (
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	ao "webapp/accountonboarding"
	u "webapp/utils"
)

func main() {
	log.Printf("%sGo worker starting..%s", u.ColorGreen, u.ColorReset)

	// Load the Temporal Cloud from env
	clientOptions, err := u.LoadClientOptions(u.SDKMetrics, "8079")
	if err != nil {
		log.Fatalf("Failed to load Temporal Cloud environment: %v", err)
	}

	log.Println("Go worker connecting to server..")

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create Temporal client.", err)
	}
	defer c.Close()

  // Set custom worker name
	hostname, _ := os.Hostname()
	workername := "AccOnboardWorker." + hostname + ":" + fmt.Sprintf("%d", os.Getpid())

	log.Println("Go worker (" + workername + ") initialising..")

	w := worker.New(c, ao.AccOnboardingTaskQueueName, worker.Options{Identity: workername})

	// Register Workflows
	log.Println("Go worker registering for Workflow accountonboarding.AccountApplicationWorkflow..")
	w.RegisterWorkflow(ao.AccountApplicationWorkflow)

	log.Println("Go worker registering for Workflow accountonboarding.CreditCheckWorkflow..")
	w.RegisterWorkflow(ao.CreditCheckWorkflow)

	log.Println("Go worker registering for Workflow accountonboarding.FraudCheckWorkflow..")
	w.RegisterWorkflow(ao.FraudCheckWorkflow)

	// Register Activities
	log.Println("Go worker registering for Activity accountonboarding.CheckBlacklist..")
	w.RegisterActivity(ao.CheckBlacklist)

	log.Println("Go worker registering for Activity accountonboarding.CheckFraudRisk..")
	w.RegisterActivity(ao.CheckFraudRisk)

	log.Println("Go worker registering for Activity accountonboarding.GetFraudApprover..")
	w.RegisterActivity(ao.GetFraudApprover)

	log.Println("Go worker registering for Activity accountonboarding.SendNotificationEmail..")
	w.RegisterActivity(ao.SendNotificationEmail)

	log.Println("Go worker registering for Activity accountonboarding.SendApprovalEmail..")
	w.RegisterActivity(ao.SendApprovalEmail)

	log.Println("Go worker registering for Activity accountonboarding.ProvisionAccount..")
	w.RegisterActivity(ao.ProvisionAccount)

	log.Println("Go worker registering for Activity accountonboarding.GenerateWorkflowToken..")
	w.RegisterActivity(ao.GenerateWorkflowToken)

	// Start listening on the Task Queue.
	log.Printf("%sGo worker listening on %s task queue..%s", u.ColorGreen, ao.AccOnboardingTaskQueueName, u.ColorReset)
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start AccountOnboardingWorkflow Worker", err)
	}

	log.Printf("%sGo worker stopped.%s", u.ColorGreen, u.ColorReset)
}
