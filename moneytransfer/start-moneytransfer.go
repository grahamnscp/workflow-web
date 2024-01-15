package moneytransfer

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "math/rand"
  "strconv"

  "go.temporal.io/sdk/client"

  u "webapp/utils"
)

/*
 * StartMoneyTransfer - App entry point to run Temporal Workflows
 */
func StartMoneyTransfer(pmnt *PaymentDetails, txfrId int) (wfinfo *WorkflowInfo, starterr error) {

  thisid := fmt.Sprint(rand.Intn(99999))
  log.Printf("StartMoneyTransfer-%s: called, PaymentDetails: %#v", thisid, *pmnt)

  // Initialise return object
  wfinfo = &WorkflowInfo{
    Id:         txfrId,
    WorkflowID: fmt.Sprintf("go-txfr-webtask-wkfl-%s", thisid),
    RunID:      "",
    TaskQueue:  MoneyTransferTaskQueueName,
    Info:       "",
    Status:     "ERROR",
  }

  // Load the Temporal Cloud from env
  clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
  if err != nil {
    log.Printf("StartMoneyTransfer-%s: Failed to load Temporal Cloud environment: %v", thisid, err)
    wfinfo.Info = err.Error()
    return wfinfo, err
  }
  c, err := client.Dial(clientOptions)
  if err != nil {
    log.Printf("StartMoneyTransfer-%s: Unable to create Temporal client: %v", thisid, err)
    wfinfo.Info = err.Error()
    return wfinfo, err
  }
  defer c.Close()

  // Temporal Client Start Workflow Options
  workflowID := fmt.Sprintf("go-txfr-webtask-wkfl-%s", thisid)

  workflowOptions := client.StartWorkflowOptions{
    ID:        workflowID,
    TaskQueue: MoneyTransferTaskQueueName,
  }

  // delay between withdraw and deposit for demo purposes (seconds)
  var delay int
  delay, err = strconv.Atoi(fmt.Sprint(DelayTimerBetweenWithdrawDeposit))
  if err != nil {
    delay = 15
  }

  // ExecuteWorkflow moneytransfer.Transfer
  log.Printf("StartMoneyTransfer-%s: Starting moneytransfer workflow on %s task queue", thisid, MoneyTransferTaskQueueName)

  we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, TransferWorkflow, *pmnt, delay)

  if err != nil {
    log.Printf("StartMoneyTransfer-%s: Error, Unable to execute workflow %v", thisid, err)
    wfinfo.Info = err.Error()
    return wfinfo, err
  }
  wfinfo.WorkflowID = we.GetID()
  wfinfo.RunID = we.GetRunID()
  log.Printf("StartMoneyTransfer-%s: %sStarted workflow: WorkflowID: %s, RunID: %s%s",
    thisid, u.ColorYellow, wfinfo.WorkflowID, wfinfo.RunID, u.ColorReset)

   // write workflow start details to local table
    _ = UpdateTransferRequestStarted(wfinfo)


  // Check workflow status
  var result string

  err = we.Get(context.Background(), &result)

  if err != nil {
    log.Printf("StartMoneyTransfer-%s: %sWorkflow returned failure:%s %v", thisid, u.ColorRed, u.ColorReset, err)
    wfinfo.Info = err.Error()
    return wfinfo, err
  }

  data, err := json.MarshalIndent(result, "", "  ")
  if err != nil {
    log.Printf("StartMoneyTransfer-%s: Unable to format result in JSON format: %v", thisid, err)
    wfinfo.Info = err.Error()
    return wfinfo, err
  }
  log.Printf("StartMoneyTransfer-%s: %sWorkflow result: %s%s", thisid, u.ColorYellow, string(data), u.ColorReset)
  wfinfo.Info = trimQuotes(string(data))
  wfinfo.Status = "COMPLETED"

  // done
  log.Printf("StartMoneyTransfer-%s: done.", thisid)

  return wfinfo, err
}

/* UpdateTransferRequestStarted */
func UpdateTransferRequestStarted (wfinfo *WorkflowInfo) error {

  log.Println("UpdateTransferRequestStarted: called (Id:", wfinfo.Id, ")")

  // Get database connection
  dbc, _ := u.GetDBConnection()
  defer dbc.Close()

  sqlStatement := fmt.Sprintf("UPDATE moneytransfer.transfer SET t_wkfl_id='%s',t_run_id='%s',t_taskqueue='%s' WHERE id=%d", wfinfo.WorkflowID, wfinfo.RunID, wfinfo.TaskQueue, wfinfo.Id)
  stmtIns, dberr := dbc.Prepare(sqlStatement)
  if dberr != nil {
    log.Println("UpdateTransferRequestStarted: Prepare failed! ", dberr)
    return dberr
  }
  _, dberr = stmtIns.Exec()
  if dberr != nil {
    log.Println("UpdateTransferRequestStarted: update Exec failed! ", dberr)
    return dberr
  }

  return nil
}

/* trimQuotes from string */
func trimQuotes(s string) string {
  if len(s) >= 2 {
    if s[0] == '"' && s[len(s)-1] == '"' {
      return s[1 : len(s)-1]
    }
  }
  return s
}
