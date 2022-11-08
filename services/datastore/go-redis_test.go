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
package datastore

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestDatastoreNewClient(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		give string
		err  error
	}{
		{
			give: "192.0.2.0:8000",
			err:  redisError{errors.New("redis server at address is not reachable"), "192.0.2.0:8000"},
		},
		{
			give: "192.0.2.0:8000:12",
			err:  redisError{errors.New("address is invalid"), "192.0.2.0:8000:12"},
		},
		{
			give: "123123.0:http",
			err:  redisError{errors.New("IP Address is invalid"), "123123.0"},
		},
		{
			give: "192.0.2.0:a",
			err:  redisError{errors.New("redis port is not a number"), "a"},
		},
		{
			give: "192.0.2.0:9999999",
			err:  redisError{errors.New("redis port is not within a valid range"), "9999999"},
		},
	}
	errorTest := redisError{errors.New("this is an error"), "hi"}
	t.Log(errorTest)

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			_, err := NewClient(ctx, tt.give, "", 0)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestDatastoreHashGet(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		give []string
		exp  []byte
		err  error
	}{
		{
			give: []string{"123", "1"},
			exp:  []byte("1234"),
			err:  nil,
		},
		{
			give: []string{"123", "2"},
			exp:  nil,
			err:  redisError{errors.New("failed to get, key or field does not exist"), "123 2"},
		},
		{
			give: []string{"1", "2"},
			exp:  nil,
			err:  redisError{errors.New("failed to get, key or field does not exist"), "1 2"},
		},
	}

	server, _ := miniredis.Run()
	defer server.Close()
	client, _ := NewClient(ctx, server.Addr(), "", 0)
	defer client.Close()
	client.HashBytesUpdate(ctx, "123", "1", []byte("1234"))
	for _, tt := range tests {
		t.Run(tt.give[0], func(t *testing.T) {
			res, err := client.HashBytesGet(ctx, tt.give[0], tt.give[1])
			assert.Equal(t, tt.exp, res)
			assert.Equal(t, tt.err, err)
		})
	}

}

func TestDatastoreHashUpdate(t *testing.T) {
	tests := []struct {
		key   string
		field string
		give  []byte
		exp   []byte
		err   error
	}{
		{
			key:   "1",
			field: "2",
			give:  []byte{1, 2, 3},
			exp:   []byte{1, 2, 3},
			err:   nil,
		},
		{
			key:   "3",
			field: "3",
			give:  []byte{4, 5, 6, 7},
			exp:   []byte{4, 5, 6, 7},
			err:   nil,
		},
		{
			key:   "123",
			field: "456",
			give:  []byte{5, 6, 7, 8},
			exp:   []byte{5, 6, 7, 8},
			err:   nil,
		},
		{
			key:   "12345",
			field: "456",
			give:  []byte{5, 6, 7, 8},
			exp:   nil,
			err:   redisError{errors.New("failed to update, key already in use for a different data structure"), "12345"},
		},
	}
	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	client, _ := NewClient(ctx, server.Addr(), "", 0)
	defer client.Close()

	// Test for failure
	client.AddToSet(ctx, "12345", "456")

	for _, tt := range tests {
		t.Run(fmt.Sprintf("key: %s, field: %s, give: %s, exp: %s", tt.key, tt.field, tt.give, tt.exp), func(t *testing.T) {
			err := client.HashBytesUpdate(ctx, tt.key, tt.field, tt.give)
			assert.Equal(t, err, tt.err)
			obj, _ := client.HashBytesGet(ctx, tt.key, tt.field)
			assert.Equal(t, tt.exp, obj)
		})
	}

}

