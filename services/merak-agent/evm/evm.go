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

package evm

import (
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	"github.com/futurewei-cloud/merak/services/common/metrics"
)

// Interface for the Merak EVM
type Evm interface {
	CreateDevice(m *metrics.MerakMetrics) error
	DeleteDevice(m *metrics.MerakMetrics) error
	MoveDeviceToNamespace(m *metrics.MerakMetrics) error
	AssignIP(m *metrics.MerakMetrics) error
	SetMTUProbing(m *metrics.MerakMetrics) error
	BringLoUp(m *metrics.MerakMetrics) error
	AssignMac(m *metrics.MerakMetrics) error
	AddGateway(m *metrics.MerakMetrics) error
	BringDeviceUp(m *metrics.MerakMetrics) error
	GetName() string
	GetIP() string
	GetMac() string
	GetCidr() string
	GetGw() string
	GetDeviceId() string
	GetRemoteId() string
	GetStatus() common_pb.Status
	SetName(name string)
	SetIP(ip string) error
	SetMac(mac string) error
	SetCidr(cidr string) error
	SetGw(gw string) error
	SetDeviceId(id string)
	SetRemoteId(id string)
	SetStatus(status common_pb.Status)
}
