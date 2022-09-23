package types

import (
	"context"

	"github.com/PlayEconomy37/Play.Common/filters"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Generic repository interface
type Repository[T Entity] interface {
	GetById(ctx context.Context, id primitive.ObjectID) (T, error)
	GetAll(ctx context.Context, name string, minPrice float64, maxPrice float64, filteringOpts filters.Filters) ([]T, filters.Metadata, error)
	Create(ctx context.Context, entity T) (primitive.ObjectID, error)
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
