package types

// MongoEntity is an interface that serves as a constraint for our generic implemenation of a MongoDB repository.
// The type passed to our mongo repository should have an ID field and a version field.
type MongoEntity[K, T any] interface {
	GetID() K
	GetVersion() int32
	SetVersion(version int32) T
}
