package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
  "flag"
  "os"

	"go.temporal.io/sdk/client"

	ao "webapp/accountonboarding"
	u "webapp/utils"
)

func main() {

	log.Println("workflow start program..")

  email, err := parseCLIArgs(os.Args[1:])
  if err != nil {
    log.Fatalf("Parameter --email <applicant email address> is required")
  }

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
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("go-accapp-wkfl-%d", rand.Intn(99999)),
		TaskQueue: ao.AccOnboardingTaskQueueName,
	}

	// Sample workflow data
	appl := &ao.ApplicationForm{
		FirstName: "test",
		Surname:   "user",
		Email:     *email,
	}

	log.Println("Starting AccountApplication workflow with email:", email, "..")
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, ao.AccountApplicationWorkflow, *appl)

	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Printf("%sWorkflow started:%s (WorkflowID: %s, RunID: %s)", u.ColorYellow, u.ColorReset, we.GetID(), we.GetRunID())
}

func parseCLIArgs(args []string) (*string, error) {

  set := flag.NewFlagSet("start-accapp-workflow", flag.ExitOnError)
  email := set.String("email", "", "email address for application")

  if err := set.Parse(args); err != nil {
    return nil, fmt.Errorf("failed parsing args: %w", err)

  } else if *email == "" {
    return nil, fmt.Errorf("--email argument is required")
  }
  return email, nil
}
