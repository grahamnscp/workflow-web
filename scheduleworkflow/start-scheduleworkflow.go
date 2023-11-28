package scheduleworkflow

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"go.temporal.io/sdk/client"

	u "webapp/utils"
)

func StartScheduleWorkflow(sd ScheduleDetails) (starterr error) {

	thisid := fmt.Sprint(rand.Intn(99999))
	log.Printf("StartScheduleWorkflow-%s: %scalled, email: %s, description: %s, minutes: %d%s\n", thisid, u.ColorYellow, sd.Email, sd.Description, sd.Minutes, u.ColorReset)

	ctx := context.Background()

	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("StartScheduleWorkflow-%s: Failed to load Temporal Cloud environment, err: %v", thisid, err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalf("StartScheduleWorkflow-%s: Unable to create Temporal client, err: %v", thisid, err)
	}
	defer c.Close()

	// define schedule and workflow ID values
	scheduleID := "schedule_" + thisid
	workflowID := "schedule_workflow_" + thisid

	sd.Id = scheduleID

	// pass the ScheduleDetails struct as argument to workflow
	args := []interface{}{sd}

	// Create the schedule, starting with no spec means the schedule will not run.
	scheduleHandle, err := c.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID:   scheduleID,
		Spec: client.ScheduleSpec{},
		Action: &client.ScheduleWorkflowAction{
			ID:        workflowID,
			Workflow:  ScheduleWorkflow,
			Args:      args,
			TaskQueue: ScheduleWFTaskQueueName,
		},
	})
	if err != nil {
		log.Fatalf("StartScheduleWorkflow-%s: Unable to create schedule, err: %v", thisid, err)
	}

	// Update the schedule with a spec to run periodically,
	log.Printf("StartScheduleWorkflow-%s: %sUpdating schedule, ScheduleID: %s to run %d minutes past the hour%s\n", thisid, u.ColorYellow, scheduleHandle.GetID(), sd.Minutes, u.ColorReset)

	err = scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(schedule client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			schedule.Description.Schedule.Spec = &client.ScheduleSpec{
				// Run the schedule at 5pm on Friday
				Calendars: []client.ScheduleCalendarSpec{
					{
						// Year range to match.
						// default: empty that matches all years
						//Year []ScheduleRange

						// Month range to match (1-12)
						// default: matches all months
						//Month []ScheduleRange

						// DayOfMonth range to match (1-31)
						// default: matches all days
						//DayOfMonth []ScheduleRange

						// DayOfWeek range to match (0-6; 0 is Sunday)
						// default: matches all days of the week
						//DayOfWeek []ScheduleRange
						//DayOfWeek: []client.ScheduleRange{
						//  {
						//    Start: 1,
						//    End:   5,
						//  },
						//},

						// Hour range to match (0-23).
						// default: matches 0
						//Hour []ScheduleRange
						Hour: []client.ScheduleRange{
							{
								Start: 0,
								End:   23,
							},
						},

						// Minute range to match (0-59).
						// default: matches 0
						//Minute []ScheduleRange
						Minute: []client.ScheduleRange{
							{
								Start: sd.Minutes,
							},
						},

						// Second range to match (0-59).
						// default: matches 0
						//Second []ScheduleRange

						// Comment - Description of the intention of this schedule.
						//Comment string
						Comment: sd.Description,
					},
				},
				// Run the schedule every 5 minutes
				//Intervals: []client.ScheduleIntervalSpec{
				//  {
				//    Every: 5 * time.Minute,
				//  },
				//},
			}
			// Start the schedule paused initially
			schedule.Description.Schedule.State.Paused = true
			//schedule.Description.Schedule.State.LimitedActions = true
			//schedule.Description.Schedule.State.RemainingActions = 3

			return &client.ScheduleUpdate{
				Schedule: &schedule.Description.Schedule,
			}, nil
		},
	})
	if err != nil {
		log.Fatalf("StartScheduleWorkflow-%s: Unable to update schedule, err: %v", thisid, err)
	}

	// Unpause schedule (as created State.Paused initially)
	log.Printf("StartScheduleWorkflow-%s: %sUnpausing schedule, ScheduleID: %s%s\n", thisid, u.ColorYellow, scheduleHandle.GetID(), u.ColorReset)

	err = scheduleHandle.Unpause(ctx, client.ScheduleUnpauseOptions{})
	if err != nil {
		log.Fatalf("StartScheduleWorkflow-%s: Unable to unpause schedule, err: %v", thisid, err)
	}

	// Run schedule for a number of actions:
	//  ref: schedule.Description.Schedule.State.LimitedActions = true
	//       schedule.Description.Schedule.State.RemainingActions = 3
	//
	//log.Println("StartScheduleWorkflow: Waiting for schedule to complete 3 actions", "ScheduleID", scheduleHandle.GetID())
	//
	//for {
	//  description, err := scheduleHandle.Describe(ctx)
	//  if err != nil {
	//    log.Fatalln("StartScheduleWorkflow: Unable to describe schedule", err)
	//  }
	//  if description.Schedule.State.RemainingActions != 0 {
	//    log.Printf("StartScheduleWorkflow-%s: Schedule (%s) has %d remaining actions", thisid, scheduleHandle.GetID(), description.Schedule.State.RemainingActions)
	//    time.Sleep(15 * time.Minute)
	//  } else {
	//    break
	//  }
	//}

	return nil
}
