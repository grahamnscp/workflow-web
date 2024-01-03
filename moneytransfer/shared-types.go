package moneytransfer

import (
  "os"
)

var MoneyTransferTaskQueueName = os.Getenv("TRANSFER_MONEY_TASK_QUEUE")
var DelayTimerBetweenWithdrawDeposit = os.Getenv("DELAY_TIMER_BETWEEN_WITHDRAW_DEPOSIT")

type PaymentDetails struct {
  SourceAccount string
  TargetAccount string
  ReferenceID   string
  Amount        int
}

type WorkflowInfo struct {
  Id         int
  WorkflowID string
  RunID      string
  TaskQueue  string
  Info       string
  Status     string
}

