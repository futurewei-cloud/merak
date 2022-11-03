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
	"encoding/json"
	"time"

	"github.com/futurewei-cloud/merak/services/merak-compute/datastore"
)

type VM struct {
	ID            string
	Name          string
	VPC           string
	SubnetID      string
	TenantID      string
	ProjectID     string
	CIDR          string
	Gateway       string
	SecurityGroup string
	HostIP        string
	HostMAC       string
	HostName      string
	Status        string
	CreatedAt     string
	UpdatedAt     string
}

// Create a new VM
func NewVM(
	ID,
	Name,
	VPC,
	SubnetID,
	TenantID,
	ProjectID,
	CIDR,
	Gateway,
	SecurityGroup,
	HostIP,
	HostMAC,
	HostName,
	Status string) *VM {
	v := &VM{
		ID:            ID,
		Name:          Name,
		VPC:           VPC,
		SubnetID:      SubnetID,
		TenantID:      TenantID,
		ProjectID:     ProjectID,
		CIDR:          CIDR,
		Gateway:       Gateway,
		SecurityGroup: SecurityGroup,
		HostIP:        HostIP,
		HostMAC:       HostMAC,
		HostName:      HostName,
		Status:        Status,
		CreatedAt:     time.Now().String(),
		UpdatedAt:     time.Now().String(),
	}
	return v
}

func (v *VM) UpdateStore(ctx context.Context, id string, field string, datastore *datastore.DB) error {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	datastore.Update(ctx, id, field, jsonBytes)
	return nil
}

func GetVM(ctx context.Context, id string, field string, datastore *datastore.DB) (*VM, error) {
	vString, err := datastore.Get(ctx, id, field)
	if err != nil {
		return nil, err
	}
	var vm VM
	err = json.Unmarshal(vString, &vm)
	if err != nil {
		return nil, err
	}
	return &vm, nil
}

func SetAddVM(ctx context.Context, id string, datastore *datastore.DB) error {
	return nil
}
