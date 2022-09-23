package mongo

import (
	"context"
	"errors"

	"github.com/PlayEconomy37/Play.Common/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// We'll return this error when trying to do operations on an item
// that doesn't exist in our database
var ErrRecordNotFound = errors.New("record not found")

type MongoRepository[T any] struct {
	collection *mongo.Collection
}

// Creates a new mongoDB repository
func NewRepository[T any](client *mongo.Client, database, collection string) types.Repository[T] {
	return &MongoRepository[T]{
		collection: client.Database(database).Collection(collection),
	}
}

// Retrieves a specific document from the collection by its id
func (repo MongoRepository[T]) GetById(ctx context.Context, id primitive.ObjectID) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var item T

	err := repo.collection.
		FindOne(ctx, bson.M{"_id": id}).
		Decode(&item)

	// If there was no matching item found, Decode() will return
	// a mongo.ErrNoDocuments error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return item, ErrRecordNotFound
		default:
			return item, err
		}
	}

	return item, nil
}

// Retrieves all documents from the collection
func (repo MongoRepository[T]) GetAll(ctx context.Context) ([]T, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var items []T

	cursor, err := repo.collection.Find(ctx, bson.M{})
	if err != nil {
		return items, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var item T

		if err = cursor.Decode(&item); err != nil {
			return items, err
		}

		items = append(items, item)

	}

	if err := cursor.Err(); err != nil {
		return items, err
	}

	return items, nil
}

// Retrieves all documents from the collection with the specified filter
func (repo MongoRepository[T]) GetAllByFilter(ctx context.Context, filter primitive.M) ([]T, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var items []T

	cursor, err := repo.collection.Find(ctx, filter)
	if err != nil {
		return items, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var item T

		if err = cursor.Decode(&item); err != nil {
			return items, err
		}

		items = append(items, item)

	}

	if err := cursor.Err(); err != nil {
		return items, err
	}

	return items, nil
}

// Inserts a new document in the collection
func (repo MongoRepository[T]) Create(ctx context.Context, entity T) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := repo.collection.InsertOne(ctx, entity)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return (result.InsertedID).(primitive.ObjectID), nil
}

// Updates a specific document from the collection
func (repo MongoRepository[T]) Update(ctx context.Context, id primitive.ObjectID, entity T) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := repo.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": entity})
	if err != nil {
		return err
	}

	// No document with given id was found in the database
	if result.MatchedCount == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Deletes a specific document from the collection
func (repo MongoRepository[T]) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := repo.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	// No document with given id was found in the database
	if result.DeletedCount == 0 {
		return ErrRecordNotFound
	}

	return nil
}
