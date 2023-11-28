package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	sw "webapp/scheduleworkflow"

	"go.temporal.io/sdk/client"

	u "webapp/utils"
)

/* ListSchedules */
func ListSchedules(w http.ResponseWriter, r *http.Request) {

	log.Println("ListSchedules: called")

	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("ListSchedules: Failed to load Temporal Cloud environment: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("ListSchedules: Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()

	schedDets := sw.ScheduleDetails{}
	schedDetsList := []sw.ScheduleDetails{}

	// list schedules (type: *internal.scheduleListIteratorImpl)
	listView, err := c.ScheduleClient().List(ctx, client.ScheduleListOptions{
		PageSize: 10,
	})
	if err != nil {
		fmt.Println("ListSchedules: Failed to get list of schedules, ", err)
	}
	for listView.HasNext() {
		schedulelistentry, _ := listView.Next() // (type: *internal.ScheduleListEntry)
		schedDets.Id = schedulelistentry.ID

		// get scheduleHandle using id
		scheduleHandle := c.ScheduleClient().GetHandle(ctx, schedDets.Id)
		description, err := scheduleHandle.Describe(ctx)
		if err != nil {
			fmt.Println("ListSchedules: Failed to get scheduleHandle.Describe, ", err)
		}
		schedDets.Description = description.Schedule.Spec.Calendars[0].Comment
		schedDets.Minutes = description.Schedule.Spec.Calendars[0].Minute[0].Start

		schedDetsList = append(schedDetsList, schedDets)
	}

	u.Render(w, "templates/ListSchedules.html", schedDetsList)
}

/* NewSchedule */
func NewSchedule(w http.ResponseWriter, r *http.Request) {

	log.Println("NewSchedule: called")

	log.Println("NewSchedule: method:", r.Method) //get request method
	if r.Method == "GET" {
		u.Render(w, "templates/NewSchedule.html", nil)
		return
	}

	var sd sw.ScheduleDetails

	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	sd.Email = r.FormValue("email")
	sd.Description = r.FormValue("description")
	minutes, _ := strconv.Atoi(r.FormValue("minutes"))
	sd.Minutes = minutes

	log.Printf("NewSchedule: Creating new scheduleworkflow for: %s, comment: %s, minutes: %d", sd.Email, sd.Description, sd.Minutes)

	err := sw.StartScheduleWorkflow(sd)

	if err != nil {
		log.Println("NewSchedule: StartScheduleWorkflow returned error,", err)
	}
	u.Render(w, "templates/NewSchedule.html", struct{ Success bool }{true})
}

/* ShowSchedule Details, template calls update post and delete*/
func ShowSchedule(w http.ResponseWriter, r *http.Request) {

	log.Println("ShowSchedule: called")

	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	schedDets := sw.ScheduleDetails{}
	schedDets.Id = r.FormValue("id")

	// Connect to Temporal
	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("ShowSchedule: Failed to load Temporal Cloud environment: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("ShowSchedule: Unable to create client", err)
	}
	defer c.Close()

	// get scheduleHandle using schedule id
	ctx := context.Background()
	scheduleHandle := c.ScheduleClient().GetHandle(ctx, schedDets.Id)
	description, err := scheduleHandle.Describe(ctx)
	if err != nil {
		fmt.Println("ShowSchedule: Failed to get scheduleHandle.Describe, ", err)
	}
	schedDets.Description = description.Schedule.Spec.Calendars[0].Comment
	schedDets.Minutes = description.Schedule.Spec.Calendars[0].Minute[0].Start

	u.Render(w, "templates/ShowSchedule.html", schedDets)
}

/* UpdateSchedule */
func UpdateSchedule(w http.ResponseWriter, r *http.Request) {

	log.Println("UpdateSchedule: called")

	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	scheduleID := r.FormValue("id")
	comment := r.FormValue("description")
	minutes, _ := strconv.Atoi(r.FormValue("minutes"))

	log.Printf("UpdateSchedule: scheduleID: %s, Description: %s, Minutes: %d", scheduleID, comment, minutes)

	// Connect to Temporal
	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("UpdateSchedule: Failed to load Temporal Cloud environment: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("UpdateSchedule: Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()

	// get scheduleHandle using id
	scheduleHandle := c.ScheduleClient().GetHandle(ctx, scheduleID)

	if scheduleHandle == nil {
		log.Printf("UpdateSchedule: Unable to read schedule for ScheduleID: %s\n", scheduleID)
		u.Render(w, "templates/UpdateSchedulePost.html", struct{ Success bool }{false})
	}

	// update the schedule
	log.Printf("UpdateSchedule: Updating schedule ScheduleID: %s -\n", scheduleHandle.GetID())

	// fetch existing Schedule Description
	description, err := scheduleHandle.Describe(ctx)
	if err != nil {
		fmt.Println("UpdateSchedule: Failed to get scheduleHandle.Describe, ", err)
	}

	// Update the the existing description struct with new values
	if comment != "" {
		log.Printf("UpdateSchedule: Updating schedule Description: '%s'\n", comment)
		description.Schedule.Spec.Calendars[0].Comment = comment
	}
	if minutes != 0 {
		log.Printf("UpdateSchedule: Updating schedule Minutes: %d\n", minutes)
		description.Schedule.Spec.Calendars[0].Minute[0].Start = minutes
		description.Schedule.Spec.Calendars[0].Minute[0].End = minutes
	}

	// Update the schedule with modified Description
	var sui = client.ScheduleUpdateInput{
		Description: *description,
	}

	err = scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(schedule client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			schedule = sui
			return &client.ScheduleUpdate{
				Schedule: &schedule.Description.Schedule,
			}, nil
		},
	})
	if err != nil {
		fmt.Println("UpdateSchedule: Failed to update schedule via DoUpdate, ", err)
	}

	u.Render(w, "templates/UpdateSchedulePost.html", struct{ Success bool }{true})
}

/* DeleteSchedule */
func DeleteSchedule(w http.ResponseWriter, r *http.Request) {

	log.Println("DeleteSchedule: called")

	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	scheduleID := r.FormValue("id")

	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("DeleteSchedule: Failed to load Temporal Cloud environment: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("DeleteSchedule: Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()

	// get scheduleHandle using id
	scheduleHandle := c.ScheduleClient().GetHandle(ctx, scheduleID)

	// just delete it!
	log.Println("DeleteSchedule: Deleting ScheduleID:", scheduleHandle.GetID())
	scheduleHandle.Delete(ctx)

	u.Render(w, "templates/DeleteSchedule.html", struct{ Success bool }{true})
}
