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

  log.Println("Query fraud.applicationdetails from workflow: ", *wkflId)
  resp, err := c.QueryWorkflow(context.Background(), *wkflId, "", "fraud.applicationdetails")
  if err != nil {
    log.Fatalln("Unable to query workflow", err)
  }
  var appdetails ao.ApplicationForm
  if err := resp.Get(&appdetails); err != nil {
    log.Fatalln("Unable to decode appdetails query result", err)
  }

	log.Println("Query fraud.approvalstatus from workflow : ", *wkflId)
  resp, err = c.QueryWorkflow(context.Background(), *wkflId, "", "fraud.approvalstatus")
  if err != nil {
    log.Fatalln("Unable to query workflow", err)
  }
  var fraudapprovalstatus ao.ApprovalStatus
  if err := resp.Get(&fraudapprovalstatus); err != nil {
    log.Fatalln("Unable to decode fraudapprovalstatus query result", err)
  }

	log.Println("Query fraud.parentworkflowid from workflow : ", *wkflId)
  resp, err = c.QueryWorkflow(context.Background(), *wkflId, "", "fraud.parentworkflowid")
  if err != nil {
    log.Fatalln("Unable to query workflow", err)
  }
  var parentwkflId string
  if err := resp.Get(&parentwkflId); err != nil {
    log.Fatalln("Unable to decode parentwkflId query result", err)
  }

  log.Println("WorkflowId:", *wkflId)
  log.Println("  Application Details:", appdetails)
  log.Println("  Approval Status:", fraudapprovalstatus)
  log.Println("  Parent Workflow ID: ", parentwkflId)
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
