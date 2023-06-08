package moneytransfer

import (
	"log"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"webapp/utils"
)

// DB: account_id account_number account_name account_balance datestamp
type DbAccount struct {
	AccountId      int
	AccountNumber  int
	AccountName    string
	AccountBalance float64
}

type DbAccounts []DbAccount

func ReadDbAccounts() DbAccounts {

	log.Println("ReadAccounts: called")

	// Get database connection
	dbc, _ := utils.GetDBConnection()
	defer dbc.Close()

	sqlStatement := `SELECT account_id, account_number, account_name, account_balance FROM dataentry.accounts`
	rows, dberr := dbc.Query(sqlStatement)
	if dberr != nil {
		if dberr == sql.ErrNoRows {
			log.Println("ReadAccounts: no account entres found")
		} else {
			log.Fatal(dberr)
		}
	}
	defer rows.Close()

	acc := DbAccount{}
	dbaccounts := []DbAccount{}

	for rows.Next() {
		rows.Scan(&acc.AccountId, &acc.AccountNumber, &acc.AccountName, &acc.AccountBalance)
		dbaccounts = append(dbaccounts, acc)
	}
	return dbaccounts
}
