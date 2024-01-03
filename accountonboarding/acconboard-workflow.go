package accountonboarding

import (
  "fmt"
  "time"

  enumspb "go.temporal.io/api/enums/v1"
  "go.temporal.io/sdk/workflow"

  u "webapp/utils"
)

/* Workflow - AccountApplicationWorkflow */
func AccountApplicationWorkflow(ctx workflow.Context, appform ApplicationForm) (string, error) {

  appworkflow := &ApplicationWFDetails{
    WFID:  workflow.GetInfo(ctx).WorkflowExecution.ID,
    RunID: workflow.GetInfo(ctx).WorkflowExecution.RunID,
    Token: "",
  }
  logger := workflow.GetLogger(ctx)
  logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Started", "-", appworkflow.WFID)

  // Workflow Variables
  application := &ApplicationForm{
    FirstName:   appform.FirstName,
    Surname:     appform.Surname,
    Email:       appform.Email,
    Account:     0,
    AccountName: "new",
    URL:         "",
  }
  appstatus := &ApplicationStatus{
    Status:    "Processing",
    Credit:    "Pending",
    Fraud:     "Pending",
    L1Appr:    "-",
    L2Appr:    "-",
    Approved:  false,
    Denied:    false,
    Cancelled: false,
  }

  // Declare Query and Signal handlers
  selector := workflow.NewSelector(ctx)

  // Query Handlers
  // - query application status handler
  QueryAppStatus := "application.status"
  err := workflow.SetQueryHandler(ctx, QueryAppStatus, func() (ApplicationStatus, error) {
    logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorCyan, "Received Query - QueryAppStatus:",
      appstatus, u.ColorReset)
    return *appstatus, nil
  })
  if err != nil {
    logger.Info("AccAppl-Workflow: SetQueryHandler: QueryAppStatus handler failed.", "Error", err)
    return "Error", err
  }

  // - query application status handler
  QueryAppDetails := "application.details"
  err = workflow.SetQueryHandler(ctx, QueryAppDetails, func() (ApplicationForm, error) {
    logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorCyan, "Received Query - QueryAppDetails:",
      application, u.ColorReset)
    return *application, nil
  })
  if err != nil {
    logger.Info("AccAppl-Workflow: SetQueryHandler: QueryAppDetails handler failed.", "Error", err)
    return "Error", err
  }

  // Signal Handlers
  // - Credit Check signal handler
  creditCkCh := workflow.GetSignalChannel(ctx, "creditcheckresult")
  selector.AddReceive(creditCkCh, func(ch workflow.ReceiveChannel, _ bool) {

    var creditCheckResultSignal string
    ch.Receive(ctx, &creditCheckResultSignal)

    logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorYellow, "Received Signal - creditcheckresult:",
      creditCheckResultSignal, u.ColorReset)

    appstatus.Credit = creditCheckResultSignal

    if appstatus.Credit == "BadCredit" {
      appstatus.Denied = true
    } // or "Clear"
  })

  // - Fraud Check signal handler
  fraudCkCh := workflow.GetSignalChannel(ctx, "fraudcheckresult")
  selector.AddReceive(fraudCkCh, func(ch workflow.ReceiveChannel, _ bool) {

    var fraudCheckResultSignal string
    ch.Receive(ctx, &fraudCheckResultSignal)

    logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorYellow, "Received Signal - fraudcheckresult:",
      fraudCheckResultSignal, u.ColorReset)

    // Process signal result
    switch fraudCheckResultSignal {
    case "Pass1":
      // Approved at level 1
      appstatus.Fraud = "Approved"
      appstatus.Denied = false
      appstatus.L1Appr = "L1Approved"
      appstatus.L2Appr = "Skipped"

    case "Pass2":
      // Approved at level 2
      appstatus.Fraud = "Approved"
      appstatus.Denied = false
      appstatus.L1Appr = "L1Approved"
      appstatus.L2Appr = "L2Approved"

    case "Failed":
      // Denied at Level 2 -> Application Denied
      appstatus.Fraud = "Failed"
      appstatus.Denied = true
      appstatus.L1Appr = "L1Approved"
      appstatus.L2Appr = "L2Denied"
    }
  })

  // - Cancel application with reason signal handler
  cancelCh := workflow.GetSignalChannel(ctx, "cancelapplication")
  selector.AddReceive(cancelCh, func(ch workflow.ReceiveChannel, _ bool) {

    var cancelApplicationSignal bool
    ch.Receive(ctx, &cancelApplicationSignal)

    logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorYellow, "Received Signal - cancelapplication:",
      cancelApplicationSignal, u.ColorReset)

    appstatus.Cancelled = true
  })

  // Main
  logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Processing application for -", application.Email)

  // upsert Account Application as ACTIVE
  _ = u.UpsertSearchAttribute(ctx, "CustomStringField", "ACTIVE-ACCAPP")

  // - TODO: Check for existing Inflight Application

  // Activity - Notification Email - Application Received
  ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: time.Minute,
  })
  var NotificationSuccessful bool
  err = workflow.ExecuteActivity(ctx, SendNotificationEmail, EmailNotificationStageReceived, application).Get(ctx, &NotificationSuccessful)
  if err != nil {
    logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorRed, "SendNotificationEmail Activity, stage:",
      EmailNotificationStageReceived, ", err:", err, u.ColorReset)
  }

  // Credit Check Child Workflow
  // - timer then approve unless in credit blacklist collection
  // - if credit check fails -> Reject Application with Credit reason
  creditCheckWFId := fmt.Sprintf("%s-credchk", appworkflow.WFID)
  childWFOptions := workflow.ChildWorkflowOptions{
    ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
    WorkflowID:        creditCheckWFId,
  }
  logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Launching Credit Check Child Workflow -", creditCheckWFId)

  cchildWFCtx := workflow.WithChildOptions(ctx, childWFOptions)
  workflow.ExecuteChildWorkflow(cchildWFCtx, CreditCheckWorkflow, application)

  // Fraud Check Child Workflow
  // - levels with status in search attribute
  // - automatic L1 approve if risk collection value <=5 -> level 1 on timer -> approve or next level approval
  // - level 2 approval - manual approval -> email activity for approval (with url link token to reenter via web ui)
  // - level 2 approve or deny with reason - signal parent wf
  fraudCheckWFId := fmt.Sprintf("%s-fraudchk", appworkflow.WFID)
  fchildWFOptions := workflow.ChildWorkflowOptions{
    ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
    WorkflowID:        fraudCheckWFId,
  }
  logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Launching Fraud Check Child Workflow -", fraudCheckWFId)

  fchildWFCtx := workflow.WithChildOptions(ctx, fchildWFOptions)
  workflow.ExecuteChildWorkflow(fchildWFCtx, FraudCheckWorkflow, application)

  // Activity - Notification Email - Application InProgress
  err = workflow.ExecuteActivity(ctx, SendNotificationEmail, EmailNotificationStageInProgress, application).Get(ctx, &NotificationSuccessful)
  if err != nil {
    logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorRed, "SendNotificationEmail Activity, stage:",
      EmailNotificationStageInProgress, ", err:", err, u.ColorReset)
  }

  // Loop unless cancelled or complete
  for !appstatus.Cancelled {

    // sleep waiting for signal
    workflow.AwaitWithTimeout(ctx, time.Duration(AccOnboardingSleep)*time.Second, selector.HasPending)

    // Check if cancel signal received during period (will interrupt Sleep)
    for selector.HasPending() {
      selector.Select(ctx)

      if appstatus.Cancelled {
        logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Cancelled", "-", appworkflow.WFID)

        // upsert Account Application as CANCELLED
        _ = u.UpsertSearchAttribute(ctx, "CustomStringField", "CANCELLED-ACCAPP")

        // Activity - Notification Email - Cancelled
        err := workflow.ExecuteActivity(ctx, SendNotificationEmail, EmailNotificationStageCancelled, application).Get(ctx, &NotificationSuccessful)
        if err != nil {
          logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorRed, "SendNotificationEmail Activity, stage:",
            EmailNotificationStageCancelled, ", err:", err, u.ColorReset)
        }

        // Cancel Fraud child workflow - use signal to fraudCheckWFId cancelfraudcheck
        logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Cancelling Fraud Child Workflow -", fraudCheckWFId)
        signaled := workflow.SignalExternalWorkflow(ctx, fraudCheckWFId, "", "cancelfraudcheck", true)

        signalErr := signaled.Get(ctx, nil)
        if signalErr != nil {
          logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Failed to signal Cancel to Fraud Child Workflow -", fraudCheckWFId)
        }

        // slight delay before exiting to allow return from child workflow
        workflow.Sleep(ctx, time.Duration(5)*time.Second)

        // Exit workflow
        return "Workflow Cancelled.", nil

      } else {
        // Workflow is not cancelled, process post signal received..
        logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Processing - App Status: (", *appstatus, ")")

        // Check status of all checks
        if (appstatus.Credit) == "Clear" && (appstatus.Fraud == "Approved") {
          appstatus.Approved = true
        }

        if appstatus.Approved {

          logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Application Approved")
          _ = u.UpsertSearchAttribute(ctx, "CustomStringField", "APPROVED-ACCAPP")

          // Activity - Provision Account
          bankaccount := BankAccount{
            AccountName: application.FirstName,
            AccountNum:  0,
          }
          err = workflow.ExecuteActivity(ctx, ProvisionAccount, application).Get(ctx, &bankaccount)
          if err != nil {
            logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorRed, "ProvisionAccount Activity, err:", err, u.ColorReset)
          }

          // Add application.Account and application.URL for details access page and add balance
          application.Account = bankaccount.AccountNum
          application.AccountName = bankaccount.AccountName
          application.URL = fmt.Sprintf("%s?ref=%d", urlBankUserHome, application.Account)

          // Activity - Notification Email - Complete: Approved
          err := workflow.ExecuteActivity(ctx, SendNotificationEmail, EmailNotificationStageComplete, application).Get(ctx, &NotificationSuccessful)
          if err != nil {
            logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorRed, "SendNotificationEmail Activity, stage:",
              EmailNotificationStageComplete, ", err:", err, u.ColorReset)
          }

          // Exit workflow
          logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Complete - Approved", "-", appworkflow.WFID)
          return "Workflow Completed - Application Approved", nil

        } else if appstatus.Denied {

          logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Application Denied")
          _ = u.UpsertSearchAttribute(ctx, "CustomStringField", "DENIED-ACCAPP")

          // Activity - Notification Email - Rejected
          err := workflow.ExecuteActivity(ctx, SendNotificationEmail, EmailNotificationStageRejected, application).Get(ctx, &NotificationSuccessful)
          if err != nil {
            logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorRed, "SendNotificationEmail Activity, stage:",
              EmailNotificationStageRejected, ", err:", err, u.ColorReset)
          }

          // slight delay before exiting to allow return from child workflow
          workflow.Sleep(ctx, time.Duration(5)*time.Second)

          logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Complete - Denied", "-", appworkflow.WFID)
          return "Workflow Completed - Application Denied", nil
        }
      }
    }
  }

  // Workflow Done
  logger.Info(u.ColorGreen, "AccAppl-Workflow:", u.ColorReset, "Completed for -", appworkflow.WFID)
  return "Workflow Completed.", nil
}
