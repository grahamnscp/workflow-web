package accountonboarding

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"path"
	"text/template"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	mail "github.com/xhit/go-simple-mail/v2"

	u "webapp/utils"
)

// Credit mock APIs
func CheckBlacklistedAPI(applicant string) (bool, error) {

	log.Println("CheckBlacklistedAPI: Called -", applicant)

	mDBClient := u.MongoDBGetConnection()
	defer mDBClient.Disconnect(context.Background())
	db := mDBClient.Database(u.MongoDBName).Collection("creditblacklist")

	findOptions := options.Find()
	findOptions.SetLimit(10)

	cur, err := db.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Printf("GetFraudApproverAPI:%s Failed to access database, err: %v%s", u.ColorRed, err, u.ColorReset)
		return false, err
	}

	found := false

	for cur.Next(context.TODO()) {
		var elem Blacklist
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		if elem.Email == applicant {
			found = true
		}
	}
	log.Println("CheckBlacklistedAPI: Applicant's email was found:", found)

	return found, nil
}

// Fraud mock APIs
func GetFraudRiskAPI(applicant string) (int, error) {

	log.Println("GetFraudListAPI: Called -", applicant)

	mDBClient := u.MongoDBGetConnection()
	defer mDBClient.Disconnect(context.Background())
	db := mDBClient.Database(u.MongoDBName).Collection("fraudrisk")

	findOptions := options.Find()
	findOptions.SetLimit(10)

	cur, err := db.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Printf("GetFraudApproverAPI:%s Failed to access database, err: %v%s", u.ColorRed, err, u.ColorReset)
		return 0, err
	}

	// low risk unless logged as otherwise
	risk := 1

	for cur.Next(context.TODO()) {
		var elem FraudRisk
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		if elem.Email == applicant {
			risk = elem.Risk
		}
	}
	log.Println("GetFraudListAPI: Applicant's fraud risk is:", risk)

	return risk, nil
}

func GetFraudApproverAPI() (string, error) {

	log.Println("GetFraudApproverAPI: Called")

	mDBClient := u.MongoDBGetConnection()
	defer mDBClient.Disconnect(context.Background())
	db := mDBClient.Database(u.MongoDBName).Collection("bankusers")

	findOptions := options.Find()
	findOptions.SetLimit(10)

	cur, err := db.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Printf("GetFraudApproverAPI:%s Failed to access database, err: %v%s", u.ColorRed, err, u.ColorReset)

		return "", err
	}

	var approver string = "admin@examplebank.co"
	var level int = 0

	for cur.Next(context.TODO()) {
		var elem BankUser
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		if elem.ApprovalLevel > 1 {
			approver = elem.Email
			level = elem.ApprovalLevel
			break
		}
	}
	log.Println("GetFraudApproverAPI: Approver Email:", approver, ", Approval Level:", level)

	return approver, nil
}

// Token Functions mock APIs
func TokenGenerateAPI(wfid string, runid string) string {
	//log.Println("TokenGenerateAPI: Called for WFID:", wfid, ", RUNID:", runid)

	rawToken := path.Join(wfid, runid)
	token := base64.URLEncoding.EncodeToString([]byte(rawToken))

	//log.Println("TokenGenerateAPI: Generated Token:", token)
	return token
}

func TokenDecodeAPI(token string) (string, string, error) {
	//log.Println("TokenDecodeAPI: Called with Token:", token)

	var rawToken []byte

	rawToken, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", "", err
	}
	wfid := path.Dir(string(rawToken))
	runid := path.Base(string(rawToken))

	log.Println("TokenDecodeAPI: Decoded WFID:", wfid, ", RUNID:", runid)
	return wfid, runid, nil
}

