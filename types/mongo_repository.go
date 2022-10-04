package types

import (
	"context"

	"github.com/PlayEconomy37/Play.Common/filters"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoRepository is a generic MongoDB repository interface
type MongoRepository[K any, T MongoEntity[K, T]] interface {
	GetByID(ctx context.Context, id K) (T, error)
	GetByFilter(ctx context.Context, filter primitive.M) (T, error)
	GetAll(ctx context.Context, filter primitive.M, findOpts filters.Filters) ([]T, filters.Metadata, error)
	Create(ctx context.Context, entity T) (*K, error)
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id K) error
}
