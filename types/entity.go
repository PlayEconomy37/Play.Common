package types

import "go.mongodb.org/mongo-driver/bson/primitive"

// Constraints for our generic implemenation of a mongoDB repository.
// The type passed to our mongo repository should have an ID field and a version field.
type Entity[T any] interface {
	GetID() primitive.ObjectID
	GetVersion() int32
	SetVersion(version int32) T
}
