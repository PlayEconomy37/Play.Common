package mongo

import (
	"context"

	"github.com/Play-Economy37/Play.Common/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

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
func (repo MongoRepository[T]) GetById(ctx context.Context, id string) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var item T

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return item, err
	}

	err = repo.collection.
		FindOne(ctx, bson.M{"_id": objectId}).
		Decode(&item)
	if err != nil {
		return item, err
	}

	return item, nil
}

// Retrieves a specific document from the collection with the specified filter
func (repo MongoRepository[T]) GetOneByFilter(ctx context.Context, filter primitive.M) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var item T

	err := repo.collection.
		FindOne(ctx, filter).
		Decode(&item)
	if err != nil {
		return item, err
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

	return items, nil
}

// Inserts a new document in the collection
func (repo MongoRepository[T]) Create(ctx context.Context, entity T) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err := repo.collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	return nil
}

// Updates a specific document from the collection
func (repo MongoRepository[T]) Update(ctx context.Context, id string, entity T) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = repo.collection.UpdateOne(ctx, bson.M{"_id": objectId}, bson.M{"$set": entity})

	if err != nil {
		return err
	}

	return nil
}

// Deletes a specific document from the collection
func (repo MongoRepository[T]) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = repo.collection.DeleteOne(ctx, bson.M{"_id": objectId})

	if err != nil {
		return err
	}

	return nil
}
