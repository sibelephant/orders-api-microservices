// Package order provides Redis-based repository implementation for order management.
package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sibelephant/orders-api/model"
)

// RedisRepo implements the order repository using Redis as the storage backend.
type RedisRepo struct {
	Client *redis.Client
}

// orderIDKey generates a namespaced Redis key for an order ID.
func orderIDKey(id uint64) string {
	return fmt.Sprintf("order: %d", id)
}

// Insert creates a new order using SetNX and maintains the orders set atomically.
func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order : %w", err)
	}

	key := orderIDKey(order.OrderID)
	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add to orders set:%w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

// ErrNotExist is returned when an order doesn't exist.
var ErrNotExist = errors.New("order does not exist")

// FindByID retrieves an order by ID and deserializes from JSON.
func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("get order :%w", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order json:%w", err)
	}

	return order, nil
}

// DeleteByID removes an order and its set membership atomically.
func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)

	// First check if the order exists
	exists, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check order existence: %w", err)
	}
	if exists == 0 {
		return ErrNotExist
	}

	txn := r.Client.TxPipeline()

	txn.Del(ctx, key)
	txn.SRem(ctx, "orders", key)

	_, err = txn.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to exec transaction: %w", err)
	}

	return nil
}

// Update modifies an existing order using SetXX.
func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}

	key := orderIDKey(order.OrderID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("set order: %w", err)
	}

	return nil
}

// FindAllPage defines pagination parameters.
type FindAllPage struct {
	Size   uint // Maximum number of orders to return
	Offset uint // Starting cursor for pagination
}

// FindResult contains paginated orders and next cursor.
type FindResult struct {
	Orders []model.Order // Retrieved orders
	Cursor uint64        // Next page cursor
}

// FindAll retrieves paginated orders using SSCAN and MGET for efficiency.
func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", uint64(page.Offset), "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
			Cursor: cursor,
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x := x.(string)
		var order model.Order

		err := json.Unmarshal([]byte(x), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode order json: %w", err)
		}

		orders[i] = order
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
