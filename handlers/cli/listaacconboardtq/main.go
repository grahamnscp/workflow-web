package main

import (
  "fmt"
	"context"
	"encoding/json"
	"log"
	"os"

	"go.temporal.io/sdk/client"

	apicompb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"

	u "webapp/utils"
  ao "webapp/accountonboarding"
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

	ctx := context.Background()

	// Query using SearchAttribute
	query := fmt.Sprintf("ExecutionStatus = 'Running' and TaskQueue = '%s'", ao.AccOnboardingTaskQueueName)

	//type: executions []*workflowpb.WorkflowExecutionInfo
	var exec *apicompb.WorkflowExecution
	var nextPageToken []byte

	for hasMore := true; hasMore; hasMore = len(nextPageToken) > 0 {
		resp, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     namespace,
			PageSize:      10,
			NextPageToken: nextPageToken,
			Query:         query,
		})
		if err != nil {
			log.Fatal("ListWorkflows returned an error,", err)
		}

		if len(resp.Executions) > 0 {

			log.Printf("Listing Running Workflow Executions (%d):", len(resp.Executions))

			for i := range resp.Executions {

				exec = resp.Executions[i].Execution

				// get SearchAttribute CustomStringField value
				pc := converter.NewJSONPayloadConverter()
				sas, _ := pc.ToPayload(resp.Executions[i].GetSearchAttributes())
				sajsondata := string(sas.Data)
				var sadata map[string]interface{}
				_ = json.Unmarshal([]byte(sajsondata), &sadata)
				sattr := u.DecodeB64(sadata["indexed_fields"].(map[string]interface{})["CustomStringField"].(map[string]interface{})["data"].(string))

				// found an active workflow execution:
				log.Printf("  WorkflowId: %v, RunId: %v, SearchAttribute: %s\n", exec.WorkflowId, exec.RunId, sattr)

				// Describe the workflow
				describe, err := c.DescribeWorkflowExecution(context.Background(), exec.WorkflowId, exec.RunId)
				if err != nil {
					log.Fatalln("fail to descibe workflow", err)
				}

				// Check for pending activities
				pendingActivity := describe.GetPendingActivities()
				if pendingActivity != nil {
					for _, pendingActivity := range describe.GetPendingActivities() {
						log.Printf("    Pending Activity: '%s' has '%d' Retries, Error: '%v'", pendingActivity.GetActivityType().Name, pendingActivity.GetAttempt(), pendingActivity.GetLastFailure().Message)
					}
				}
			}
		} else {
			log.Println("No running executions found")
		}
		nextPageToken = resp.NextPageToken
	}

}
