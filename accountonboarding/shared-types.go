package accountonboarding

import (
	"os"
)

var AccOnboardingTaskQueueName = os.Getenv("ACC_ONBOARDING_TASK_QUEUE")
var AccOnboardingSleep int = 120
var CreditCheckDelay int = 15
var FraudCheckDelay int = 25
var FraudCheckSleep int = 600
var NewAccountBonusBalance int = 50

type ApplicationForm struct {
	FirstName   string
	Surname     string
	Email       string
	Account     int
	AccountName string
	URL         string
}

type ApplicationStatus struct {
	Status    string
	Credit    string
	Fraud     string
	L1Appr    string
	L2Appr    string
	Approved  bool
	Denied    bool
	Cancelled bool
}

type ApplicationWFDetails struct {
	WFID  string
	RunID string
	Token string
}

type ApprovalStatus struct {
	Approver          string
	ApproverResponded bool
	Approved          bool
	Reason            string
	FraudRisk         int
}

type BankAccount struct {
	AccountName string
	AccountNum  int
}

// Pseudo process state email notification stages
var EmailNotificationStageReceived int = 1
var EmailNotificationStageInProgress int = 2
var EmailNotificationStageComplete int = 3
var EmailNotificationStageRejected int = 4
var EmailNotificationStageCancelled int = 5

var EmailNotificationStageReceivedTemplate string = "ApplicationReceived.html"
var EmailNotificationStageInProgressTemplate string = "ApplicationInProgress.html"
var EmailNotificationStageCompleteTemplate string = "ApplicationComplete.html"
var EmailNotificationStageRejectedTemplate string = "ApplicationRejected.html"
var EmailNotificationStageCancelledTemplate string = "ApplicationCancelled.html"
var EmailFraudApprovalTemplate string = "FraudApproval.html"

var EmailNotificationStageReceivedSubject string = "Application Received"
var EmailNotificationStageInProgressSubject string = "Application InProgress"
var EmailNotificationStageCompleteSubject string = "Application Completed"
var EmailNotificationStageRejectedSubject string = "Application Rejected"
var EmailNotificationStageCancelledSubject string = "Application Cancelled"

var emailFromAddress string = "noreply@examplebank.co"

var urlBankUserHome string = "http://localhost:8085/bankuserhome"
var urlBankFraudApproval string = "http://localhost:8085/accappfraudreview"

// localhost mailserver (mailhog alias to localhost)
var SMTPHost string = "localhost"
var SMTPPort int = 1025

// mongodb collections
// - creditblacklist
type Blacklist struct {
	Email string
}

// - bankusers
type BankUser struct {
	Username      string
	Email         string
	ApprovalLevel int
}

// - fraudrisk
type FraudRisk struct {
	Email string
	Risk  int
}

