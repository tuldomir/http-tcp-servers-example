package database

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// // DB .
// type DB interface {
// 	IncrementBy(context.Context, string, int64) (int64, error)
// }

// RedisDB .
type RedisDB struct {
	Client *redis.Client
}

// NewDB .
func NewDB(client *redis.Client) *RedisDB {
	return &RedisDB{Client: client}
}

// Stop .
func (db *RedisDB) Stop() error {
	return db.Client.Close()
}

// IncrementBy .
func (db *RedisDB) IncrementBy(ctx context.Context, key string, val int64) (int64, error) {

	pipe := db.Client.TxPipeline()
	defer pipe.Close()

	incr := pipe.IncrBy(ctx, key, val)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), err
}
