package moneytransfer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"

	u "webapp/utils"
)

/*
 * QueryMoneyTransfer - App entry point to run Temporal Workflows
 */
func QueryMoneyTransfer(w http.ResponseWriter, wfinfo *WorkflowInfo) (err error) {

	log.Printf("QueryMoneyTransfers: called")

	// Load the Temporal Cloud from env
	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Printf("QueryMoneyTransfer: Failed to load Temporal Cloud environment: %v", err)
		return err
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Printf("QueryMoneyTransfer: Unable to create Temporal client: %v", err)
		return err
	}
	defer c.Close()

	// Fetch the execution for the workflow
	execution := &commonpb.WorkflowExecution{
		WorkflowId: wfinfo.WorkflowID,
		RunId:      wfinfo.RunID,
	}
	log.Printf("QueryMoneyTransfer: execution: %v", *execution)

	history, err := getHistory(c, context.Background(), execution)
	if err != nil {
		log.Printf("QueryMoneyTransfer: getHistory error: %v", err)
		return err
	}
	historyEventCount := len(history)
	if historyEventCount == 0 {
		// Nothing to recover
		log.Printf("QueryMoneyTransfer: history len 0")
		return err
	}
	log.Printf("QueryMoneyTransfer: %d %v events retrieved", historyEventCount, reflect.TypeOf(history))

	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Workflow History for: %s, %s</h1>", wfinfo.WorkflowID, wfinfo.RunID)

	for h := range history {
		// type: history.HistoryEvent
		fmt.Printf("----------\nQueryMoneyTransfer: history[%d]:\n", h)
		fmt.Fprintln(w, "<h2>  History:", h, "</h2>")

		pc := converter.NewJSONPayloadConverter()
		payload, _ := pc.ToPayload(history[h])

		fmt.Printf("-----\n")
		jsondatastr := string(payload.Data)

		// Unmarshall json data into a map container for decoded the JSON structure into
		var c map[string]interface{}
		err := json.Unmarshal([]byte(jsondatastr), &c)
		if err != nil {
			panic(err)
		}
		event_type_num := int32(c["event_type"].(float64))
		event_type := string(enumspb.EventType_name[event_type_num])
		//task_id := int(c["task_id"].(float64))

		fmt.Printf("%sWorkflow Event %d: %s%s\n", u.ColorGreen, event_type_num, event_type, u.ColorReset)
		fmt.Fprintln(w, "<h3>    Workflow Event: ", event_type, "</h3>")

		switch event_type_num {
		case enumspb.EventType_value["WorkflowExecutionStarted"]:
			//fmt.Printf("case WorkflowExecutionStarted\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

			attributes := c["Attributes"].(map[string]interface{})["workflow_execution_started_event_attributes"]
			fmt.Printf("attributes: %v\n", attributes)
			fmt.Fprintln(w, "<br><font color=green><b>attributes: ", attributes, "</b></font><br>")

		case enumspb.EventType_value["WorkflowTaskScheduled"]:
			//fmt.Printf("case WorkflowTaskScheduled\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

		case enumspb.EventType_value["WorkflowTaskStarted"]:
			//fmt.Printf("case WorkflowTaskStarted\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

		case enumspb.EventType_value["WorkflowTaskCompleted"]:
			//fmt.Printf("case WorkflowTaskCompleted\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

		case enumspb.EventType_value["UpsertWorkflowSearchAttributes"]:
			//fmt.Printf("case UpsertWorkflowSearchAttributes\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

			attributes := c["Attributes"].(map[string]interface{})["upsert_workflow_search_attributes_event_attributes"]
			fmt.Printf("attributes: %v\n", attributes)
			search_attribute := attributes.(map[string]interface{})["search_attributes"].(map[string]interface{})["indexed_fields"]
			fmt.Printf("search_attribute: %v\n", search_attribute)

			attr := search_attribute.(map[string]interface{})["CustomStringField"].(map[string]interface{})["data"].(string)
			//fmt.Fprintln(w, "<font color=green><b>search_attribute: ", search_attribute, "</b></font><br>")
			fmt.Fprintf(w, "<br><font color=green><b>search_attribute: CustomStringFieldi: %s </b></font><br>", u.DecodeB64(attr))

		case enumspb.EventType_value["ActivityTaskScheduled"]:
			//fmt.Printf("case ActivityTaskScheduled\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

			attributes := c["Attributes"].(map[string]interface{})["activity_task_scheduled_event_attributes"]
			fmt.Printf("attributes: %v\n", attributes)
			activity_type := attributes.(map[string]interface{})["activity_type"].(map[string]interface{})["name"].(string)
			fmt.Printf("activity_type: %s\n", activity_type)
			fmt.Fprintln(w, "<br><font color=green><b>activity_type: ", activity_type, "</b></font><br>")

		case enumspb.EventType_value["ActivityTaskStarted"]:
			//fmt.Printf("case ActivityTaskStarted\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

		case enumspb.EventType_value["ActivityTaskCompleted"]:
			//fmt.Printf("case ActivityTaskCompleted\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

		case enumspb.EventType_value["ActivityTaskFailed"]:
			//fmt.Printf("case ActivityTaskFailed\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

		case enumspb.EventType_value["TimerStarted"]:
			//fmt.Printf("case TimerStarted\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

			attributes := c["Attributes"].(map[string]interface{})["timer_started_event_attributes"]
			fmt.Printf("attributes: %v\n", attributes)
			fmt.Fprintln(w, "<br><font color=green><b>attributes: ", attributes, "</b></font><br>")

		case enumspb.EventType_value["TimerFired"]:
			//fmt.Printf("case TimerFired\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

			attributes := c["Attributes"].(map[string]interface{})["timer_fired_event_attributes"]
			fmt.Printf("attributes: %v\n", attributes)
			fmt.Fprintln(w, "<font color=green><b>attributes: ", attributes, "</b></font><br>")

		case enumspb.EventType_value["WorkflowExecutionCompleted"]:
			//fmt.Printf("case WorkflowExecutionCompleted\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

			attributes := c["Attributes"].(map[string]interface{})["workflow_execution_completed_event_attributes"]
			fmt.Printf("attributes: %v\n", attributes)
			fmt.Fprintln(w, "<font color=green><b>attributes: ", attributes, "</b></font><br>")

		case enumspb.EventType_value["WorkflowExecutionFailed"]:
			//fmt.Printf("case WorkflowExecutionFailed\n")
			fmt.Printf("%sEvent Data: %v%s\n", u.ColorGreen, jsondatastr, u.ColorReset)
			fmt.Fprintln(w, "<font color=blue>Event Data: ", jsondatastr, "</font><br>")

			attributes := c["Attributes"].(map[string]interface{})["workflow_execution_failed_event_attributes"]
			fmt.Printf("attributes: %v\n", attributes)
			fmt.Fprintln(w, "<font color=green><b>attributes: ", attributes, "</b></font><br>")
		}
	}
	fmt.Fprintln(w, "<br><br><br><br>")

	// done
	if err != nil {
		log.Printf("QueryMoneyTransfer: done, err: %v", err)
	} else {
		log.Printf("QueryMoneyTransfer: done")
	}
	return err

}

/* getHistory */
func getHistory(c client.Client, ctx context.Context, execution *commonpb.WorkflowExecution) ([]*historypb.HistoryEvent, error) {

	log.Printf("QueryMoneyTransfer: _getHistory called")
	//fmt.Printf("enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT has type: %v\n", reflect.TypeOf(enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT))
	iter := c.GetWorkflowHistory(ctx,
		execution.GetWorkflowId(),
		execution.GetRunId(),
		false,
		enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var events []*historypb.HistoryEvent
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
