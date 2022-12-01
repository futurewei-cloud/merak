package evm

import common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"

// Interface for the Merak EVM
type Evm interface {
	CreateDevice() error
	DeleteDevice() error
	CreateNamespace() error
	DeleteNamespace() error
	CreateVethPair() error
	MoveVethToNamespace() error
	AssignIP() error
	BringInnerVethUp() error
	SetMTUProbing() error
	BringOuterVethUp() error
	BringLoUp() error
	AssignMac() error
	AddGateway() error
	CreateBridge() error
	DeleteBridge() error
	AddVethToBridge() error
	AddDeviceToBridge() error
	BringBridgeUp() error
	BringDeviceUp() error
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