// Provision Account API
func ProvisionAccountAPI(accountName string, email string) (string, int, error) {

	log.Println("ProvisionAccountAPI: Called for", accountName)

	if !checkBankService() {
		return accountName, 0, &BankIntermittentError{}
	}

	// Get database connection
	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	// if already exists add number to end of name
	sqlStatement := fmt.Sprintf("select account_name from dataentry.accounts where account_name='%s'", accountName)
	rows, dberr := dbc.Query(sqlStatement)
	if dberr != nil {
		log.Println("ProvisionAccountAPI: db check query failed, dberr:", dberr)

	} else {
		defer rows.Close()

		checkaccname := "new"
		for rows.Next() {
			rows.Scan(&checkaccname)
		}
		if checkaccname == "new" {
			// account entry not found which is good
		} else {
			// account entry already exists, increment name
			randChars := make([]byte, 2)
			for i := range randChars {
				allowedChars := "0123456789"
				randChars[i] = allowedChars[rand.Intn(len(allowedChars))]
			}
			accountName = fmt.Sprintf("%s%s", accountName, string(randChars))

			log.Println("ProvisionAccountAPI: Account name set to:", accountName)
		}
	}

	// Insert new account
	sqlStatement = fmt.Sprintf("INSERT INTO dataentry.accounts (account_number,account_name,account_balance,email) SELECT max(account_number)+1, '%s', %d, '%s' FROM dataentry.accounts",
		accountName, NewAccountBonusBalance, email)

	stmtIns, dberr := dbc.Prepare(sqlStatement)
	if dberr != nil {
		log.Println("ProvisionAccountAPI: account insert Prepare failed! ", dberr)
		return accountName, 0, dberr
	}
	_, dberr = stmtIns.Exec()
	if dberr != nil {
		log.Println("ProvisionAccountAPI: account insert Exec failed! ", dberr)
		return accountName, 0, dberr
	}
	log.Println("ProvisionAccountAPI: New account added to database.")

	// Read new account number
	sqlStatement = fmt.Sprintf("select account_number from dataentry.accounts where account_name='%s'", accountName)
	rows, dberr = dbc.Query(sqlStatement)
	if dberr != nil {
		if dberr == sql.ErrNoRows {
			log.Println("ProvisionAccountAPI: account entry not found")
		} else {
			log.Println("ProvisionAccountAPI: db query failed, dberr:", dberr)
		}
	}
	defer rows.Close()

	// read entry
	var accountNum int
	for rows.Next() {
		rows.Scan(&accountNum)
	}

	log.Printf("ProvisionAccountAPI: Account for %s has AccountNum: %d",
		accountName, accountNum)

	return accountName, accountNum, nil
}

// Email notification API
func SendEmailNotificationAPI(processStage int, af ApplicationForm) error {

	log.Println("SendEmailNotificationAPI: Called")

	var emailTemplate, emailSubject string

	switch processStage {
	case EmailNotificationStageReceived:
		emailTemplate = fmt.Sprintf("templates/%s", EmailNotificationStageReceivedTemplate)
		emailSubject = EmailNotificationStageReceivedSubject

	case EmailNotificationStageInProgress:
		emailTemplate = fmt.Sprintf("templates/%s", EmailNotificationStageInProgressTemplate)
		emailSubject = EmailNotificationStageInProgressSubject

	case EmailNotificationStageComplete:
		emailTemplate = fmt.Sprintf("templates/%s", EmailNotificationStageCompleteTemplate)
		emailSubject = EmailNotificationStageCompleteSubject

	case EmailNotificationStageRejected:
		emailTemplate = fmt.Sprintf("templates/%s", EmailNotificationStageRejectedTemplate)
		emailSubject = EmailNotificationStageRejectedSubject

	case EmailNotificationStageCancelled:
		emailTemplate = fmt.Sprintf("templates/%s", EmailNotificationStageCancelledTemplate)
		emailSubject = EmailNotificationStageCancelledSubject
	}
	log.Printf("SendEmailNotificationAPI: Sending to email: %s, subject: %s", af.Email, emailSubject)

	// Generate the content
	htmlContentTemplate, err := template.ParseFiles(emailTemplate)
	if err != nil {
		log.Printf("SendEmailNotificationAPI:%s Failed to Parse template file,%s %v", u.ColorRed, u.ColorReset, err)
		return err
	}

	// local stream variable for template content
	var htmlContent bytes.Buffer

	// Execute the html template with the ApplicationForm content
	err = htmlContentTemplate.Execute(&htmlContent, af)
	if err != nil {
		log.Printf("SendEmailNotificationAPI:%s Failed to Execute template,%s %v", u.ColorRed, u.ColorReset, err)
		return err
	}

	err = SendEmailAPI(af.Email, emailFromAddress, emailSubject, htmlContent)
	if err != nil {
		log.Printf("SendEmailNotificationAPI:%s SendEmailAPI Failed,%s %v", u.ColorRed, u.ColorReset, err)
		return err
	}

	return nil
}

