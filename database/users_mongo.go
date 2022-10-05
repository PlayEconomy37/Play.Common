package database

import (
	"context"

	"github.com/PlayEconomy37/Play.Common/permissions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User is a struct that defines an user in our application
type User struct {
	ID          int64    `json:"id" bson:"_id"`
	Permissions []string `json:"permissions" bson:"permissions"`
	Activated   bool     `json:"activated" bson:"activated"`
	Version     int32    `json:"version" bson:"version"`
}

// GetID returns the id of an user.
// This method is necessary for our generic constraint of our mongo repository
// and for the user interface in our Authenticate and RequirePermission middlewares.
func (u User) GetID() int64 {
	return u.ID
}

// GetVersion returns the version of an user.
// This method is necessary for our generic constraint of our mongo repository.
func (u User) GetVersion() int32 {
	return u.Version
}

// SetVersion sets the version of an user to the given value and returns the user.
// This method is necessary for our generic constraint of our mongo repository.
func (u User) SetVersion(version int32) User {
	u.Version = version

	return u
}

// GetPermissions returns the permissions of an user.
// This method is necessary for the user interface in our Authenticate and RequirePermission middlewares.
func (u User) GetPermissions() permissions.Permissions {
	return u.Permissions
}

// CreateUsersCollection creates users collection in MongoDB database
func CreateUsersCollection(client *mongo.Client, databaseName string) error {
	db := client.Database(databaseName)

	// JSON validation schema
	jsonSchema := bson.M{
		"bsonType":             "object",
		"required":             []string{"permissions", "version"},
		"additionalProperties": false,
		"properties": bson.M{
			"_id": bson.M{
				"bsonType":    "long",
				"description": "User ID",
			},
			"permissions": bson.M{
				"bsonType":    "array",
				"description": "User permissions",
			},
			"activated": bson.M{
				"bsonType":    "bool",
				"description": "Flag to check if user is activated or not",
			},
			"version": bson.M{
				"bsonType":    "int",
				"description": "Document version",
			},
		},
	}

	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}

	// Create collection
	opts := options.CreateCollection().SetValidator(validator)
	err := db.CreateCollection(context.Background(), UsersCollection, opts)
	if err != nil {
		// Returns error if collection already exists so we ignore it
		return nil
	}

	return nil
}
