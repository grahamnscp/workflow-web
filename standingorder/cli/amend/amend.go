package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	u "webapp/utils"
)

func main() {

	wkflId, amount, period, err := amendparseCLIArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("Parameter --workflowi-id <workflow id> is required")
	}
	log.Println("  Amend Parameters: --amount", *amount, "--period", *period)

	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("Failed to load Temporal Cloud environment: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// check workflow is active first
	foundwf, err := checkWorkflowActive(c, *wkflId)
	if err != nil || !foundwf {
		log.Fatalln("checkWorkflowActive Failed,", err, "( found:", foundwf, ")")
	}

	// Signal the Workflow Executions to amend values:
	if *amount != "" {
		log.Println("Sending amend signal to workflow:", *wkflId, "for sorderamount:", *amount)
		amt, _ := strconv.Atoi(*amount)
		err = c.SignalWorkflow(context.Background(), *wkflId, "", "sorderamount", amt)
		if err != nil {
			log.Fatalln("Unable to signal workflow", err)
		}
	}
	if *period != "" {
		log.Println("Sending amend signal to workflow:", *wkflId, "for sorderschedule:", *period)
		dur, _ := strconv.Atoi(*period)
		err = c.SignalWorkflow(context.Background(), *wkflId, "", "sorderschedule", dur)
		if err != nil {
			log.Fatalln("Unable to signal workflow", err)
		}
	}
}

func amendparseCLIArgs(args []string) (*string, *string, *string, error) {
	set := flag.NewFlagSet("hello-workflow", flag.ExitOnError)

	wkflId := set.String("workflow-id", "", "Workflow Id to access")

	amount := set.String("amount", "", "New money transfer amount")
	period := set.String("period", "", "New money transfer schedule duration")

	if err := set.Parse(args); err != nil {
		return nil, nil, nil, fmt.Errorf("failed parsing args: %w", err)
	} else if *wkflId == "" {
		return nil, nil, nil, fmt.Errorf("--workflow-id argument is required")
	}
	return wkflId, amount, period, nil
}

func checkWorkflowActive(c client.Client, wkflId string) (bool, error) {

	// Query using SearchAttribute
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	query := "CustomStringField='ACTIVE-SORDER' and CloseTime is null"
	var exec *commonpb.WorkflowExecution
	var nextPageToken []byte
	for hasMore := true; hasMore; hasMore = len(nextPageToken) > 0 {
		resp, err := c.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     namespace,
			PageSize:      10,
			NextPageToken: nextPageToken,
			Query:         query,
		})
		if err != nil {
			log.Fatalln("ListWorkflows returned an error,", err)
			return false, errors.New("Failed to list workflows")
		}

		for i := range resp.Executions {
			exec = resp.Executions[i].Execution
			if exec.WorkflowId == wkflId {
				return true, nil
			}
		}
		nextPageToken = resp.NextPageToken
	}
	return false, errors.New("workflow not active")
}
