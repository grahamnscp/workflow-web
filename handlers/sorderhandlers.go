package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	mt "webapp/moneytransfer"
	so "webapp/standingorder"
	u "webapp/utils"
)

/* Flat local struct to pass standing order data to template */
type SOrderData struct {
	WorkflowID             string
	PaymentOrigin          string
	PaymentDestination     string
	PaymentAmount          int
	PaymentReference       string
	SchedulePeriodDuration int
}

/* ListSOrders */
func ListSOrders(w http.ResponseWriter, r *http.Request) {

	log.Println("ListSOrders: called")

	namespace := os.Getenv("TEMPORAL_NAMESPACE")

	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("ListSOrders: Failed to load Temporal Cloud environment: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("ListSOrders: Unable to create client", err)
	}
	defer c.Close()

	// Query using SearchAttribute
	//query := "CustomStringField='ACTIVE-SORDER' and CloseTime is null"
	query := "CustomStringField='ACTIVE-SORDER' and ExecutionStatus = 'Running'"

	log.Printf("ListSOrders: Listing ACTIVE-SORDER Workflows:")

	sorder := SOrderData{}
	sorders := []SOrderData{}

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
			log.Fatal("ListSOrders: ListWorkflows returned an error,", err)
		}

		for i := range resp.Executions {
			exec = resp.Executions[i].Execution
			log.Printf("ListSOrders:   Execution: WorkflowId: %v, RunId: %v\n", exec.WorkflowId, exec.RunId)

			sorder.WorkflowID = exec.WorkflowId

			// Query variables form workflow
			resp, err := c.QueryWorkflow(context.Background(), sorder.WorkflowID, "", "payment.reference")
			if err != nil {
				log.Fatalln("ListSOrders: Unable to query workflow,", err)
			} else {
				var result interface{}
				if err := resp.Get(&result); err != nil {
					log.Fatalln("ListSOrders: Unable to decode query result,", err)
				} else {
					sorder.PaymentReference = fmt.Sprintf("%v", result)
				}
			}
			log.Printf("ListSOrders:   Workflow Query Details: %v\n", sorder)
			sorders = append(sorders, sorder)
		}
		nextPageToken = resp.NextPageToken
	}

	u.Render(w, "templates/ListSOrders.html", sorders)
}

/* NewSOrder */
func NewSOrder(w http.ResponseWriter, r *http.Request) {

	log.Println("NewSOrder: called")

	log.Println("NewSOrder: method:", r.Method) //get request method
	if r.Method == "GET" {
		u.Render(w, "templates/NewSOrder.html", nil)
		return
	}

	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	amt, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("amount")))
	period, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("period")))
	if period == 0 {
		period = 120
	}
	pmnt := &mt.PaymentDetails{
		SourceAccount: r.FormValue("origin"),
		TargetAccount: r.FormValue("destination"),
		ReferenceID:   r.FormValue("reference"),
		Amount:        amt,
	}
	schl := &so.PaymentSchedule{
		PeriodDuration: time.Duration(period) * time.Second,
		Active:         true,
	}
	log.Println("NewSOrder: Processing new Standing Order:", pmnt, schl)

	// have data from form, create new Temporal workflow

	// Temporal Client Start Workflow Options
	wkflowid := fmt.Sprintf("go-txfr-sorder-wkfl-%d", rand.Intn(99999))
	log.Println("NewSOrder: Submitting Temporal Workflow:", wkflowid)

	// Load the Temporal Cloud from env
	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("NewSOrder: Failed to load Temporal Cloud environment: %v", err)
	}
	log.Println("NewSOrder: connecting to temporal server..")
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("NewSOrder: Unable to create Temporal client.", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        wkflowid,
		TaskQueue: so.StandingOrdersTaskQueueName,
	}

	log.Println("NewSOrder: Starting standingorder workflow ..")
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, so.StandingOrderWorkflow, *pmnt, *schl)

	if err != nil {
		log.Fatalln("NewSOrder: Unable to execute workflow", err)
	}
	log.Printf("NewSOrder: %sWorkflow started:%s (WorkflowID: %s, RunID: %s)", u.ColorYellow, u.ColorReset, we.GetID(), we.GetRunID())

	// Render acknowledgement page
	u.Render(w, "templates/NewSOrder.html", struct{ Success bool }{true})
}

