package database

import (
	"context"
	"log"
	"time"

	"github.com/PlayEconomy37/Play.Common/configuration"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

// NewMongoClient creates a new mongo client with given configuration
func NewMongoClient(cfg *configuration.Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// MongoDB connection options
	maxOpenConns := uint64(cfg.DB.MaxOpenConns)
	maxIdleTime := time.Duration(cfg.DB.MaxIdleTimeMS)
	opts := options.Client()
	opts.Monitor = otelmongo.NewMonitor() // Opentelemetry tracing
	opts.MaxPoolSize = &maxOpenConns
	opts.MaxConnIdleTime = &maxIdleTime
	opts.ApplyURI(cfg.DB.Dsn)

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
