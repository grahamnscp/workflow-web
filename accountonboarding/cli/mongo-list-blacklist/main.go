package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	ao "webapp/accountonboarding"
	u "webapp/utils"
)

type Blacklist struct {
	Email string
}

func main() {

	fmt.Println("main: mongodb test..")

	mDBClient := u.MongoDBGetConnection()
	defer mDBClient.Disconnect(context.Background())

	// read collection
	db := mDBClient.Database(u.MongoDBName).Collection("creditblacklist")

	findOptions := options.Find()
	findOptions.SetLimit(10)

	//Define an array in which you can store the decoded documents
	var results []Blacklist

	applicant := "sal@aol.us"
	found := false

	//Passing the bson.D{{}} as the filter matches  documents in the collection
	cur, err := db.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}

	//Finding multiple documents returns a cursor
	//Iterate through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var elem Blacklist
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		if elem.Email == applicant {
			fmt.Println(applicant, "is in blacklist")
			found = true
		}
		results = append(results, elem)
	}

	fmt.Println("Blacklist:", results)
	fmt.Println("Applicant's email was found:", found)

	// Check using function
	inBL, _ := ao.CheckBlacklistedAPI(applicant)
	if inBL {
		fmt.Println("function ao.CheckBlacklisted found applicant's email,", applicant, "in blacklist")
	}
}
