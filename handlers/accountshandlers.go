package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"net/http"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	u "webapp/utils"
)

// DB: account_id account_number account_name account_balance email datestamp
type Account struct {
	AccountId      int
	AccountNumber  int
	AccountName    string
	AccountBalance float64
	Email          string
}

type BankStatusType struct {
	Up string
}

/* Index Home */
func Home(w http.ResponseWriter, r *http.Request) {

	log.Println("Home: called")
	u.Render(w, "templates/Home.html", nil)
}

/* ListAccounts */
func ListAccounts(w http.ResponseWriter, r *http.Request) {

	log.Println("ListAccounts: called")

	// Get database connection
	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	sqlStatement := `SELECT account_id, account_number, account_name, account_balance FROM dataentry.accounts`
	rows, dberr := dbc.Query(sqlStatement)
	if dberr != nil {
		if dberr == sql.ErrNoRows {
			log.Println("ListAccounts: no account entres found")
		} else {
			log.Fatal(dberr)
		}
	}
	defer rows.Close()

	acc := Account{}
	accounts := []Account{}

	for rows.Next() {
		rows.Scan(&acc.AccountId, &acc.AccountNumber, &acc.AccountName, &acc.AccountBalance)
		accounts = append(accounts, acc)
	}

	u.Render(w, "templates/ListAccounts.html", accounts)
}

/* ShowAccount */
func ShowAccount(w http.ResponseWriter, r *http.Request) {

	log.Println("ShowAccount: called")

	// URL Parameters
	var name string
	params := r.URL.Query()
	for k, v := range params {
		if k == "name" {
			name = strings.Join(v, "")
		}
		log.Println("ShowAccount: url params:", k, " => ", v)
	}

	// Get database connection
	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	sqlStatement := fmt.Sprintf("SELECT account_id,account_number,account_name,account_balance,email FROM dataentry.accounts WHERE account_name='%s'", name)
	rows, dberr := dbc.Query(sqlStatement)
	if dberr != nil {
		if dberr == sql.ErrNoRows {
			log.Println("ShowAccount: no account entry found")
		} else {
			log.Fatal(dberr)
		}
	}
	defer rows.Close()

	// read entry
	acc := Account{}

	for rows.Next() {
		rows.Scan(&acc.AccountId, &acc.AccountNumber, &acc.AccountName, &acc.AccountBalance, &acc.Email)
	}

	// Display details for requested entry
	u.Render(w, "templates/ShowAccount.html", acc)
}

/* NewAccount */
func NewAccount(w http.ResponseWriter, r *http.Request) {

	log.Println("NewAccount: called")

	log.Println("NewAccount: method:", r.Method) //get request method
	if r.Method == "GET" {
		u.Render(w, "templates/NewAccount.html", nil)
		return
	}

	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	accNum, _ := strconv.Atoi(r.FormValue("accountnumber"))
	accBal, _ := strconv.ParseFloat(strings.TrimSpace(r.FormValue("accountbalance")), 64)
	newacc := Account{
		AccountNumber:  accNum,
		AccountName:    r.FormValue("accountname"),
		AccountBalance: accBal,
		Email:          r.FormValue("accountemail"),
	}
	log.Println("NewAccount: New Account Submitted:", newacc)

	// Get database connection
	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	sqlStatement := fmt.Sprintf("INSERT INTO dataentry.accounts (account_number, account_name, account_balance, email) VALUES (%d,'%s',%f,'%s')",
		newacc.AccountNumber, newacc.AccountName, newacc.AccountBalance, newacc.Email)
	stmtIns, dberr := dbc.Prepare(sqlStatement)
	if dberr != nil {
		log.Fatal("NewAccount: account insert Prepare failed! ", dberr)
	}
	_, dberr = stmtIns.Exec()
	if dberr != nil {
		log.Fatal("NewAccount: account insert Exec failed! ", dberr)
	}
	log.Println("NewAccount: New account added to database.")

	// Render acknowledgement page
	u.Render(w, "templates/NewAccount.html", struct{ Success bool }{true})
}

/* DeleteAccount */
func DeleteAccount(w http.ResponseWriter, r *http.Request) {

	log.Println("DeleteAccount: called")

	// URL Parameters
	var name string
	params := r.URL.Query()
	for k, v := range params {
		if k == "name" {
			name = strings.Join(v, "")
		}
		log.Println("DeleteAccount: Received URL Params:", k, " => ", v)
	}

	// Get database connection
	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	type delstatstr struct {
		Success bool
	}
	delstat := delstatstr{true}

	sqlStatement := fmt.Sprintf("DELETE FROM dataentry.accounts WHERE account_name = '%s'", name)
	stmtDel, dberr := dbc.Prepare(sqlStatement)
	if dberr != nil {
		log.Fatal("DeleteAccount: delete account Prepare failed! ", dberr)
		delstat.Success = false
	}
	result, dberr := stmtDel.Exec()
	if dberr != nil {
		log.Fatal("DeleteAccount: delete account Exec failed! ", dberr)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		delstat.Success = false
	}
	log.Println("Delete account:", name, "Rows affected:", rowsAffected, "deletestat:", delstat.Success)

	// Display deleteaccount confirmation
	u.Render(w, "templates/DeleteAccount.html", delstat)
}

/* BankStatus */
func BankStatus(w http.ResponseWriter, r *http.Request) {

	log.Println("BankStatus: called")

	// Get database connection
	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	sqlStatement := "SELECT up FROM dataentry.bankapistatus"
	rows, dberr := dbc.Query(sqlStatement)
	if dberr != nil {
		if dberr == sql.ErrNoRows {
			log.Println("BankStatus: no db entry found")
		} else {
			log.Fatal(dberr)
		}
	}
	defer rows.Close()

	// read entry
	var bankUPStatus int
	bankStatus := BankStatusType{Up: "true"}

	for rows.Next() {
		rows.Scan(&bankUPStatus)
	}
	log.Println("BankUPStatus:", bankUPStatus)

	if bankUPStatus != 1 {
		bankStatus.Up = "down"
	}

	// Display details
	u.Render(w, "templates/BankStatus.html", bankStatus)
}

/* OpenBank */
func OpenBank(w http.ResponseWriter, r *http.Request) {

	log.Println("OpenBank: called")

	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	sqlStatement := "UPDATE dataentry.bankapistatus SET up=1"

	stmtIns, dberr := dbc.Prepare(sqlStatement)
	if dberr != nil {
		log.Fatal("OpenBank: bank status update Prepare failed! ", dberr)
	}
	_, dberr = stmtIns.Exec()
	if dberr != nil {
		log.Fatal("OpenBank: bank status update Exec failed! ", dberr)
	}
	log.Println("OpenBank: Status Updated.")

	u.Render(w, "templates/OpenBank.html", struct{ Success bool }{true})
}

/* CloseBank */
func CloseBank(w http.ResponseWriter, r *http.Request) {

	log.Println("CloseBank: called")

	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	sqlStatement := "UPDATE dataentry.bankapistatus SET up=503"

	stmtIns, dberr := dbc.Prepare(sqlStatement)
	if dberr != nil {
		log.Fatal("CloseBank: bank status update Prepare failed! ", dberr)
	}
	_, dberr = stmtIns.Exec()
	if dberr != nil {
		log.Fatal("CloseBank: bank status update Exec failed! ", dberr)
	}
	log.Println("CloseBank: Status Updated.")

	u.Render(w, "templates/CloseBank.html", struct{ Success bool }{true})
}
