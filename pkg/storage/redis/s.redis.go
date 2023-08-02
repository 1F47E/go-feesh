package storage_redis

import (
	"context"
	"encoding/json"

	"github.com/1F47E/go-feesh/pkg/entity/models/tx"

	redis "github.com/redis/go-redis/v9"
)

type Redis struct {
	ctx context.Context
	db  *redis.Client
}

func New(ctx context.Context) (*Redis, error) {
	r := Redis{
		ctx: ctx,
		db: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
	}
	// check connection
	err := r.db.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *Redis) TxGet(txid string) (*tx.Tx, error) {
	val, err := r.db.Get(r.ctx, txid).Result()
	if err != nil {
		return nil, err
	}
	var tx tx.Tx
	err = json.Unmarshal([]byte(val), &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *Redis) TxAdd(tx *tx.Tx) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return r.db.Set(r.ctx, tx.Hash, data, 0).Err()
}
