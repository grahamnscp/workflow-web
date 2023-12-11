package standingorder

import (
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/workflow"

	mt "webapp/moneytransfer"
	u "webapp/utils"
)

func StandingOrderWorkflow(ctx workflow.Context, pdetails mt.PaymentDetails, pschedule PaymentSchedule) (string, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorReset, "Started", "-", workflow.GetInfo(ctx).WorkflowExecution.ID)

	// Counter for continue-as-new
	var transferCounter = 0

	// local workflow variable
	sorder := StandingOrder{
		Schedule: pschedule,
		Details:  pdetails,
	}

	// upsert StandingOrder as ACTIVE
	_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "ACTIVE-SORDER")

	// Define query handlers for variables
	//
	// PaymentOrigin query handler
	QueryPaymentOrigin := "payment.origin"
	err := workflow.SetQueryHandler(ctx, QueryPaymentOrigin, func() (string, error) {
		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorCyan, "Received Query - QueryPaymentOrigin:",
			sorder.Details.SourceAccount, u.ColorReset)
		return sorder.Details.SourceAccount, nil
	})
	if err != nil {
		logger.Info("Workflow: SetQueryHandler: QueryPaymentOrigin handler failed.", "Error", err)
		return "Error", err
	}

	// PaymentDestination query handler
	QueryPaymentDestination := "payment.destination"
	err = workflow.SetQueryHandler(ctx, QueryPaymentDestination, func() (string, error) {
		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorCyan, "Received Query - QueryPaymentDestination:",
			sorder.Details.TargetAccount, u.ColorReset)
		return sorder.Details.TargetAccount, nil
	})
	if err != nil {
		logger.Info("Workflow: SetQueryHandler: QueryPaymentDestination handler failed.", "Error", err)
		return "Error", err
	}

	// PaymentAmount query handler
	QueryPaymentAmount := "payment.amount"
	err = workflow.SetQueryHandler(ctx, QueryPaymentAmount, func() (string, error) {
		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorCyan, "Received Query - QueryPaymentAmount:",
			sorder.Details.Amount, u.ColorReset)
		return fmt.Sprint(sorder.Details.Amount), nil
	})
	if err != nil {
		logger.Info("Workflow: SetQueryHandler: QueryPaymentAmount handler failed.", "Error", err)
		return "Error", err
	}

	// PaymentReference query handler
	QueryPaymentReference := "payment.reference"
	err = workflow.SetQueryHandler(ctx, QueryPaymentReference, func() (string, error) {
		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorCyan, "Received Query - QueryPaymentReference:",
			sorder.Details.ReferenceID, u.ColorReset)
		return sorder.Details.ReferenceID, nil
	})
	if err != nil {
		logger.Info("Workflow: SetQueryHandler: QueryPaymentReference handler failed.", "Error", err)
		return "Error", err
	}

	// SchedulePeriod query handler
	QuerySchedulePeriodDuration := "schedule.periodduration"
	err = workflow.SetQueryHandler(ctx, QuerySchedulePeriodDuration, func() (string, error) {
		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorCyan, "Received Query - QuerySchedulePeriodDuration:",
			sorder.Schedule.PeriodDuration, u.ColorReset)
		return fmt.Sprint(sorder.Schedule.PeriodDuration), nil
	})
	if err != nil {
		logger.Info("Workflow: SetQueryHandler: QuerySchedulePeriodDuration handler failed.", "Error", err)
		return "Error", err
	}

	// Define signals for payment amount, schedule period, cancel
	//

	selector := workflow.NewSelector(ctx)

	// payment amount signal
	amountCh := workflow.GetSignalChannel(ctx, "sorderamount")
	selector.AddReceive(amountCh, func(ch workflow.ReceiveChannel, _ bool) {
		// do this when signal received

		// read contents from signal into variable
		var amountSignal int
		ch.Receive(ctx, &amountSignal)

		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorYellow, "Received Signal - sorderamount:",
			amountSignal, u.ColorReset)

		// update workflow variable value
		sorder.Details.Amount = amountSignal
	})

	// payment reference signal
	referenceCh := workflow.GetSignalChannel(ctx, "sorderreference")
	selector.AddReceive(referenceCh, func(ch workflow.ReceiveChannel, _ bool) {
		// do this when signal received

		// read contents from signal into variable
		var referenceSignal string
		ch.Receive(ctx, &referenceSignal)

		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorYellow, "Received Signal - sorderreference:",
			referenceSignal, u.ColorReset)

		// update workflow variable value
		sorder.Details.ReferenceID = referenceSignal
	})

	// schedule period signal
	//  ?(should break out of loop, but then sleep for period before calling child workflow..)
	scheduleCh := workflow.GetSignalChannel(ctx, "sorderschedule")
	selector.AddReceive(scheduleCh, func(ch workflow.ReceiveChannel, _ bool) {
		// do this when signal received

		// read contents from signal into variable
		var scheduleSignal int
		ch.Receive(ctx, &scheduleSignal)

		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorYellow, "Received Signal - sorderschedule:",
			scheduleSignal, u.ColorReset)

		// update workflow variable value
		sorder.Schedule.PeriodDuration = time.Duration(scheduleSignal) * time.Second
	})

	// cancel subscription signal
	cancelCh := workflow.GetSignalChannel(ctx, "cancelsorder")
	selector.AddReceive(cancelCh, func(ch workflow.ReceiveChannel, _ bool) {
		// do this when signal received

		// read contents from signal
		var cancelSOrderSignal bool
		ch.Receive(ctx, &cancelSOrderSignal)

		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorYellow, "Received Signal - cancelsorder:",
			cancelSOrderSignal, u.ColorReset)

		// update workflow variable value
		sorder.Schedule.Active = false
	})

	/* Main */
	amended := false
	for sorder.Schedule.Active {

		logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorReset, "Sleeping until next Scheduled Payment (", sorder.Schedule.PeriodDuration, ")..")

		// Sleep for time but interrupt if cancel signal comes in:
		workflow.AwaitWithTimeout(ctx, sorder.Schedule.PeriodDuration, selector.HasPending)

		// Check if cancel signal received during period (will interrupt Sleep)
		for selector.HasPending() {
			selector.Select(ctx)
			if sorder.Schedule.Active {
				// received non-cancel signal
				amended = true
			}
		}
		if !sorder.Schedule.Active {
			logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorReset, "Was Cancelled.")
			_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "CANCELLED-SORDER")

		} else if amended {
			logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorReset, "Standing Order has been amended:", sorder.Details)
			amended = false

		} else {
			// Standing Order still Active (/not Cancelled)
			logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorReset, "Performing scheduled payment:", sorder.Details)

			// Call Transfer workflow as a child workflow to implement the payment
			// note: child workflow can be fullfilled by different taskqueue / registered worker

			//thisid := GenRandString(5) // non-deterministic on worker restart and workflow queries!!!
			encodedRandom := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
				return rand.Intn(99999)
			})
			var thisid int
			encodedRandom.Get(&thisid)
			logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorReset, "ChildWorkflow:", fmt.Sprintf("go-txfr-sorder-payment-%d", thisid))

			cwo := workflow.ChildWorkflowOptions{
				WorkflowID: fmt.Sprintf("go-txfr-sorder-payment-%d", thisid),
				TaskQueue:  mt.MoneyTransferTaskQueueName,
				//TaskQueue: StandingOrdersTaskQueueName,
			}
			childCtx := workflow.WithChildOptions(ctx, cwo)

			var delay int = 5 // just to slow it down for demos
			var result string
			err := workflow.ExecuteChildWorkflow(childCtx, mt.TransferWorkflow, sorder.Details, delay).Get(childCtx, &result)

			if err != nil {
				logger.Error(u.ColorGreen, "S/O-Workflow:", u.ColorRed, "Child workflow Transfer failed!", u.ColorReset, err)
				// do some failed sorder Activity.. email notification etc..
				sorder.Schedule.Active = false
				_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "FAILED-SORDER")

				// ToDo: Call a sorder notification activity here..
				logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorRed, "Scheduled payment:", sorder.Details, "Completed with Error,", u.ColorReset, err)

			} else {
				// Increment the payment counter
				transferCounter++

				// This Transfer is complete, no more work to do this period
				logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorReset, "Scheduled payment[", transferCounter, "]:", sorder.Details, "Completed with result:", result)
			}
			if transferCounter == maxSOrdersPerWF {

		    logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorYellow, "Workflow count reached[", transferCounter, "], Continuing as New..")

				transferCounter = 0
				_ = u.UpsertSearchAttribute(ctx, "CustomStringField", "CASN-SORDER")

				return "Workflow Continued as New", workflow.NewContinueAsNewError(ctx, StandingOrderWorkflow, pdetails, pschedule)
			}
		}
	}

	logger.Info(u.ColorGreen, "S/O-Workflow:", u.ColorReset, "Complete", "-", workflow.GetInfo(ctx).WorkflowExecution.ID)
	return "Workflow Completed.", nil
}
