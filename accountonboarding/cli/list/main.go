package main

import (
	"context"
	"log"
	"os"

	"go.temporal.io/sdk/client"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/workflowservice/v1"

	u "webapp/utils"
)

func main() {

	namespace := os.Getenv("TEMPORAL_NAMESPACE")

	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("Failed to load Temporal Cloud environment: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// Query using SearchAttribute
	query := "CustomStringField='ACTIVE-ACCAPP' and CloseTime is null"

	//log.Printf("Listing Workflows with query=(%s):", query)
	log.Printf("Listing ACTIVE-ACCAPP Workflows:")

	//var executions []*workflowpb.WorkflowExecutionInfo
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
			log.Fatal("ListWorkflows returned an error,", err)
		}

		for i := range resp.Executions {
			exec = resp.Executions[i].Execution
			log.Printf("  Execution: WorkflowId: %v, RunId: %v\n", exec.WorkflowId, exec.RunId)
		}

		nextPageToken = resp.NextPageToken
	}

}
