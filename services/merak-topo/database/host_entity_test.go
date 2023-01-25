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
	client := redis.NewClient(&redis.Options{

		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	Rdb = client

	h0 := NewHostEntity(
		"10.0.0.1",
		STATUS_DONE,
		"10.0.0.0/24 dev wlp3s0 proto kernel scope link src 10.0.0.73 metric 600",
	)
	err := SetValue(h0.Ip, h0)
	assert.Nil(t, err)
}

func TestHostEntityDBFail(t *testing.T) {

	client := redis.NewClient(&redis.Options{

		Addr:     "localhost:0000",
		Password: "",
		DB:       0,
	})

	Rdb = client

	h1 := NewHostEntity(
		"10.0.0.1",
		STATUS_DEPLOYING,
		" 172.17.0.0/16 dev docker0 proto kernel scope link src 172.17.0.1 linkdown",
	)
	err := SetValue(h1.Ip, h1)
	assert.Error(t, err)
}