/* AmendSOrder */
func AmendSOrder(w http.ResponseWriter, r *http.Request) {

	log.Println("AmendSOrder: called")

	// URL Parameters
	var wkflId string
	params := r.URL.Query()
	for k, v := range params {
		if k == "id" {
			wkflId = strings.Join(v, "")
		}
		log.Println("AmendSOrder: url params:", k, " => ", v)
	}

	log.Println("AmendSOrder: method:", r.Method) //get request method

	// Load the Temporal Cloud from env
	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("AmendSOrder: Failed to load Temporal Cloud environment: %v", err)
	}
	log.Println("AmendSOrder: connecting to temporal server..")
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("AmendSOrder: Unable to create Temporal client.", err)
	}
	defer c.Close()

	if r.Method == "GET" {

		// Populate form with current data:
		sorder := SOrderData{}
		sorder.WorkflowID = wkflId

		// Query variables form workflow
		resp, err := c.QueryWorkflow(context.Background(), sorder.WorkflowID, "", "payment.origin")
		if err != nil {
			log.Fatalln("AmendSOrder: Unable to query workflow", err)
		} else {
			var result interface{}
			if err := resp.Get(&result); err != nil {
				log.Fatalln("AmendSOrder: Unable to decode query result", err)
			} else {
				sorder.PaymentOrigin = fmt.Sprintf("%v", result)
			}
		}
		resp, err = c.QueryWorkflow(context.Background(), sorder.WorkflowID, "", "payment.destination")
		if err != nil {
			log.Fatalln("AmendSOrder: Unable to query workflow", err)
		} else {
			var result interface{}
			if err := resp.Get(&result); err != nil {
				log.Fatalln("AmendSOrder: Unable to decode query result", err)
			} else {
				sorder.PaymentDestination = fmt.Sprintf("%v", result)
			}
		}
		resp, err = c.QueryWorkflow(context.Background(), sorder.WorkflowID, "", "payment.amount")
		if err != nil {
			log.Fatalln("AmendSOrder: Unable to query workflow", err)

		} else {
			var result interface{}
			if err := resp.Get(&result); err != nil {
				log.Fatalln("AmendSOrder: Unable to decode query result", err)
			} else {
				sorder.PaymentAmount, _ = strconv.Atoi(fmt.Sprintf("%v", result))
			}
		}
		resp, err = c.QueryWorkflow(context.Background(), sorder.WorkflowID, "", "payment.reference")
		if err != nil {
			log.Fatalln("AmendSOrder: Unable to query workflow", err)
		} else {
			var result interface{}
			if err := resp.Get(&result); err != nil {
				log.Fatalln("AmendSOrder: Unable to decode query result", err)
			} else {
				sorder.PaymentReference = fmt.Sprintf("%v", result)
			}
		}
		resp, err = c.QueryWorkflow(context.Background(), sorder.WorkflowID, "", "schedule.periodduration")
		if err != nil {
			log.Fatalln("AmendSOrder: Unable to query workflow", err)
		} else {
			var result interface{}
			if err := resp.Get(&result); err != nil {
				log.Fatalln("AmendSOrder: Unable to decode query result", err)
			} else {
				d, _ := time.ParseDuration(fmt.Sprint(result))
				sorder.SchedulePeriodDuration = int(d.Seconds())
			}
		}

		log.Println("AmendSOrder: Displaying Standing Order:", sorder)

		u.Render(w, "templates/AmendSOrder.html", sorder)

	} else if r.Method == "POST" {

		// Form submission received with data to amend..

		r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
		wkflId := strings.TrimSpace(r.FormValue("wkflid"))
		amount, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("amount")))
		period, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("period")))
		reference := strings.TrimSpace(r.FormValue("reference"))

		log.Println("AmendSOrder: Processing new Standing Order: WorkflowId(", wkflId, "), Amount(", amount, "), Period(", period, "), Ref(", reference, ")")

		// Send Temporal Signals..

		// check workflow is active first
		foundwf, err := checkWorkflowActive(c, wkflId)
		if err != nil {
			log.Fatalln("AmendSOrder: checkWorkflowActive Failed,", err, "( found:", foundwf, ")")
		}

		if foundwf {
			// Signal the Workflow Executions to amend values:
			if amount != 0 {
				log.Println("AmendSOrder: Sending amend signal to workflow:", wkflId, "for sorderamount:", amount)
				err = c.SignalWorkflow(context.Background(), wkflId, "", "sorderamount", amount)
				if err != nil {
					log.Fatalln("AmendSOrder: Unable to signal workflow", err)
				}
			}
			if reference != "" {
				log.Println("AmendSOrder: Sending amend signal to workflow:", wkflId, "for sorderreference:", reference)
				err = c.SignalWorkflow(context.Background(), wkflId, "", "sorderreference", reference)
				if err != nil {
					log.Fatalln("AmendSOrder: Unable to signal workflow", err)
				}
			}
			if period != 0 {
				log.Println("AmendSOrder: Sending amend signal to workflow:", wkflId, "for sorderschedule:", period)
				err = c.SignalWorkflow(context.Background(), wkflId, "", "sorderschedule", period)
				if err != nil {
					log.Fatalln("AmendSOrder: Unable to signal workflow", err)
				}
			}
			u.Render(w, "templates/AmendSOrderPost.html", struct{ Success bool }{true})
		} else {
			u.Render(w, "templates/AmendSOrderPost.html", struct{ Success bool }{false})
		}
	}
}

