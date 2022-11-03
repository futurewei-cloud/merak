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
	"time"

	"github.com/go-redis/redis/v9"
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
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
	Status string) (*VM, error) {
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
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	return v, nil
}

func (*VM) Update(redisClient redis.Client) error {
	return nil
}

func GetVM(id string, redisClient redis.Client) (*VM, error) {
	return nil, nil
}

func SetAddVM(id string, redisClient redis.Client) error {
	return nil
}
