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

var (
	// We'll return this error when trying to fetch an item
	// that doesn't exist in our database
	ErrRecordNotFound = errors.New("record not found")

	// We'll return this error when trying to update an item
	// in which the document version does not match
	// (or the record has been deleted).
	ErrEditConflict = errors.New("edit conflict")
)

type MongoRepository[T types.Entity[T]] struct {
	collection *mongo.Collection
}

// Creates a new mongoDB repository
func NewRepository[T types.Entity[T]](client *mongo.Client, database, collection string) types.Repository[T] {
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
func (repo MongoRepository[T]) GetAll(
	ctx context.Context,
	name string,
	minPrice float64,
	maxPrice float64,
	filteringOpts filters.Filters,
) ([]T, filters.Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Find options
	findOptions := options.Find()
	findOptions.SetSkip(int64(filteringOpts.Offset()))
	findOptions.SetLimit(int64(filteringOpts.Limit()))
	findOptions.SetSort(bson.M{filteringOpts.SortColumn(): filteringOpts.SortDirection()})
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
		filter["price"] = bson.M{"$lte": maxPrice}
	}

	var items []T

	cursor, err := repo.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return items, filters.Metadata{}, err
	}

	// Get total number of records that exist in database with given filters
	count, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return items, filters.Metadata{}, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var item T

		if err = cursor.Decode(&item); err != nil {
			return items, filters.Metadata{}, err
		}

		items = append(items, item)

	}

	if err := cursor.Err(); err != nil {
		return items, filters.Metadata{}, err
	}

	// Generate a Metadata struct, passing in the total document count and pagination
	// parameters from the client
	metadata := filters.CalculateMetadata(int(count), filteringOpts.Page, filteringOpts.PageSize)

	return items, metadata, nil
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

	result, err := repo.collection.UpdateOne(
		ctx,
		bson.M{"_id": entity.GetID(), "version": entity.GetVersion()},
		bson.M{"$set": entity.SetVersion(entity.GetVersion() + 1)},
	)
	if err != nil {
		return err
	}

	// No document with given id was found in the database
	if result.MatchedCount == 0 {
		return ErrEditConflict
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
