package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"go.temporal.io/sdk/client"

	apicompb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"

	a "webapp/accountonboarding"
	m "webapp/moneytransfer"
	s "webapp/standingorder"

	u "webapp/utils"
)

// Local structs
type PendingActivity struct {
	Name    string
	Retries string
	Msg     string
}
type RunningWF struct {
	Id              string
	RunId           string
	SearchAttribute string
	HasPending      bool
	PendingActivity PendingActivity
}

/* ListRunningWFs */
func ListRunningWFs(w http.ResponseWriter, r *http.Request) {

	//log.Println("ListRunningWFs: called, method:", r.Method)

	// Request Parameter?
	var querystr string = "1"
	params := r.URL.Query()
	for k, v := range params {
		if k == "query" {
			querystr = strings.Join(v, "")
		}
		//log.Println("ShowTransfer: url params:", k, " => ", v)
	}
	querynum, _ := strconv.Atoi(querystr)

	// Default query is all Running
	query := "ExecutionStatus = 'Running'"

	switch querynum {
	case 1: // Default: All Running Workflows
		query = "ExecutionStatus = 'Running'"

	case 2: // Money Transfers TQ
		query = fmt.Sprintf("ExecutionStatus = 'Running' and TaskQueue = '%s'", m.MoneyTransferTaskQueueName)

	case 3: // Standing Orders TQ
		query = fmt.Sprintf("ExecutionStatus = 'Running' and TaskQueue = '%s'", s.StandingOrdersTaskQueueName)

	case 4: // Account Onboarding TQ
		query = fmt.Sprintf("ExecutionStatus = 'Running' and TaskQueue = '%s'", a.AccOnboardingTaskQueueName)
	}
	log.Printf("ListRunningWFs: query: \"%s\"", query)

	// Variables
	pa := PendingActivity{}
	wf := RunningWF{}
	wfs := []RunningWF{}

	namespace := os.Getenv("TEMPORAL_NAMESPACE")

	// Temporal Service connection
	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Printf("ListRunningWFs: Failed to load Temporal Cloud environment: %v", err)
		return
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Println("ListRunningWFs: Unable to create client", err)
		return
	}
	defer c.Close()

	ctx := context.Background()

	// Query Temporal Service
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
			log.Println("ListRunningWFs: ListWorkflows returned an error,", err)
			return
		}

		if len(resp.Executions) > 0 {

			log.Printf("ListRunningWFs: Listing Running Workflow Executions (%d):", len(resp.Executions))

			for i := range resp.Executions {

        // this Workflow Execution
				exec = resp.Executions[i].Execution

				// get SearchAttribute CustomStringField value
				pc := converter.NewJSONPayloadConverter()
				sas, _ := pc.ToPayload(resp.Executions[i].GetSearchAttributes())
				sajsondata := string(sas.Data)
				var sadata map[string]interface{}
				_ = json.Unmarshal([]byte(sajsondata), &sadata)
				sattr := u.DecodeB64(sadata["indexed_fields"].(map[string]interface{})["CustomStringField"].(map[string]interface{})["data"].(string))
				sattr = u.ClearString(sattr)

        // populate the ui form struct
				wf = RunningWF{
					Id:              exec.WorkflowId,
					RunId:           exec.RunId,
					SearchAttribute: sattr,
					HasPending:      false,
					PendingActivity: pa,
				}

				log.Printf("  WorkflowId: %v, RunId: %v, SearchAttribute: '%s'\n", exec.WorkflowId, exec.RunId, sattr)

				// Describe the workflow
				describe, err := c.DescribeWorkflowExecution(context.Background(), exec.WorkflowId, exec.RunId)
				if err != nil {
					log.Println("fail to descibe workflow", err)
					return
				}

				// Check for pending activities
				pendingActivity := describe.GetPendingActivities()
				if pendingActivity != nil {
					for _, pendingActivity := range describe.GetPendingActivities() {
						log.Printf("%s    Pending Activity: '%s' has '%d' Retries, Error: '%v'%s", u.ColorRed, 
              pendingActivity.GetActivityType().Name, pendingActivity.GetAttempt(), pendingActivity.GetLastFailure().Message, u.ColorReset)
						pa := PendingActivity{
							Name:    pendingActivity.GetActivityType().Name,
							Retries: fmt.Sprintf("%d", pendingActivity.GetAttempt()),
							Msg:     fmt.Sprintf("%v", pendingActivity.GetLastFailure().Message),
						}

						// Add pending activity list to this wf
						wf.HasPending = true
						wf.PendingActivity = pa
					}
				}

				// Add the worflow to the list
				wfs = append(wfs, wf)
			}

		} else {
			log.Println("ListRunningWFs: No running workflow executions found")
		}
		nextPageToken = resp.NextPageToken
	}

	u.Render(w, "templates/ListRunningWFs.html", wfs)
}
