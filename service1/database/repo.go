package database

import "context"

// DB .
type DB interface {
	IncrementBy(context.Context, string, int64) (int64, error)
}
