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
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestComputeEntityNew(t *testing.T) {
	c0 := NewComputeEntity(
		"10.244.0.45",
		"10.200.0.1",
		"b15e479d-0566-464d-b1ba-916acd812b7f",
		"ff:ff:ff:ff:ff:ff",
		"vhost-0",
		STATUS_DEPLOYING,
		"vh0-eth1",
		"kind-control-plane",
	)

	c1 := NewComputeEntity(
		"10.244.0.35",
		"10.200.0.6",
		"ff35d20a-9af6-4e0a-aa5c-0b6a7a6e7c61",
		"ff:ff:ff:ff:ff:ff",
		"vhost-8",
		STATUS_DELETING,
		"vh8-eth1",
		"kind-control-plane",
	)

	assert.NotEqual(t, c0, c1)
}

func TestComputeEntitiyDBSuccess(t *testing.T) {

	client := redis.NewClient(&redis.Options{

		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	Rdb = client

	c0 := NewComputeEntity(
		"10.244.0.45",
		"10.200.0.1",
		"b15e479d-0566-464d-b1ba-916acd812b7f",
		"ff:ff:ff:ff:ff:ff",
		"vhost-0",
		STATUS_DEPLOYING,
		"vh0-eth1",
		"kind-control-plane",
	)
	err := SetValue(c0.Name, c0)
	assert.Nil(t, err)
}
