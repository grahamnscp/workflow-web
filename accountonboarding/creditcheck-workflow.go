package accountonboarding

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	u "webapp/utils"
)

/* Workflow - CreditCheckWorkflow */
func CreditCheckWorkflow(ctx workflow.Context, appform ApplicationForm) (string, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info(u.ColorGreen, "CreditChk-Workflow:", u.ColorReset, "Started", "-", workflow.GetInfo(ctx).WorkflowExecution.ID)

	// Workflow Variables
	application := &ApplicationForm{
		FirstName: appform.FirstName,
		Surname:   appform.Surname,
		Email:     appform.Email,
	}

	// upsert Account Application as ACTIVE
	_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "ACTIVE-CREDCK")

	logger.Info(u.ColorGreen, "CreditChk-Workflow:", u.ColorReset, "Performing Credit Check for -", application.Email)

	workflow.Sleep(ctx, time.Duration(CreditCheckDelay)*time.Second)

	parent := workflow.GetInfo(ctx).ParentWorkflowExecution

	// Lookup applicant in Credit Blacklist
	blackListed := false
	ctx2 := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})
	var blackListedResult bool
	err := workflow.ExecuteActivity(ctx2, CheckBlacklist, application.Email).Get(ctx, &blackListedResult)
	if err != nil {
		logger.Info(u.ColorGreen, "CreditChk-Workflow:", u.ColorReset, "CheckBlacklist Activity returned:", err)
	}
	blackListed = blackListedResult
	logger.Info(u.ColorGreen, "CreditChk-Workflow:", u.ColorReset, "Email:", application.Email, "in blacklist:", blackListedResult)

	if !blackListed {

		// Check Clear
		logger.Info(u.ColorGreen, "CreditChk-Workflow:", u.ColorReset, "Credit Check CLEAR for -", application.Email)
		_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "CLEAR-CREDCK")

		signaled := workflow.SignalExternalWorkflow(ctx, parent.ID, "", "creditcheckresult", "Clear")

		signalErr := signaled.Get(ctx, nil)
		if signalErr != nil {
			// parent wf could have completed if other check failed
			return "Complete - Failed to Signal Parent - Clear", nil
		}

	} else {
		// Applicant is blacklisted
		logger.Info(u.ColorGreen, "CreditChk-Workflow:", u.ColorReset, "Credit Check FAILED for -", application.Email)
		_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "BAD-CREDCK")

		signaled := workflow.SignalExternalWorkflow(ctx, parent.ID, "", "creditcheckresult", "BadCredit")

		signalErr := signaled.Get(ctx, nil)
		if signalErr != nil {
			// parent wf could have completed if other check failed
			return "Complete - Failed to Signal Parent - Bad Credit", nil
		}
	}

	// Workflow Done
	logger.Info(u.ColorGreen, "CreditChk-Workflow:", u.ColorReset, "Complete", "-", workflow.GetInfo(ctx).WorkflowExecution.ID)
	return fmt.Sprintf("Workflow Completed - blacklisted: %t", blackListed), nil
}
