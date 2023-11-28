package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"go.temporal.io/sdk/client"

	u "webapp/utils"
)

func main() {

	// parse cli args
	schid, comment, minutes, err := cliparseCLIArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("scheduleid parameter not obtained from cli: %v", err)
	}
	scheduleID := *schid

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

	// get scheduleHandle using id
	scheduleHandle := c.ScheduleClient().GetHandle(ctx, scheduleID)

	// update the schedule
	log.Printf("Updating schedule ScheduleID: %s -\n", scheduleHandle.GetID())

	// fetch existing Schedule Description
	description, err := scheduleHandle.Describe(ctx)
	if err != nil {
		fmt.Println("Failed to get scheduleHandle.Describe, ", err)
	}
	// Update the Description
	//..

	if *comment != "" {
		log.Printf("Updating schedule Description: '%s'\n", *comment)
		description.Schedule.Spec.Calendars[0].Comment = *comment
	}
	if *minutes != "" {
		mins, _ := strconv.Atoi(*minutes)
		log.Printf("Updating schedule Minutes: %d\n", mins)
		description.Schedule.Spec.Calendars[0].Minute[0].Start = mins
		description.Schedule.Spec.Calendars[0].Minute[0].End = mins
	}

	var sui = client.ScheduleUpdateInput{
		Description: *description,
	}

	// Update the schedule with modified Description
	err = scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(schedule client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			schedule = sui
			return &client.ScheduleUpdate{
				Schedule: &schedule.Description.Schedule,
			}, nil
		},
	})
	if err != nil {
		fmt.Println("Failed to update schedule via DoUpdate, ", err)
	}
}

func cliparseCLIArgs(args []string) (*string, *string, *string, error) {
	set := flag.NewFlagSet("cli", flag.ExitOnError)

	schid := set.String("scheduleid", "", "Schedule Id to act on")

	comment := set.String("description", "", "Schedule Calendar Comment Description Field")
	minutes := set.String("minutes", "", "Schedule Calendar Minutes Start Field")

	if err := set.Parse(args); err != nil {
		return nil, nil, nil, fmt.Errorf("failed parsing args: %w", err)
	} else if *schid == "" {
		return nil, nil, nil, fmt.Errorf("--scheduleid argument is required")
	}
	return schid, comment, minutes, nil
}
