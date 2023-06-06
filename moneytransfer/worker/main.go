package main

import (
  "log"

  "go.temporal.io/sdk/client"
  "go.temporal.io/sdk/worker"

  mt "webapp/moneytransfer"
)

func main() {
  log.Printf("%sGo worker starting..%s", mt.ColorGreen, mt.ColorReset)

  // Load the Temporal Cloud from env
  clientOptions, err := mt.LoadClientOption()
  if err != nil {
    log.Fatalf("Failed to load Temporal Cloud environment: %v", err)
  }

  log.Println("Go worker connecting to server..")

  c, err := client.Dial(clientOptions)
  if err != nil {
    log.Fatalln("Unable to create Temporal client.", err)
  }
  defer c.Close()

  log.Println("Go worker initialising..")
  w := worker.New(c, mt.MoneyTransferTaskQueueName, worker.Options{})

  // This worker hosts both Workflow and Activity functions.
  log.Println("Go worker registering for Workflow moneytransfer.Transfer..")
  w.RegisterWorkflow(mt.Transfer)

  log.Println("Go worker registering for Activity moneytransfer.Withdraw..")
  w.RegisterActivity(mt.Withdraw)

  log.Println("Go worker registering for Activity moneytransfer.Deposit..")
  w.RegisterActivity(mt.Deposit)

  log.Println("Go worker registering for Activity moneytransfer.Refund..")
  w.RegisterActivity(mt.Refund)

  // Start listening to the Task Queue.
  log.Printf("%sGo worker listening on %s task queue..%s", mt.ColorGreen, mt.MoneyTransferTaskQueueName, mt.ColorReset)
  err = w.Run(worker.InterruptCh())
  if err != nil {
    log.Fatalln("Unable to start MoneyTransfer Worker", err)
  }

  log.Printf("%sGo worker stopped.%s", mt.ColorGreen, mt.ColorReset)
}

