package types

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Generic repository interface
type Repository[T any] interface {
	GetById(ctx context.Context, id primitive.ObjectID) (T, error)
	GetAll(ctx context.Context) ([]T, error)
	GetAllByFilter(ctx context.Context, filter primitive.M) ([]T, error)
	Create(ctx context.Context, entity T) (primitive.ObjectID, error)
	Update(ctx context.Context, id primitive.ObjectID, entity T) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
