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

// Insert creates a new order in Redis storage using a transactional approach.
// It uses SetNX (Set if Not eXists) to ensure we don't overwrite existing orders.
// Additionally, it maintains an "orders" set for efficient listing operations.
// The order is serialized to JSON before storage.
//
// Redis Transaction Details:
// - Uses TxPipeline() for atomic multi-command execution
// - Ensures both order storage AND set membership happen together
// - If any command fails, the entire transaction is discarded
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

	// Start a Redis transaction pipeline for atomic operations
	// TxPipeline queues commands but doesn't execute until Exec() is called
	txn := r.Client.TxPipeline()

	// SetNX ensures atomic "create only if not exists" operation
	// The third parameter (0) means no expiration time
	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		// Discard the transaction if SetNX fails
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	// Add the order key to a Redis set for efficient listing/pagination
	// SADD adds the key to the "orders" set, enabling FindAll operations
	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add to orders set:%w", err)
	}

	// Execute all queued commands atomically
	// Either all commands succeed, or none are applied
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
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
// It uses a transactional approach to ensure both the order data AND its
// membership in the "orders" set are removed atomically.
//
// Redis Transaction Details:
// - Uses TxPipeline() for atomic multi-command execution
// - DEL removes the order data from Redis
// - SREM removes the order key from the "orders" set
// - Both operations must succeed or the transaction is rolled back
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

	// Start a Redis transaction pipeline for atomic operations
	txn := r.Client.TxPipeline()

	// Delete the key from Redis
	// DEL command removes the order data completely
	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		// Note: DEL command typically doesn't return redis.Nil
		// This check might be unnecessary, but kept for consistency
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		// Handle Redis errors (connection issues, etc.)
		return fmt.Errorf("delete order: %w", err)
	}

	// Remove the order key from the "orders" set
	// SREM ensures the order won't appear in FindAll results
	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove from orders set : %w", err)
	}

	// Execute all queued commands atomically
	// Ensures data consistency between order storage and set membership
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec transaction: %w", err)
	}

	return nil
}

// Update modifies an existing order in Redis storage.
// It uses SetXX (Set if eXists) to ensure we only update existing orders.
// Unlike Insert, this doesn't use transactions since it only updates the order data,
// not the set membership (the order key remains the same).
//
// Redis SetXX Details:
// - SetXX only succeeds if the key already exists
// - Returns redis.Nil if the key doesn't exist (order not found)
// - Overwrites existing data with new JSON serialization
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - order: The updated order model to be stored
//
// Returns:
//   - error: ErrNotExist if order doesn't exist, or other errors for Redis/JSON issues
func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	// Serialize the updated order struct to JSON format
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}

	// Generate the Redis key using the order's ID
	key := orderIDKey(order.OrderID)

	// SetXX ensures we only update existing orders (no creation)
	// The third parameter (0) means no expiration time
	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		// redis.Nil indicates the key doesn't exist
		return ErrNotExist
	} else if err != nil {
		// Handle other Redis errors (connection issues, etc.)
		return fmt.Errorf("set order: %w", err)
	}

	return nil
}

// FindAllPage represents pagination parameters for listing orders.
// This struct defines how many orders to return and where to start.
type FindAllPage struct {
	Size   uint // Maximum number of orders to return (limit)
	Offset uint // Number of orders to skip (for pagination)
}

// FindAll retrieves a paginated list of all orders from Redis storage.
// It uses the "orders" Redis set to efficiently find all order keys,
// then fetches the actual order data for each key.
//
// Redis Set Operations:
// - Uses SSCAN to iterate through the "orders" set efficiently
// - Supports pagination through cursor-based iteration
// - Avoids loading all keys into memory at once (memory efficient)
//
// Implementation Strategy:
// 1. Scan the "orders" set to get order keys
// 2. Use pagination to limit the number of keys retrieved
// 3. For each key, fetch the actual order data using GET
// 4. Deserialize JSON data back into Order structs
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - page: Pagination parameters (size and offset)
//
// Returns:
//   - []model.Order: Slice of orders for the requested page
//   - error: Any errors encountered during retrieval or JSON parsing
//
// Note: This method is currently a placeholder and returns nil.
// A full implementation would use SSCAN for efficient set iteration.
func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) ([]model.Order, error) {
	// TODO: Implement pagination using Redis SSCAN command
	// 1. Use SSCAN to iterate through "orders" set with cursor
	// 2. Apply offset/limit logic for pagination
	// 3. For each key, GET the order data
	// 4. Unmarshal JSON and collect results
	// 5. Return the paginated slice of orders

	return nil, nil
}
