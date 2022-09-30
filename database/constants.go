package database

import "time"

// defaultTimeout is a constant that defines the default context timeout
// for database operations
const defaultTimeout = 3 * time.Second

// DefaultPrice is used as a default value when validating `min_price` and `max_price`.
// Very specific value that is difficult to replicate by client.
const DefaultPrice = 51.43243344285539
