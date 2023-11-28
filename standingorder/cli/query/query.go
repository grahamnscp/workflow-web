package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"

	u "webapp/utils"
)

func main() {

	wkflId, err := queryparseCLIArgs(os.Args[1:])
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

	log.Println("Querying workflow: ", *wkflId)

	// Query all variables in workflow..
	resp, err := c.QueryWorkflow(context.Background(), *wkflId, "", "payment.origin")
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}

	resp, err = c.QueryWorkflow(context.Background(), *wkflId, "", "payment.destination")
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result2 interface{}
	if err := resp.Get(&result2); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}

	resp, err = c.QueryWorkflow(context.Background(), *wkflId, "", "payment.amount")
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result3 interface{}
	if err := resp.Get(&result3); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}

	resp, err = c.QueryWorkflow(context.Background(), *wkflId, "", "payment.reference")
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result4 interface{}
	if err := resp.Get(&result4); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}

	resp, err = c.QueryWorkflow(context.Background(), *wkflId, "", "schedule.periodduration")
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result5 interface{}
	if err := resp.Get(&result5); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}

	log.Println("WorkflowId:", wkflId)
	log.Println("  Payment.Origin Account Name:", result)
	log.Println("  Payment Destination Account Name:", result2)
	log.Println("  Payment Amount:", result3)
	log.Println("  Payment Reference:", result4)
	log.Println("  Schedule Period Duration:", result5)
}

func queryparseCLIArgs(args []string) (*string, error) {
	set := flag.NewFlagSet("query", flag.ExitOnError)
	wkflId := set.String("workflow-id", "", "Workflow Id to access")
	if err := set.Parse(args); err != nil {
		return nil, fmt.Errorf("failed parsing args: %w", err)
	} else if *wkflId == "" {
		return nil, fmt.Errorf("--workflow-id argument is required")
	}
	return wkflId, nil
}
