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
	"github.com/futurewei-cloud/merak/services/merak-compute/datastore"
	"github.com/stretchr/testify/assert"
)

func TestHostNew(t *testing.T) {
	h0 := NewHost(
		"0",
		"h1",
		"0",
		"0",
		"0",
		[]string{"10", "1"},
	)
	h1 := NewHost(
		"0",
		"v2",
		"0",
		"0",
		"0",
		[]string{"10", "1"},
	)
	assert.NotEqual(t, h0, h1)
}

func TestHostDBSuccess(t *testing.T) {
	ctx := context.Background()
	server, _ := miniredis.Run()
	datastore := datastore.NewClient(ctx, server.Addr(), "", 0)
	h0 := NewHost(
		"0",
		"h2",
		"0",
		"0",
		"0",
		[]string{"10", "1"},
	)

	oldUpdateAt := h0.UpdatedAt

	time.Sleep(time.Nanosecond * 1)
	h0.UpdateHostStore(ctx, h0.ID, datastore)
	h0New, err := GetHostStore(ctx, h0.ID, h0.ID, datastore)
	if err != nil {
		t.Error("Failed to get Host!", err)
	}

	// Not equal since UpdatedAt field is changed
	assert.NotEqual(t, oldUpdateAt, h0New.UpdatedAt)

	// Test equality for all other fields fields
	h0Reflect := reflect.ValueOf(*h0)
	h0NewReflect := reflect.ValueOf(*h0New)
	for i := 0; i < h0Reflect.NumField(); i++ {
		assert.Equal(t, h0Reflect.Field(i).Interface(), h0NewReflect.Field(i).Interface())
	}
}

func TestHostDBFail(t *testing.T) {
	ctx := context.Background()

	datastore := datastore.NewClient(ctx, "10.0.0.1", "", 0)
	h0 := NewHost(
		"0",
		"h2",
		"0",
		"0",
		"0",
		[]string{"10", "1"},
	)

	assert.Error(t, h0.UpdateHostStore(ctx, h0.ID, datastore))
	_, err := GetHostStore(ctx, h0.ID, "1", datastore)
	assert.Error(t, err)
}
