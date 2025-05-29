package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Collection *mongo.Collection

func Connect() {
	uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	Collection = client.Database("timestampdb").Collection("hits")
}

func GetPunchesByMonth(year int, month time.Month) ([]time.Time, error) {
	loc := time.Now().Location()
	start := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	}

	cur, err := Collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	var punches []time.Time
	for cur.Next(context.Background()) {
		var doc struct {
			Timestamp time.Time `bson:"timestamp"`
		}
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		punches = append(punches, doc.Timestamp)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return punches, nil
}
