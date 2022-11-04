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

//This is a wrapper for the redis go library "go-redis"
package datastore

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v9"
)

type redisError struct {
	Err     error
	Message string
}

type DB struct {
	Client *redis.Client
}

func (r redisError) Error() string {
	return fmt.Sprintf("%s: %v", r.Message, r.Err)
}

// Returns a new go-redis client
func NewClient(ctx context.Context, address string, password string, db int) *DB {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	return &DB{client}
}

// Returns the byte array value at the given id and field
func (store *DB) Get(ctx context.Context, id string, field string) ([]byte, error) {
	res := store.Client.HGet(ctx, id, field)
	if err := res.Err(); err != nil {
		return nil, err
	}
	if res != nil {
		return []byte(res.Val()), nil
	}
	return nil, redisError{errors.New("key does not exist"), id}
}

// Updates the byte array value at the give id and field with the given value
func (store *DB) Update(ctx context.Context, id string, field string, obj []byte) error {
	if err := store.Client.HSet(ctx, id, field, obj).Err(); err != nil {
		return err
	}
	return nil
}

func (store *DB) Delete(ctx context.Context, id string) error {
	if err := store.Client.HDel(ctx, id).Err(); err != nil {
		return err
	}
	return nil
}

// Returns a string array at the given ID
func (store *DB) GetList(ctx context.Context, id string) ([]string, error) {
	res := store.Client.LRange(ctx, id, 0, -1)
	if err := res.Err(); err != nil {
		return nil, err
	}
	if res != nil {
		return res.Val(), nil
	}
	return nil, redisError{errors.New("key does not exist"), id}
}

// Appends the given value to the list at the given ID
func (store *DB) AppendToList(ctx context.Context, id string, obj string) error {
	if err := store.Client.RPush(ctx, id, obj).Err(); err != nil {
		return err
	}
	return nil
}

// Deletes the list at the given ID
func (store *DB) DeleteList(ctx context.Context, id string) error {
	if err := store.Client.Del(ctx, id).Err(); err != nil {
		return err
	}
	return nil
}

// Returns a string array representation of the values at the given ID
func (store *DB) GetSet(ctx context.Context, id string) ([]string, error) {
	res := store.Client.SMembers(ctx, id)
	if err := res.Err(); err != nil {
		return nil, err
	}
	if res != nil {
		return res.Val(), nil
	}
	return nil, redisError{errors.New("key does not exist"), id}
}

// Adds given value to the set at the given ID
func (store *DB) AddToSet(ctx context.Context, id string, obj string) error {
	if err := store.Client.SAdd(
		ctx,
		id,
		obj,
	).Err(); err != nil {
		return err
	}
	return nil
}

// Deletes the given value from the set at the given
func (store *DB) DeleteFromSet(ctx context.Context, id string, obj string) error {
	if err := store.Client.SRem(ctx, id, obj).Err(); err != nil {
		return err
	}
	return nil
}
