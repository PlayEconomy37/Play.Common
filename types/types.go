package types

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Map that wraps whatever data we want to send back to the client
type Envelope map[string]any

// Generic repository interface
type Repository[T any] interface {
	GetById(ctx context.Context, id string) (T, error)
	GetOneByFilter(ctx context.Context, filter primitive.M) (T, error)
	GetAll(ctx context.Context) ([]T, error)
	GetAllByFilter(ctx context.Context, filter primitive.M) ([]T, error)
	Create(ctx context.Context, entity T) error
	Update(ctx context.Context, id string, entity T) error
	Delete(ctx context.Context, id string) error
}
