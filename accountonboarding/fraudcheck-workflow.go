package accountonboarding

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	u "webapp/utils"
)

/* Workflow - FraudCheckWorkflow */
func FraudCheckWorkflow(ctx workflow.Context, appform ApplicationForm) (string, error) {

	appworkflow := &ApplicationWFDetails{
		WFID:  workflow.GetInfo(ctx).WorkflowExecution.ID,
		RunID: workflow.GetInfo(ctx).WorkflowExecution.RunID,
		Token: "",
	}
	logger := workflow.GetLogger(ctx)
	logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Started", "-", appworkflow.WFID)

	// Workflow Variables
	application := &ApplicationForm{
		FirstName: appform.FirstName,
		Surname:   appform.Surname,
		Email:     appform.Email,
	}

	approvalStatus := &ApprovalStatus{
		Approver:          "",
		ApproverResponded: false,
		Approved:          false,
		Reason:            "",
		FraudRisk:         0,
	}

	var parentWorkflow string
	var workflowCancelled = false

	// Query Handlers
	// - query application details handler - fraud.applicationdetails
	QueryApplicationDetails := "fraud.applicationdetails"
	err := workflow.SetQueryHandler(ctx, QueryApplicationDetails, func() (ApplicationForm, error) {
		logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorCyan, "Received Query - QueryApplicationDetails:",
			application, u.ColorReset)
		return *application, nil
	})
	if err != nil {
		logger.Info("AccAppl-Workflow: SetQueryHandler: QueryApplicationDetails handler failed.", "Error", err)
		return "Error", err
	}

	// - query approval status handler - fraud.approvalstatus
	QueryApprovalStatus := "fraud.approvalstatus"
	err = workflow.SetQueryHandler(ctx, QueryApprovalStatus, func() (ApprovalStatus, error) {
		logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorCyan, "Received Query - QueryApprovalStatus:",
			approvalStatus, u.ColorReset)
		return *approvalStatus, nil
	})
	if err != nil {
		logger.Info("AccAppl-Workflow: SetQueryHandler: QueryApprovalStatus handler failed.", "Error", err)
		return "Error", err
	}

	// - query parent workflow id handler - fraud.parentworkflowid
	QueryParentWorkflowID := "fraud.parentworkflowid"
	err = workflow.SetQueryHandler(ctx, QueryParentWorkflowID, func() (string, error) {
		logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorCyan, "Received Query - QueryParentWorkflowID:",
			parentWorkflow, u.ColorReset)
		return parentWorkflow, nil
	})
	if err != nil {
		logger.Info("AccAppl-Workflow: SetQueryHandler: QueryParentWorkflowID handler failed.", "Error", err)
		return "Error", err
	}

	// Signal Handlers
	// - Fraud Approver Signal handler - fraudapproverresult
	selector := workflow.NewSelector(ctx)

	fraudapproverCkCh := workflow.GetSignalChannel(ctx, "fraudapproverresult")
	selector.AddReceive(fraudapproverCkCh, func(ch workflow.ReceiveChannel, _ bool) {

		// read contents from signal
		var fraudApproverResultSignal string
		ch.Receive(ctx, &fraudApproverResultSignal)

		logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorYellow, "Received Signal - fraudapproverresult:",
			fraudApproverResultSignal, u.ColorReset)

		// update workflow variables
		approvalStatus.ApproverResponded = true

		if fraudApproverResultSignal == "Approved" {
			approvalStatus.Approved = true
			approvalStatus.Reason = "Approved - Pass level 2"
		} else {
			approvalStatus.Approved = false
			approvalStatus.Reason = fraudApproverResultSignal
		}
	})

	// - Cancel signal handler - cancelfraudcheck
	cancelCh := workflow.GetSignalChannel(ctx, "cancelfraudcheck")
	selector.AddReceive(cancelCh, func(ch workflow.ReceiveChannel, _ bool) {

		var cancelFraudCheckSignal bool
		ch.Receive(ctx, &cancelFraudCheckSignal)

		logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorYellow, "Received Signal - cancelfraudcheck:",
			cancelFraudCheckSignal, u.ColorReset)

		// cancel workflow - from parent probably
		workflowCancelled = true
	})

	// Main
	logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Performing Fraud Check for -", application.Email)

	// upsert Fraud Check as ACTIVE
	_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "ACTIVE-FRAUDCK")

	// store parent workflow details for later signalling
	parentWorkflow = workflow.GetInfo(ctx).ParentWorkflowExecution.ID

	// demo delay - to slow things down
	workflow.AwaitWithTimeout(ctx, time.Duration(FraudCheckDelay)*time.Second, selector.HasPending)

	for selector.HasPending() {
		selector.Select(ctx)

		// First check if workflow has been cancelled and try up as required
		if workflowCancelled {
			logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Workflow Cancelled1 for -", application.Email)
			_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "CANCELLED-FRAUDCK")

			return "Complete - Workflow Cancelled1", nil
		}
	}

	// Activity: CheckFraudRisk
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})
	err = workflow.ExecuteActivity(ctx, CheckFraudRisk, application.Email).Get(ctx, &approvalStatus.FraudRisk)
	if err != nil {
		logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "CheckFraudRisk Activity returned:", err)
	}
	logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Email:", application.Email, "has Fraud Risk:", approvalStatus.FraudRisk)

	// route workflow based on fraud risk
	if approvalStatus.FraudRisk <= 5 {

		// Low Fraud Risk, just pass the check at level 1
		logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Fraud Check PASS1 for -", application.Email)
		_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "PASS-L1-FRAUDCK")

		signaled := workflow.SignalExternalWorkflow(ctx, parentWorkflow, "", "fraudcheckresult", "Pass1")

		signalErr := signaled.Get(ctx, nil)
		if signalErr != nil {
			// parent wf could have completed if other check failed
			return "Completed - Failed to Signal Parent - Check Pass1", nil
		}
		approvalStatus.Reason = "Low Risk - Pass level 1"

	} else {
		if approvalStatus.FraudRisk == 10 {

			// Risk level is 10 so auto Deny
			logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Level 2 Fraud Check - Auto Deny for -", application.Email)
			_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "AUTOFAILED-L2-FRAUDCK")

			approvalStatus.Approver = "Auto Deny"
			approvalStatus.Reason = "Auto L2 Deny"
			approvalStatus.Approved = false
			approvalStatus.ApproverResponded = true

		} else {

			// Fraud risk is high, perform level 2 interactive check
			logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Level 2 Fraud Check required for -", application.Email)
			_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "PENDING-L2-FRAUDCK")

			// Activity: GetFraudApprover
			err := workflow.ExecuteActivity(ctx, GetFraudApprover).Get(ctx, &approvalStatus.Approver)
			if err != nil {
				logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "GetFraudApprover Activity failed, err::", err)
			}
			logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Level 2 Fraud Check selected approver -", approvalStatus.Approver)

			// Local Activity: GenerateWorkflowToken for approval email url
			ao := workflow.LocalActivityOptions{
				StartToCloseTimeout: 10 * time.Second,
			}
			ctx = workflow.WithLocalActivityOptions(ctx, ao)
			err = workflow.ExecuteLocalActivity(ctx, GenerateWorkflowToken, appworkflow.WFID, appworkflow.RunID).Get(ctx, &appworkflow.Token)
			if err != nil {
				logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorRed, "GenerateWorkflowToken Local Activity failed, err:", err, u.ColorReset)
			}
			logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Approval email token:", appworkflow.Token)

			// Set Approval URL for approval email content
			application.URL = fmt.Sprintf("%s?ref=%s", urlBankFraudApproval, appworkflow.Token)
			logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Approver URL:", application.URL)

			// Activity: SendApprovalEmail
			var NotificationSuccessful bool
			err = workflow.ExecuteActivity(ctx, SendApprovalEmail, approvalStatus, application).Get(ctx, &NotificationSuccessful)
			if err != nil {
				logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorRed, "SendApprovalEmail Activity failed, err:", err, u.ColorReset)
			}
		}

		// Loop waiting for approval outcome
		for !approvalStatus.ApproverResponded {

			logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Level 2 Fraud Check waiting for Approval from -", approvalStatus.Approver)

			// long sleep waiting for signal
			// - TODO: send periodic reminder to reviewer
			workflow.AwaitWithTimeout(ctx, time.Duration(FraudCheckSleep)*time.Second, selector.HasPending)

			for selector.HasPending() {
				selector.Select(ctx)

				// Check if workflow has been cancelled and try up as required
				if workflowCancelled {
					logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Workflow Cancelled2 for -", application.Email)
					_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "CANCELLED-FRAUDCK")

					return "Complete - Workflow Cancelled2", nil
				}
			}
		}

		// Approver responded, process result..
		if approvalStatus.Approved {

			// Approver responded - workflow approved
			logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Fraud Check PASS2 for -", application.Email)
			_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "PASS-L2-FRAUDCK")

			signaled := workflow.SignalExternalWorkflow(ctx, parentWorkflow, "", "fraudcheckresult", "Pass2")

			signalErr := signaled.Get(ctx, nil)
			if signalErr != nil {
				// parent wf could have completed if other check failed
				return "Complete - Failed to Signal Parent - Check Pass2", nil
			}
		} else {

			// Approver responded workflow Fraud Check Failed or Risk 10 Auto Fail
			logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Fraud Check FAILED2 for -", application.Email)
			_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "FAILED-L2-FRAUDCK")

			signaled := workflow.SignalExternalWorkflow(ctx, parentWorkflow, "", "fraudcheckresult", "Failed")

			signalErr := signaled.Get(ctx, nil)
			if signalErr != nil {
				// parent wf could have completed if other check failed
				return "Complete - Failed to Signal Parent - Check Failed", nil
			}

		}
	}

	// Workflow Done
	logger.Info(u.ColorGreen, "FraudChk-Workflow:", u.ColorReset, "Complete - Approved:", approvalStatus.Approved, "Reason:",
		approvalStatus.Reason, "-", appworkflow.WFID)

	return fmt.Sprintf("Workflow Completed - Approved: %t, Reason: %s", approvalStatus.Approved, approvalStatus.Reason), nil
}
