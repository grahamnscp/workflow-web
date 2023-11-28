package main

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"

	u "webapp/utils"
)

func main() {

	// temporal client
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

	// list schedules (type: *internal.scheduleListIteratorImpl)
	listView, err := c.ScheduleClient().List(ctx, client.ScheduleListOptions{
		PageSize: 10,
	})
	if err != nil {
		fmt.Println("Failed to get list of schedules, ", err)
	}
	for listView.HasNext() {
		schedulelistentry, _ := listView.Next() // (type: *internal.ScheduleListEntry)
		scheduleID := schedulelistentry.ID
		fmt.Println("Schedule.ID:", scheduleID)

		// get scheduleHandle using id
		scheduleHandle := c.ScheduleClient().GetHandle(ctx, scheduleID)
		description, err := scheduleHandle.Describe(ctx)
		if err != nil {
			fmt.Println("Failed to get scheduleHandle.Describe, ", err)
		}
		// access some info
		//fmt.Println("schedule description.Schedule.Action:", description.Schedule.Action)

		fmt.Println("  description.Schedule.Spec:", description.Schedule.Spec)
		fmt.Println("  description.Schedule.Spec.Calendars.Comment:", description.Schedule.Spec.Calendars[0].Comment)
		fmt.Printf("  description.Schedule.Spec.Calendars.Hours: %d-%d\n", description.Schedule.Spec.Calendars[0].Hour[0].Start, description.Schedule.Spec.Calendars[0].Hour[0].End)
		fmt.Println("  description.Schedule.Spec.Calendars.Minute.Start:", description.Schedule.Spec.Calendars[0].Minute[0].Start)

		//fmt.Println("schedule description.Schedule.Policy:", description.Schedule.Policy)
		//fmt.Println("schedule description.Schedule.State:", description.Schedule.State)

		//fmt.Println("schedule description.Schedule.State.RemainingActions:", description.Schedule.State.RemainingActions)
	}
}
