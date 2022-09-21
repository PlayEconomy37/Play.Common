package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

const defaultTimeout = 3 * time.Second

func NewClient(dsn string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// MongoDB connection options
	opts := options.Client()
	opts.Monitor = otelmongo.NewMonitor() // Opentelemetry tracing
	opts.ApplyURI(dsn)

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Ping MongoDB to make sure it is up and running
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	return mongoClient, nil
}
