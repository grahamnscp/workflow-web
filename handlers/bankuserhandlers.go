package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	u "webapp/utils"
)

// DB: account_id account_number account_name account_balance email datestamp
type UserAccount struct {
	AccountId      int
	AccountNumber  int
	AccountName    string
	AccountBalance float64
	Email          string
}

/* BankUserHome */
func BankUserHome(w http.ResponseWriter, r *http.Request) {

	log.Println("BankUserHome: called, method:", r.Method)

	// URL Parameters
	var account string
	params := r.URL.Query()
	for k, v := range params {
		if k == "ref" {
			account = strings.Join(v, "")
		}
		log.Println("BankUserHome: url params:", k, " => ", v)
	}

	log.Println("BankUserHome: Bank User Home for:", account)

	// Get database connection
	dbc, _ := u.GetDBConnection()
	defer dbc.Close()

	sqlStatement := `SELECT account_id, account_number, account_name, account_balance, email FROM dataentry.accounts`
	rows, dberr := dbc.Query(sqlStatement)
	if dberr != nil {
		if dberr == sql.ErrNoRows {
			log.Println("BankUserHome: no account entry found")
		} else {
			log.Fatal(dberr)
		}
	}
	defer rows.Close()

	acc := UserAccount{}

	for rows.Next() {
		rows.Scan(&acc.AccountId, &acc.AccountNumber, &acc.AccountName, &acc.AccountBalance, &acc.Email)
	}

	log.Println("BankUserHome: User Account:", acc)

	u.Render(w, "templates/BankUserHome.html", acc)
	return
}
