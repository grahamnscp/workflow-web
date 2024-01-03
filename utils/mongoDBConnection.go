package utils

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoDBGetConnection() *mongo.Client {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatal(err)
	}

	// quick ping test
	/*
		var result bson.M
		if err := client.Database(MongoDBName).RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
			log.Println("MongoDBGetConnection: Failed to ping mongo database, err:", err)
		}
	*/

	return client
}
