package evm

import (
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	"github.com/futurewei-cloud/merak/services/common/metrics"
)

// Interface for the Merak EVM
type Evm interface {
	CreateDevice(m *metrics.MerakMetrics) error
	DeleteDevice(m *metrics.MerakMetrics) error
	CreateNamespace(m *metrics.MerakMetrics) error
	DeleteNamespace(m *metrics.MerakMetrics) error
	CreateVethPair(m *metrics.MerakMetrics) error
	MoveVethToNamespace(m *metrics.MerakMetrics) error
	AssignIP(m *metrics.MerakMetrics) error
	BringInnerVethUp(m *metrics.MerakMetrics) error
	SetMTUProbing(m *metrics.MerakMetrics) error
	BringOuterVethUp(m *metrics.MerakMetrics) error
	BringLoUp(m *metrics.MerakMetrics) error
	AssignMac(m *metrics.MerakMetrics) error
	AddGateway(m *metrics.MerakMetrics) error
	CreateBridge(m *metrics.MerakMetrics) error
	DeleteBridge(m *metrics.MerakMetrics) error
	AddVethToBridge(m *metrics.MerakMetrics) error
	AddDeviceToBridge(m *metrics.MerakMetrics) error
	BringBridgeUp(m *metrics.MerakMetrics) error
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
