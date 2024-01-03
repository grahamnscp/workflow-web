package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	u "webapp/utils"
)

type Fraudrisk struct {
	Email string
	Risk  int
}

func main() {

	fmt.Println("main: mongodb test..")

	mDBClient := u.MongoDBGetConnection()
	defer mDBClient.Disconnect(context.Background())

	// read collection
	db := mDBClient.Database(u.MongoDBName).Collection("fraudrisk")

	findOptions := options.Find()
	findOptions.SetLimit(10)

	//Define an array in which you can store the decoded documents
	var results []Fraudrisk

	applicant := "casinoharry@lv.crew"
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
		var elem Fraudrisk
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		if elem.Email == applicant {
			fmt.Println(applicant, "is in fraud risk list, risk rating:", elem.Risk)
			found = true
		}
		results = append(results, elem)
	}

	fmt.Println("Fraudrisk:", results)
	fmt.Println("Applicant's email was found:", found)

}
