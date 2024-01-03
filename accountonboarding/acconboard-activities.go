package accountonboarding

import (
	"context"
	"log"

	u "webapp/utils"
)

/* Activity - Credit Check Blacklist */
func CheckBlacklist(ctx context.Context, email string) (bool, error) {

	log.Printf("%sCreditChk-CheckBlacklist-Activity:%s for email: %s.%s\n",
		u.ColorGreen, u.ColorBlue, email, u.ColorReset)

	found := false
	found, err := CheckBlacklistedAPI(email)
	if err != nil {
		log.Printf("%sCreditChk-CheckBlacklist-Activity:%s API returned failure: %v%s\n",
			u.ColorGreen, u.ColorRed, err, u.ColorReset)
		return false, err
	}
	if found {
		log.Printf("%sCreditChk-CheckBlacklist-Activity:%s %s email is in Blacklist%s\n",
			u.ColorGreen, u.ColorBlue, email, u.ColorReset)
	}
	return found, nil
}

/* Activity - Fraud Check Risk Level */
func CheckFraudRisk(ctx context.Context, email string) (int, error) {

	log.Printf("%sFraudChk-CheckFraudRisk-Activity:%s for email: %s%s\n",
		u.ColorGreen, u.ColorBlue, email, u.ColorReset)

	risk, err := GetFraudRiskAPI(email)
	if err != nil {
		log.Printf("%sFraudChk-CheckFraudRisk-Activity:%s API returned failure: %v%s\n",
			u.ColorGreen, u.ColorRed, err, u.ColorReset)
		return 0, err
	}
	log.Printf("%sFraudChk-CheckFraudRisk-Activity:%s %s email Risk is: %d%s\n",
		u.ColorGreen, u.ColorBlue, email, risk, u.ColorReset)

	return risk, nil
}

/* Activity - Fraud Check Get Approver Email */
func GetFraudApprover(ctx context.Context) (string, error) {

	log.Printf("%sFraudChk-GetFraudApprover-Activity:%s Called%s\n",
		u.ColorGreen, u.ColorBlue, u.ColorReset)

	approver, err := GetFraudApproverAPI()
	if err != nil {
		log.Printf("%sFraudChk-GetFraudApprover-Activity:%s API returned failure: %v%s\n",
			u.ColorGreen, u.ColorRed, err, u.ColorReset)
		return "", err
	}
	log.Printf("%sFraudChk-GetFraudApprover-Activity:%s Approver email: %s%s\n",
		u.ColorGreen, u.ColorBlue, approver, u.ColorReset)

	return approver, nil
}

/* Activity - Generate Workflow Token */
func GenerateWorkflowToken(wkflid string, runid string) (string, error) {
	token := TokenGenerateAPI(wkflid, runid)

	return token, nil
}

/* Activity - Provision Account */
func ProvisionAccount(ctx context.Context, af ApplicationForm) (BankAccount, error) {

	log.Printf("%sProvisionAccount-Activity:%s Called for email: %s%s\n",
		u.ColorGreen, u.ColorBlue, af.Email, u.ColorReset)

	bankaccount := BankAccount{AccountName: af.FirstName, AccountNum: 0}

	accName, accNum, err := ProvisionAccountAPI(af.FirstName, af.Email)
	if err != nil {
		log.Printf("%sProvisionAccount-Activity:%s API returned failure: %v%s\n",
			u.ColorGreen, u.ColorRed, err, u.ColorReset)
		return bankaccount, err
	}

	log.Printf("%sProvisionAccount-Activity:%s %s's new Account Number is: %d%s\n",
		u.ColorGreen, u.ColorBlue, accName, accNum, u.ColorReset)

	bankaccount.AccountName = accName
	bankaccount.AccountNum = accNum
	return bankaccount, nil
}

/* Activity - Send Notification Email */
func SendNotificationEmail(ctx context.Context, appStage int, af ApplicationForm) (bool, error) {

	log.Printf("%sSendNotificationEmail-Activity:%s Application Notification Email to %s, stage %d%s\n",
		u.ColorGreen, u.ColorBlue, af.Email, appStage, u.ColorReset)

	err := SendEmailNotificationAPI(appStage, af)
	if err != nil {
		log.Printf("%sSendNotificationEmail-Activity:%s Failed to send Application Notification Email to %s, stage %d, err:%v%s\n",
			u.ColorGreen, u.ColorRed, af.Email, appStage, err, u.ColorReset)
		return false, err
	}
	return true, nil
}

/* Activity - Send Approval Email */
func SendApprovalEmail(ctx context.Context, as ApprovalStatus, af ApplicationForm) (bool, error) {

	log.Printf("%sSendApproverEmail-Activity:%s Application Approval Email to Approver: %s%s\n",
		u.ColorGreen, u.ColorBlue, as.Approver, u.ColorReset)

	err := SendEmailApprovalAPI(as.Approver, af)
	if err != nil {
		log.Printf("%sSendApprovalEmail-Activity:%s Failed to send Application Approval Email to %s, err:%v%s\n",
			u.ColorGreen, u.ColorRed, as.Approver, err, u.ColorReset)
		return false, err
	}
	return true, nil
}
