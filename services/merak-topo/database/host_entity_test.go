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

package database

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/futurewei-cloud/merak/services/common/datastore"
	"github.com/go-redis/redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestHostEntityNew(t *testing.T) {
	h0 := NewHostEntity(
		"10.0.0.1",
		STATUS_DONE,
		"10.0.0.0/24 dev wlp3s0 proto kernel scope link src 10.0.0.73 metric 600",
	)
	h1 := NewHostEntity(
		"10.0.0.1",
		STATUS_DEPLOYING,
		" 172.17.0.0/16 dev docker0 proto kernel scope link src 172.17.0.1 linkdown",
	)

	assert.NotEqual(t, h0, h1)

}

func TestHostEntitiyDBSuccess(t *testing.T) {

	ctx := context.Background()
	server, _ := miniredis.Run()
	defer server.Close()
	datastore, _ := datastore.NewClient(ctx, server.Addr(), "", 0)
	defer datastore.Close()

	h0 := NewHostEntity(
		"10.0.0.1",
		STATUS_DONE,
		"10.0.0.0/24 dev wlp3s0 proto kernel scope link src 10.0.0.73 metric 600",
	)
	jsonBytes, _ := json.Marshal(h0)
	err := datastore.HashBytesUpdate(ctx, "host_entity", h0.Ip, jsonBytes)
	assert.Nil(t, err)
}

func TestHostEntityDBFail(t *testing.T) {
	ctx := context.Background()
	datastore := datastore.DB{Client: redis.NewClient(&redis.Options{
		Addr:     "10.0.0.1",
		Password: "0",
		DB:       0,
	})}
	defer datastore.Close()
	h1 := NewHostEntity(
		"10.0.0.1",
		STATUS_DEPLOYING,
		" 172.17.0.0/16 dev docker0 proto kernel scope link src 172.17.0.1 linkdown",
	)

	jsonBytes, _ := json.Marshal(h1)
	err := datastore.HashBytesUpdate(ctx, "host_entity", h1.Ip, jsonBytes)
	assert.Error(t, err)
}
