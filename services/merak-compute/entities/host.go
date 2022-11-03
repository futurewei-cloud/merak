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
func NewHost(ID, Name, IP, MAC, Interface string, VMs []string) (*Host, error) {
	h := &Host{
		ID:           ID,
		Name:         Name,
		IP:           IP,
		MAC:          MAC,
		Interface:    Interface,
		VMs:          VMs,
		RegisteredAt: time.Now(),
		UpdatedAt:    time.Now(),
	}
	return h, nil
}

func (*Host) Update(redisClient redis.Client) error {
	return nil
}

func GetHost(id string, redisClient redis.Client) (*VM, error) {
	return nil, nil
}

func SetAddHost(id string, redisClient redis.Client) error {
	return nil
}
