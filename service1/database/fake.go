package database

import (
	"context"
	"errors"
)

// FakeDB .
type FakeDB struct{}

// IncrementBy .
func (*FakeDB) IncrementBy(context.Context, string, int64) (int64, error) {
	return 0, errors.New("some error ocured")
}
