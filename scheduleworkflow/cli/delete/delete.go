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

	// parse cli args
	schid, err := cliparseCLIArgs(os.Args[1:])
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

	// just delete it!
	log.Println("Deleting schedule ScheduleID:", scheduleHandle.GetID())
	scheduleHandle.Delete(ctx)

}

func cliparseCLIArgs(args []string) (*string, error) {
	set := flag.NewFlagSet("cli", flag.ExitOnError)

	schid := set.String("scheduleid", "", "Schedule Id to act on")

	if err := set.Parse(args); err != nil {
		return nil, fmt.Errorf("failed parsing args: %w", err)
	} else if *schid == "" {
		return nil, fmt.Errorf("--scheduleid argument is required")
	}
	return schid, nil
}
