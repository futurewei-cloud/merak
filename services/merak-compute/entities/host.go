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

// Host entity

package entities

import (
	"context"
	"encoding/json"
	"time"

	"github.com/futurewei-cloud/merak/services/datastore"
)

type Host struct {
	ID           string
	Name         string
	IP           string
	MAC          string
	Interface    string
	VMs          []string
	RegisteredAt time.Time
	UpdatedAt    time.Time
}

// Create a new host
func NewHost(ID, Name, IP, MAC, Interface string, VMs []string) *Host {
	h := &Host{
		ID:           ID,
		Name:         Name,
		IP:           IP,
		MAC:          MAC,
		Interface:    Interface,
		VMs:          VMs,
		RegisteredAt: time.Now().Round(0), // Strip monotonic clock reading
		UpdatedAt:    time.Now().Round(0), // Strip monotonic clock reading
	}
	return h
}

// Converts Host to json byte array and updates DB
func (h *Host) UpdateHostStore(ctx context.Context, id string, datastore *datastore.DB) error {
	h.UpdatedAt = time.Now().Round(0) // Strip monotonic clock reading before marshal
	jsonBytes, _ := json.Marshal(h)   // Should never fail

	err := datastore.HashBytesUpdate(ctx, id, h.ID, jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

// Unmarshals and returns Host struct from DB
func GetHostStore(ctx context.Context, id string, field string, datastore *datastore.DB) (*Host, error) {
	vString, err := datastore.HashBytesGet(ctx, id, field)
	if err != nil {
		return nil, err
	}
	var host Host
	_ = json.Unmarshal(vString, &host) // Should never fail
	return &host, nil
}
