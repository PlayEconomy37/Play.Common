package mongo

import (
	"context"
	"errors"

	"github.com/PlayEconomy37/Play.Common/filters"
	"github.com/PlayEconomy37/Play.Common/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Very specific value that is difficult to replicate
const DEFAULT_PRICE = 51.43243344285539

// We'll return this error when trying to do operations on an item
// that doesn't exist in our database
var ErrRecordNotFound = errors.New("record not found")

type MongoRepository[T types.Entity] struct {
	collection *mongo.Collection
}

// Creates a new mongoDB repository
func NewRepository[T types.Entity](client *mongo.Client, database, collection string) types.Repository[T] {
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
func (repo MongoRepository[T]) GetAll(ctx context.Context, name string, minPrice float64, maxPrice float64, filters filters.Filters) ([]T, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Find options
	findOptions := options.Find()
	findOptions.SetSkip(filters.Offset())
	findOptions.SetLimit(filters.Limit())
	findOptions.SetSort(bson.M{filters.SortColumn(): filters.SortDirection()})
	findOptions.SetSort(bson.M{"_id": 1}) // We include a secondary sort on the id to ensure a consistent ordering

	// Set filters
	filter := bson.M{}

	if name != "" {
		filter["$text"] = bson.M{"$search": name}
	}

	if minPrice != DEFAULT_PRICE {
		filter["price"] = bson.M{"$gte": minPrice}
	}

	if maxPrice != DEFAULT_PRICE {
		filter["price"] = bson.M{"$gte": maxPrice}
	}

	var items []T

	cursor, err := repo.collection.Find(ctx, filter, findOptions)
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
func (repo MongoRepository[T]) Update(ctx context.Context, entity T) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := repo.collection.UpdateOne(ctx, bson.M{"_id": entity.ID(), "version": entity.Version()}, bson.M{"$set": entity})
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