/* CancelSOrder */
func CancelSOrder(w http.ResponseWriter, r *http.Request) {

	log.Println("CancelSOrder: called")

	// URL Parameters
	var wkflId string
	params := r.URL.Query()
	for k, v := range params {
		if k == "id" {
			wkflId = strings.Join(v, "")
		}
		log.Println("AmendSOrder: url params:", k, " => ", v)
	}

	log.Println("CancelSOrder: method:", r.Method) //get request method

	log.Println("CancelSOrder: Cancelling Standing Order:", wkflId)

	// Load the Temporal Cloud from env
	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("CancelSOrder: Failed to load Temporal Cloud environment: %v", err)
	}
	log.Println("CancelSOrder: connecting to temporal server..")
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("CancelSOrder: Unable to create Temporal client.", err)
	}
	defer c.Close()

	// check workflow is active first
	foundwf, err := checkWorkflowActive(c, wkflId)
	if err != nil {
		log.Fatalln("CancelSOrder: checkWorkflowActive Failed,", err, "( found:", foundwf, ")")
	}
	if foundwf {
		log.Println("CancelSOrder: Sending cancelsorder signal to workflow: ", wkflId)

		// Signal the Workflow Executions to cancel the standing order
		err = c.SignalWorkflow(context.Background(), wkflId, "", "cancelsorder", true)
		if err != nil {
			log.Fatalln("CancelSOrder: Unable to signal workflow", err)
		}

		u.Render(w, "templates/CancelSOrder.html", struct{ Success bool }{true})

	} else {
		log.Println("CancelSOrder: Standing Order not found/active:", wkflId)

		u.Render(w, "templates/CancelSOrder.html", struct{ Success bool }{false})
	}
}

/* checkWorkflowActive */
func checkWorkflowActive(c client.Client, wkflId string) (bool, error) {

	// Query using SearchAttribute
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	//query := "CustomStringField='ACTIVE-SORDER' and CloseTime is null"
	query := "CustomStringField='ACTIVE-SORDER' and ExecutionStatus = 'Running'"
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
			log.Fatalln("checkWorkflowActive: ListWorkflows returned an error,", err)
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
