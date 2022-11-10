/*
MIT License
Copyright(c) 2022 Futurewei Cloud

	Permission is hereby granted,
	free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
	including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
	to whom the Software is furnished to do so, subject to the following conditions:
	The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

// This is a wrapper for the redis go library "go-redis"
package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/go-redis/redis/v9"
)

type redisError struct {
	Err     error
	Message string
}

type DB struct {
	Client *redis.Client
}

var _ Store = (*DB)(nil)

func (r redisError) Error() string {
	return fmt.Sprintf("%s: %v", r.Message, r.Err)
}

// Returns a new go-redis client
func NewClient(ctx context.Context, address string, password string, db int) (*DB, error) {
	addressArray := strings.Split(address, ":")
	if len(addressArray) > 2 {
		return nil, redisError{errors.New("address is invalid"), address}
	}
	if net.ParseIP(addressArray[0]) == nil {
		return nil, redisError{errors.New("IP Address is invalid"), addressArray[0]}
	}
	portInt, err := strconv.Atoi(addressArray[1])
	if err != nil {
		return nil, redisError{errors.New("redis port is not a number"), addressArray[1]}
	}
	if portInt > constants.MAX_PORT || portInt < constants.MIN_PORT {
		return nil, redisError{errors.New("redis port is not within a valid range"), addressArray[1]}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, redisError{errors.New("redis server at address is not reachable"), fmt.Sprintf("%s:%s", addressArray[0], addressArray[1])}
	}

	return &DB{client}, nil
}

// Closes the go-redis client
func (Store *DB) Close() error {
	err := Store.Client.Close()
	return err
}

// Returns the byte array value at the given key and field
func (Store *DB) HashBytesGet(ctx context.Context, key string, field string) ([]byte, error) {
	res := Store.Client.HGet(ctx, key, field)
	if err := res.Err(); err != nil {
		return nil, redisError{errors.New("failed to get, key or field does not exist"), key + " " + field}
	}
	return []byte(res.Val()), nil
}

// Updates the byte array value at the give key and field with the given value
func (Store *DB) HashBytesUpdate(ctx context.Context, key string, field string, obj []byte) error {
	if err := Store.Client.HSet(ctx, key, field, obj).Err(); err != nil {
		return redisError{errors.New("failed to update, key already in use for a different data structure"), key}
	}
	return nil
}

// Deletes the value at the given key
func (Store *DB) HashDelete(ctx context.Context, key string, field string) error {
	if err := Store.Client.HDel(ctx, key, field).Err(); err != nil {
		return redisError{errors.New("failed to delete, key in use for a different data structure"), key}
	}
	return nil
}

// Returns a string array at the given key
func (Store *DB) GetList(ctx context.Context, key string) ([]string, error) {
	res := Store.Client.LRange(ctx, key, 0, -1)
	if err := res.Err(); err != nil {
		return nil, redisError{errors.New("failed to get, key in use for a different data structure"), key}
	}
	return res.Val(), nil
}

// Appends the given value to the list at the given key
func (Store *DB) AppendToList(ctx context.Context, key string, obj string) error {
	if err := Store.Client.RPush(ctx, key, obj).Err(); err != nil {
		return redisError{errors.New("failed to append to list, key in use for a different data structure"), key}
	}
	return nil
}

func (Store *DB) PopFromList(ctx context.Context, key string) (string, error) {
	res := Store.Client.RPop(ctx, key)
	if res.Err() != nil {
		return "", redisError{errors.New("failed to pop from list, key in use for a different data structure"), key}
	}
	return res.Val(), nil
}

// Deletes the data structure at the given key
func (Store *DB) Delete(ctx context.Context, key string) error {
	if err := Store.Client.Del(ctx, key).Err(); err != nil {
		return redisError{errors.New("failed to delete key"), key}
	}
	return nil
}

// Returns a string array representation of the values at the given key
func (Store *DB) GetSet(ctx context.Context, key string) ([]string, error) {
	res := Store.Client.SMembers(ctx, key)
	if err := res.Err(); err != nil {
		return nil, redisError{errors.New("failed to get set at key"), key}
	}
	return res.Val(), nil
}

// Adds given value to the set at the given key
func (Store *DB) AddToSet(ctx context.Context, key string, obj string) error {
	if err := Store.Client.SAdd(
		ctx,
		key,
		obj,
	).Err(); err != nil {
		return redisError{errors.New("failed to add to set at key"), key}
	}
	return nil
}

// Deletes the given value from the set at the given
func (Store *DB) DeleteFromSet(ctx context.Context, key string, obj string) error {
	if err := Store.Client.SRem(ctx, key, obj).Err(); err != nil {
		return redisError{errors.New("failed to delete from set at key"), key}
	}
	return nil
}

// TODO: @cj-chung add tests and comments for functions below
func (Store *DB) Set(ctx context.Context, key string, val interface{}) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = Store.Client.Set(ctx, key, jsonVal, 0).Err()
	if err != nil {
		return redisError{errors.New("failed to append to list, key in use for a different data structure"), fmt.Sprintf("Key: %s Value: %s Error: %s", key, jsonVal, err.Error())}
	}
	return nil
}

func (Store *DB) Get(ctx context.Context, key string) (string, error) {
	val, err := Store.Client.Get(ctx, key).Result()
	if err != nil {
		return "", redisError{errors.New("database GET failed"), fmt.Sprintf("Key: %s Error: %s", key, err.Error())}
	}
	return val, nil
}

// Function for finding an entity from database
func (Store *DB) FindEntity(ctx context.Context, id string, prefix string, entity interface{}) error {
	if id == "" {
		return redisError{errors.New("invalid ID"), id}
	}
	value, err := Store.Client.Get(ctx, prefix+id).Result()
	if err != nil {
		return redisError{errors.New("failed to get"), prefix + id}
	}
	err = json.Unmarshal([]byte(value), &entity)
	return redisError{errors.New("unmarshal failed"), fmt.Sprintf("ID: %s, Error: %s", prefix+id, err.Error())}
}

func (Store *DB) GetAllValuesWithKeyPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	keys, err := getKeys(ctx, fmt.Sprintf("%s*", prefix), Store)
	if err != nil {
		return nil, err
	}

	values, err := getKeyAndValueMap(ctx, keys, prefix, Store)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func getKeys(ctx context.Context, prefix string, store *DB) ([]string, error) {
	var allkeys []string

	iter := store.Client.Scan(ctx, 0, prefix, 0).Iterator()
	for iter.Next(ctx) {
		allkeys = append(allkeys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, redisError{errors.New("database scan keys failed "), fmt.Sprintf("Prefix: %s, Error: %s", prefix, err.Error())}
	}
	return allkeys, nil
}

func getKeyAndValueMap(ctx context.Context, keys []string, prefix string, store *DB) (map[string]string, error) {
	values := make(map[string]string)
	for _, key := range keys {
		value, err := store.Client.Get(ctx, key).Result()
		if err != nil {
			return nil, redisError{errors.New("get value error when retriving key keys "), fmt.Sprintf("Prefix: %s, Error: %s", prefix, err.Error())}
		}
		// Strip off the prefix from the key so that we save the key to the value
		strippedKey := strings.Split(key, prefix)
		values[strippedKey[1]] = value
	}
	return values, nil
}
