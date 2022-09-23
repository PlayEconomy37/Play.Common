package types

import "go.mongodb.org/mongo-driver/bson/primitive"

// Constraint for our generic implemenation of a mongoDB repository.
// The type passed to our mongo repository should have an ID field and a version field.
type Entity interface {
	ID() primitive.ObjectID
	Version() int32
}
