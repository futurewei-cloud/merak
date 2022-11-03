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
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/futurewei-cloud/merak/services/merak-compute/datastore"
	"github.com/stretchr/testify/assert"
)

func TestVM(t *testing.T) {
	ctx := context.Background()
	server, _ := miniredis.Run()
	datastore := datastore.NewClient(ctx, server.Addr(), "", 0)
	v := NewVM(
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

	v.UpdateStore(ctx, v.ID, v.ID, datastore)
	vNew, err := GetVM(ctx, v.ID, v.ID, datastore)
	if err != nil {
		t.Error("Failed to get VM!", err)
	}
	assert.Equal(t, v, vNew)
}
