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
package entities

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/futurewei-cloud/merak/services/common/datastore"
	"github.com/go-redis/redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestVMNew(t *testing.T) {
	v0 := NewVM(
		"0",
		"v1",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
	)
	v1 := NewVM(
		"0",
		"v1",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
		"0",
	)
	assert.NotEqual(t, v0, v1)
}

func TestVMDBSuccess(t *testing.T) {
	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	datastore, _ := datastore.NewClient(ctx, server.Addr(), "", 0)
	defer datastore.Close()
	v0 := NewVM(
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		"10",
		"11",
		"12",
		"13",
	)

	oldUpdateAt := v0.UpdatedAt

	time.Sleep(time.Nanosecond * 1)
	v0.UpdateVMStore(ctx, v0.ID, datastore)
	v0New, err := GetVMStore(ctx, v0.ID, v0.ID, datastore)
	if err != nil {
		t.Error("Failed to get VM!", err)
	}

	// Not equal since UpdatedAt field is changed
	assert.NotEqual(t, oldUpdateAt, v0New.UpdatedAt)

	// Test equality for all other fields fields
	v0Reflect := reflect.ValueOf(*v0)
	v0NewReflect := reflect.ValueOf(*v0New)
	for i := 0; i < v0Reflect.NumField(); i++ {
		assert.Equal(t, v0Reflect.Field(i).Interface(), v0NewReflect.Field(i).Interface())
	}
}

func TestVMDBFail(t *testing.T) {
	ctx := context.Background()

	datastore := datastore.DB{Client: redis.NewClient(&redis.Options{
		Addr:     "10.0.0.1",
		Password: "0",
		DB:       0,
	})}
	v0 := NewVM(
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		"10",
		"11",
		"12",
		"13",
	)

	assert.Error(t, v0.UpdateVMStore(ctx, v0.ID, &datastore))
	_, err := GetVMStore(ctx, v0.ID, "1", &datastore)
	assert.Error(t, err)
}
