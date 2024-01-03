package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"

	u "webapp/utils"
  ao "webapp/accountonboarding"
)

func main() {

	wkflId, err := parseCLIArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("Parameter --workflow-id <workflow id> is required")
	}

	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("Failed to load Temporal Cloud environment: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	log.Println("Query application.status of workflow: ", *wkflId)
  resp, err := c.QueryWorkflow(context.Background(), *wkflId, "", "application.status")
  if err != nil {
    log.Fatalln("Unable to query workflow", err)
  }
  var appstatus ao.ApplicationStatus
  if err := resp.Get(&appstatus); err != nil {
    log.Fatalln("Unable to decode appstatus query result", err)
  }

  log.Println("Query application.details of workflow: ", *wkflId)
  resp, err = c.QueryWorkflow(context.Background(), *wkflId, "", "application.details")
  if err != nil {
    log.Fatalln("Unable to query workflow", err)
  }
  var appdetails ao.ApplicationForm
  if err := resp.Get(&appdetails); err != nil {
    log.Fatalln("Unable to decode appdetails query result", err)
  }

  log.Println("WorkflowId:", *wkflId)
  log.Println("  Application Status:", appstatus.Status)
  log.Println("  Application Status:", appstatus)
  log.Println("  Application Details:", appdetails)
}

func parseCLIArgs(args []string) (*string, error) {

	set := flag.NewFlagSet("signal-workflow", flag.ExitOnError)
	wkflId := set.String("workflow-id", "", "Workflow Id to access")

	if err := set.Parse(args); err != nil {
		return nil, fmt.Errorf("failed parsing args: %w", err)

	} else if *wkflId == "" {
		return nil, fmt.Errorf("--workflow-id argument is required")
	}
	return wkflId, nil
}
