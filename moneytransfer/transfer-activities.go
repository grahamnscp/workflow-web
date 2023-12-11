package moneytransfer

import (
	"context"
	"fmt"
	"log"

  u "webapp/utils"
)

/* Withdraw Activity */
func Withdraw(ctx context.Context, data PaymentDetails) (string, error) {

	log.Printf("%sTransfer-Workflow-Withdraw-Activity:%s Withdrawing $%d from account: %s.%s\n",
		u.ColorGreen, u.ColorBlue, data.Amount, data.SourceAccount, u.ColorReset)

	referenceID := fmt.Sprintf("%s-withdrawal", data.ReferenceID)
	bank := BankingService{"bank-api.example.com"}
	confirmation, err := bank.Withdraw(data.SourceAccount, data.Amount, referenceID)
	if err != nil {
		log.Printf("%sTransfer-Workflow-Withdraw-Activity:%s Failed to withdraw funds from account: %s, %v%s",
			u.ColorGreen, u.ColorRed, data.SourceAccount, err, u.ColorReset)
	}
	return confirmation, err
}

/* Deposit Activity */
func Deposit(ctx context.Context, data PaymentDetails) (string, error) {

	log.Printf("%sTransfer-Workflow-Deposit-Activity:%s Depositing $%d into account: %s.%s",
		u.ColorGreen, u.ColorBlue, data.Amount, data.TargetAccount, u.ColorReset)

	referenceID := fmt.Sprintf("%s-deposit", data.ReferenceID)
	bank := BankingService{"bank-api.example.com"}
	confirmation, err := bank.Deposit(data.TargetAccount, data.Amount, referenceID)
	if err != nil {
		log.Printf("%sTransfer-Workflow-Deposit-Activity:%s Failed to deposit funds to account: %s, %v.%s",
			u.ColorGreen, u.ColorRed, data.TargetAccount, err, u.ColorReset)
	}
	return confirmation, err
}

/* Refund Activity */
func Refund(ctx context.Context, data PaymentDetails) (string, error) {

	log.Printf("%sTransfer-Workflow-Refund-Activity:%s Refunding $%v back into account: %v.%s",
		u.ColorGreen, u.ColorBlue, data.Amount, data.SourceAccount, u.ColorReset)

	referenceID := fmt.Sprintf("%s-refund", data.ReferenceID)
	bank := BankingService{"bank-api.example.com"}
	confirmation, err := bank.Deposit(data.SourceAccount, data.Amount, referenceID)
	if err != nil {
		log.Printf("%sTransfer-Workflow-Refund-Activity:%s Failed to refund funds to account: %s, %v.%s",
			u.ColorGreen, u.ColorRed, data.SourceAccount, err, u.ColorReset)
	}
	return confirmation, err
}
