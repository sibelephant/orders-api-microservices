// Package order provides Redis-based repository implementation for order management.
// This package implements the repository pattern for storing and retrieving orders
// from a Redis database, providing CRUD operations with JSON serialization.
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
// It provides methods for creating, reading, and deleting orders stored as JSON strings.
type RedisRepo struct {
	Client *redis.Client // Redis client for database operations
}

// orderIDKey generates a Redis key for storing an order by its ID.
// The key format is "order: {id}" which creates a namespaced key in Redis.
// This helps organize data and avoid key collisions with other entities.
func orderIDKey(id uint64) string {
	return fmt.Sprintf("order: %d", id)
}

// Insert creates a new order in Redis storage.
// It uses SetNX (Set if Not eXists) to ensure we don't overwrite existing orders.
// The order is serialized to JSON before storage.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - order: The order model to be stored
//
// Returns:
//   - error: nil on success, or an error describing what went wrong
func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	// Serialize the order struct to JSON format for storage
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order : %w", err)
	}

	// Generate the Redis key using the order's ID
	key := orderIDKey(order.OrderID)

	// SetNX ensures atomic "create only if not exists" operation
	// The third parameter (0) means no expiration time
	res := r.Client.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		return fmt.Errorf("failed to set: %w", err)
	}

	return nil
}

// ErrNotExist is returned when attempting to access an order that doesn't exist in Redis.
// This provides a consistent way to handle "not found" scenarios across the application.
var ErrNotExist = errors.New("order does not exist")

// FindByID retrieves an order from Redis by its unique identifier.
// It performs a GET operation on the Redis key and deserializes the JSON data
// back into an Order struct.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - id: The unique identifier of the order to retrieve
//
// Returns:
//   - model.Order: The retrieved order data
//   - error: ErrNotExist if order doesn't exist, or other errors for Redis/JSON issues
func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	// Generate the Redis key for this order ID
	key := orderIDKey(id)

	// Attempt to get the value from Redis
	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		// redis.Nil is returned when the key doesn't exist
		return model.Order{}, ErrNotExist
	} else if err != nil {
		// Handle other Redis errors (connection issues, etc.)
		return model.Order{}, fmt.Errorf("get order :%w", err)
	}

	// Deserialize the JSON string back into an Order struct
	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		// Handle JSON parsing errors (corrupted data, schema changes, etc.)
		return model.Order{}, fmt.Errorf("failed to decode order json:%w", err)
	}

	return order, nil
}

// DeleteByID removes an order from Redis storage by its unique identifier.
// It uses the DEL command to permanently remove the order data.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - id: The unique identifier of the order to delete
//
// Returns:
//   - error: ErrNotExist if order doesn't exist, or other errors for Redis issues
func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	// Generate the Redis key for this order ID
	key := orderIDKey(id)

	// Delete the key from Redis
	err := r.Client.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		// Note: DEL command typically doesn't return redis.Nil
		// This check might be unnecessary, but kept for consistency
		return ErrNotExist
	} else if err != nil {
		// Handle Redis errors (connection issues, etc.)
		return fmt.Errorf("get order:%w", err)
	}

	return nil
}