func TestDatastoreHashDelete(t *testing.T) {
	tests := []struct {
		key   string
		field string
		give  []byte
		err   error
	}{
		{
			key:   "1",
			field: "2",
			give:  []byte{1, 2, 3},
			err:   nil,
		},
		{
			key:   "3",
			field: "3",
			give:  []byte{4, 5, 6, 7},
			err:   nil,
		},
		{
			key:   "123",
			field: "456",
			give:  []byte{5, 6, 7, 8},
			err:   nil,
		},
	}
	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	client, _ := NewClient(ctx, server.Addr(), "", 0)
	defer client.Close()
	t.Run(("Test Failure"), func(t *testing.T) {
		client.AddToSet(ctx, "12345", "456")
		err := client.HashDelete(ctx, "12345", "456")
		assert.Equal(t, redisError{errors.New("failed to delete, key in use for a different data structure"), "12345"}, err)
	})

	for _, tt := range tests {
		t.Run(fmt.Sprintf("key: %s, field: %s, give: %s", tt.key, tt.field, tt.give), func(t *testing.T) {
			err := client.HashBytesUpdate(ctx, tt.key, tt.field, tt.give)
			assert.Nil(t, err)
			err = client.HashDelete(ctx, tt.key, tt.field)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestDatastoreGetList(t *testing.T) {
	tests := []struct {
		key  string
		give string
		exp  []string
		err  error
	}{
		{
			key:  "1",
			give: "1",
			exp:  []string{"1"},
			err:  nil,
		},
		{
			key:  "1",
			give: "2",
			exp:  []string{"1", "2"},
			err:  nil,
		},
		{
			key:  "1",
			give: "3",
			exp:  []string{"1", "2", "3"},
			err:  nil,
		},
	}
	testsFail := []struct {
		key  string
		give string
		exp  []string
		err  error
	}{
		{
			key: "2",
			exp: []string{},
			err: nil,
		},
		{
			key: "3",
			exp: nil,
			err: redisError{errors.New("failed to get, key in use for a different data structure"), "3"},
		},
	}
	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	client, _ := NewClient(ctx, server.Addr(), "", 0)
	defer client.Close()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Test Get Success: Append: %s Expect: %s", tt.give, tt.exp), func(t *testing.T) {
			err := client.AppendToList(ctx, tt.key, tt.give)
			assert.Nil(t, tt.err, err)
			res, err := client.GetList(ctx, tt.key)
			assert.Nil(t, tt.err, err)
			assert.Equal(t, tt.exp, res)
		})
	}
	client.HashBytesUpdate(ctx, "3", "3", []byte{1, 2, 3})
	for _, tt := range testsFail {
		t.Run("Test GetList Fail", func(t *testing.T) {
			res, err := client.GetList(ctx, tt.key)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.exp, res)
		})
	}
}
func TestDatastoreAppendToList(t *testing.T) {
	tests := []struct {
		key  string
		give string
		exp  []string
		err  error
	}{
		{
			key:  "1",
			give: "1",
			exp:  []string{"1"},
			err:  nil,
		},
		{
			key:  "1",
			give: "2",
			exp:  []string{"1", "2"},
			err:  nil,
		},
		{
			key:  "1",
			give: "3",
			exp:  []string{"1", "2", "3"},
			err:  nil,
		},
	}
	testsFail := []struct {
		key  string
		give string
		err  error
	}{
		{
			key: "2",
			err: nil,
		},
		{
			key: "3",
			err: redisError{errors.New("failed to append to list, key in use for a different data structure"), "3"},
		},
	}
	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	client, _ := NewClient(ctx, server.Addr(), "", 0)
	defer client.Close()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Test Append Success: Append: %s Expect: %s", tt.give, tt.exp), func(t *testing.T) {
			err := client.AppendToList(ctx, tt.key, tt.give)
			assert.Nil(t, tt.err, err)
			res, err := client.GetList(ctx, tt.key)
			assert.Nil(t, tt.err, err)
			assert.Equal(t, tt.exp, res)
		})
	}
	client.HashBytesUpdate(ctx, "3", "3", []byte{1, 2, 3})
	for _, tt := range testsFail {
		t.Run("Test GetList Fail", func(t *testing.T) {
			err := client.AppendToList(ctx, tt.key, "1")
			assert.Equal(t, tt.err, err)
		})
	}
}
func TestDatastoreDelete(t *testing.T) {
	tests := []struct {
		key  string
		give string
		err  error
	}{
		{
			key:  "1",
			give: "1",
			err:  nil,
		},
		{
			key:  "2",
			give: "2",
			err:  nil,
		},
		{
			key:  "3",
			give: "3",
			err:  nil,
		},
	}
	testFail := []struct {
		key  string
		give string
		err  error
	}{
		{
			key:  "4",
			give: "5",
			err:  redisError{errors.New("failed to delete key"), "4"},
		},
		{
			key:  "5",
			give: "2",
			err:  redisError{errors.New("failed to delete key"), "5"},
		},
		{
			key:  "6",
			give: "3",
			err:  redisError{errors.New("failed to delete key"), "6"},
		},
	}
	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	client, _ := NewClient(ctx, server.Addr(), "", 0)
	defer client.Close()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Test Delete Success: Append: %s", tt.give), func(t *testing.T) {
			err := client.AppendToList(ctx, tt.key, tt.give)
			assert.Nil(t, tt.err, err)
			err = client.Delete(ctx, tt.key)
			assert.Nil(t, tt.err, err)

		})
	}
	client = &DB{Client: redis.NewClient(&redis.Options{
		Addr:     "10.0.0.1",
		Password: "0",
		DB:       0,
	})}
	for _, tt := range testFail {
		t.Run(fmt.Sprintf("Test Delete Fail: Append: %s", tt.give), func(t *testing.T) {
			err := client.Delete(ctx, tt.key)
			assert.Error(t, err)

		})
	}
}
func TestDatastoreGetSet(t *testing.T) {
	tests := []struct {
		key  string
		give string
		exp  []string
		err  error
	}{
		{
			key:  "1",
			give: "1",
			exp:  []string{"1"},
			err:  nil,
		},
		{
			key:  "1",
			give: "2",
			exp:  []string{"1", "2"},
			err:  nil,
		},
		{
			key:  "1",
			give: "3",
			exp:  []string{"1", "2", "3"},
			err:  nil,
		},
	}
	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	client, _ := NewClient(ctx, server.Addr(), "", 0)
	defer client.Close()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Test Set Get Success: key: %s add: %s", tt.key, tt.give), func(t *testing.T) {
			err := client.AddToSet(ctx, tt.key, tt.give)
			assert.Nil(t, tt.err, err)
			res, err := client.GetSet(ctx, tt.key)
			assert.Equal(t, tt.exp, res)
			assert.Nil(t, err)

		})
	}
	client.HashBytesUpdate(ctx, "2", "2", []byte{1})
	res, err := client.GetSet(ctx, "2")
	expErr := redisError{errors.New("failed to get set at key"), "2"}
	assert.Equal(t, expErr, err)
	assert.Nil(t, res)
}
func TestDatastoreAddToSet(t *testing.T) {
	tests := []struct {
		key  string
		give string
		exp  []string
		err  error
	}{
		{
			key:  "1",
			give: "1",
			exp:  []string{"1"},
			err:  nil,
		},
		{
			key:  "1",
			give: "2",
			exp:  []string{"1", "2"},
			err:  nil,
		},
		{
			key:  "1",
			give: "3",
			exp:  []string{"1", "2", "3"},
			err:  nil,
		},
	}
	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	client, _ := NewClient(ctx, server.Addr(), "", 0)
	defer client.Close()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Test Set Add Success: key: %s add: %s", tt.key, tt.give), func(t *testing.T) {
			err := client.AddToSet(ctx, tt.key, tt.give)
			assert.Nil(t, tt.err, err)
			res, err := client.GetSet(ctx, tt.key)
			assert.Equal(t, tt.exp, res)
			assert.Nil(t, err)

		})
	}
	client.HashBytesUpdate(ctx, "2", "2", []byte{1})
	err := client.AddToSet(ctx, "2", "2")
	expErr := redisError{errors.New("failed to add to set at key"), "2"}
	assert.Equal(t, expErr, err)
}
func TestDatastoreDeleteSet(t *testing.T) {
	tests := []struct {
		key  string
		give string
		exp  []string
		err  error
	}{
		{
			key:  "1",
			give: "1",
			exp:  []string{"2", "3"},
			err:  nil,
		},
		{
			key:  "1",
			give: "2",
			exp:  []string{"3"},
			err:  nil,
		},
		{
			key:  "1",
			give: "3",
			exp:  []string{},
			err:  nil,
		},
	}
	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	client, _ := NewClient(ctx, server.Addr(), "", 0)
	defer client.Close()

	client.AddToSet(ctx, "1", "1")
	client.AddToSet(ctx, "1", "2")
	client.AddToSet(ctx, "1", "3")

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Test Delete Get Success: key: %s add: %s", tt.key, tt.give), func(t *testing.T) {
			err := client.DeleteFromSet(ctx, tt.key, tt.give)
			assert.Nil(t, tt.err, err)
			res, err := client.GetSet(ctx, tt.key)
			assert.Equal(t, tt.exp, res)
			assert.Nil(t, err)

		})
	}
	client.HashBytesUpdate(ctx, "2", "2", []byte{1})
	err := client.DeleteFromSet(ctx, "2", "2")
	expErr := redisError{errors.New("failed to delete from set at key"), "2"}
	assert.Equal(t, expErr, err)
}
