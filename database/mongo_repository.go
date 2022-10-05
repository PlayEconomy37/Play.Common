package database

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

var (
	// ErrRecordNotFound is returned when trying to fetch an item
	// that doesn't exist in our database
	ErrRecordNotFound = errors.New("record not found")

	// ErrEditConflict is returned when trying to update an item
	// in which the document version does not match
	// (or the record has been deleted).
	ErrEditConflict = errors.New("edit conflict")
)

// MongoRepository is a generic MongoDB repository struct
type MongoRepository[K any, T types.MongoEntity[K, T]] struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository
func NewMongoRepository[K any, T types.MongoEntity[K, T]](client *mongo.Client, database, collection string) types.MongoRepository[K, T] {
	return &MongoRepository[K, T]{
		collection: client.Database(database).Collection(collection),
	}
}

// GetByID retrieves a specific document from the collection by its id
func (repo MongoRepository[K, T]) GetByID(ctx context.Context, id K) (T, error) {
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

// GetByFilter retrieves a specific document from the collection by the given filter
func (repo MongoRepository[K, T]) GetByFilter(ctx context.Context, filter primitive.M) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var item T

	err := repo.collection.
		FindOne(ctx, filter).
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

// GetAll retrieves all documents from the collection
func (repo MongoRepository[K, T]) GetAll(
	ctx context.Context,
	filter primitive.M,
	findOpts filters.Filters,
) ([]T, filters.Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Find options
	findOptions := options.Find()
	findOptions.SetSkip(int64(findOpts.Offset()))
	findOptions.SetLimit(int64(findOpts.Limit()))
	findOptions.SetSort(
		bson.D{
			{Key: findOpts.SortColumn(), Value: findOpts.SortDirectionMongo()},
			{Key: "_id", Value: 1},
		},
	) // We include a secondary sort on the id to ensure a consistent ordering

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
	metadata := filters.CalculateMetadata(int(count), findOpts.Page, findOpts.PageSize)

	return items, metadata, nil
}

// Create inserts a new document in the collection
func (repo MongoRepository[K, T]) Create(ctx context.Context, MongoEntity T) (*K, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := repo.collection.InsertOne(ctx, MongoEntity)
	if err != nil {
		return nil, err
	}

	id, ok := (result.InsertedID).(K)
	if !ok {
		return nil, err
	}

	return &id, nil
}

// Update updates a specific document from the collection
func (repo MongoRepository[K, T]) Update(ctx context.Context, MongoEntity T) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := repo.collection.UpdateOne(
		ctx,
		bson.M{"_id": MongoEntity.GetID(), "version": MongoEntity.GetVersion()},
		bson.M{"$set": MongoEntity.SetVersion(MongoEntity.GetVersion() + 1)},
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

// Delete deletes a specific document from the collection
func (repo MongoRepository[K, T]) Delete(ctx context.Context, id K) error {
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