// Send Approval Email API
func SendEmailApprovalAPI(emailToAddress string, af ApplicationForm) error {

	log.Println("SendEmailApprovalAPI: Called")

	emailTemplate := fmt.Sprintf("templates/%s", EmailFraudApprovalTemplate)
	emailSubject := "Account Application Approval Required"

	log.Printf("SendEmailApprovalAPI: Sending to email: %s, subject: %s", emailToAddress, emailSubject)

	// Generate the content
	htmlContentTemplate, err := template.ParseFiles(emailTemplate)
	if err != nil {
		log.Printf("SendEmailApprovalAPI:%s Failed to Parse template file,%s %v", u.ColorRed, u.ColorReset, err)
		return err
	}

	var htmlContent bytes.Buffer
	err = htmlContentTemplate.Execute(&htmlContent, af)
	if err != nil {
		log.Printf("SendEmailApprovalAPI:%s Failed to Execute template,%s %v", u.ColorRed, u.ColorReset, err)
		return err
	}

	err = SendEmailAPI(emailToAddress, emailFromAddress, emailSubject, htmlContent)
	if err != nil {
		log.Printf("SendEmailApprovalAPI:%s SendEmailAPI Failed,%s %v", u.ColorRed, u.ColorReset, err)
		return err
	}

	return nil
}

// Send email backend
func SendEmailAPI(toAddress string, fromAddress string, emailSubject string, htmlContent bytes.Buffer) error {

	// create email
	email := mail.NewMSG()
	email.SetFrom(fromAddress).
		AddTo(toAddress).
		SetSubject(emailSubject).
		SetBody(mail.TextHTML, htmlContent.String())

	if email.Error != nil {
		log.Printf("SendEmailAPI:%s Failed set email form,%s %v", u.ColorRed, u.ColorReset, email.Error)
		return email.Error
	}

	// connect to mail server
	server := mail.NewSMTPClient()
	server.Host = SMTPHost
	server.Port = SMTPPort
	server.ConnectTimeout = time.Second
	server.SendTimeout = time.Second

	client, err := server.Connect()
	if err != nil {
		log.Printf("SendEmailAPI:%s Failed connect to SMTP server,%s %v", u.ColorRed, u.ColorReset, err)
		return err
	}

	// send email using client
	err = email.Send(client)
	if err != nil {
		log.Printf("SendEmailAPI:%s Failed to send email,%s %v", u.ColorRed, u.ColorReset, err)
		return fmt.Errorf("FAILED TO SEND EMAIL, ERR: %v", err)
	}

	return nil
}

// Bank backend status check
type BankIntermittentError struct{}

func (m *BankIntermittentError) Error() string {
	return "Banking Service currently unavailable"
}

func checkBankService() bool {

	var bankAPIStatus int = 200

	// Get database connection
	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	sqlStatement := `SELECT up FROM dataentry.bankapistatus`
	rows, dberr := dbc.Query(sqlStatement)
	if dberr != nil {
		if dberr == sql.ErrNoRows {
			log.Println("checkBankService: status table has no rows")
		} else {
			log.Fatal(dberr)
		}
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&bankAPIStatus)
	}

	if bankAPIStatus != 1 {
		log.Printf("%scheckBankService: Bank service API is DOWN (status: %d)%s", u.ColorCyan, bankAPIStatus, u.ColorReset)
		return bool(false)
	}
	return bool(true)
}
