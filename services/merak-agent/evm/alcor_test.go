package evm

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	"github.com/futurewei-cloud/merak/services/common/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

type mockMetrics struct {
	ServiceName     string
	OpsTotalLatency *prometheus.HistogramVec
	OpsSuccess      *prometheus.CounterVec
	OpsFail         *prometheus.CounterVec
}

func (m *mockMetrics) GetMetrics(*error) func() {
	return func() {}
}

func TestNewEvm(t *testing.T) {
	tests := []struct {
		giveName     string
		giveIP       string
		giveMAC      string
		giveRemoteID string
		giveDeviceID string
		giveCIDR     string
		giveGW       string
		giveStatus   common_pb.Status

		expRes Evm
		expErr error
		pass   bool
	}{
		{
			giveName:     "vm1",
			giveIP:       "10.0.0.2",
			giveMAC:      "00:00:00:00:00:01",
			giveRemoteID: "12345",
			giveDeviceID: "tap1",
			giveCIDR:     "10.0.0.0/16",
			giveGW:       "10.0.0.1",
			giveStatus:   common_pb.Status_DEPLOYING,

			expRes: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				mac:      "00:00:00:00:00:01",
				remoteID: "12345",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
				gw:       "10.0.0.1",
				status:   common_pb.Status_DEPLOYING,
			},
			expErr: nil,
			pass:   true,
		},
		{
			giveName:     "vm1",
			giveIP:       "10.0.02",
			giveMAC:      "00:00:00:00:00:01",
			giveRemoteID: "12345",
			giveDeviceID: "tap1",
			giveCIDR:     "10.0.0.0/16",
			giveGW:       "10.0.0.1",
			giveStatus:   common_pb.Status_DEPLOYING,

			expRes: nil,
			expErr: evmError{errors.New("Invalid IP Address"), "10.0.02"},
			pass:   false,
		},
		{
			giveName:     "vm1",
			giveIP:       "10.0.0.2",
			giveMAC:      "00:00:00:00:00,01",
			giveRemoteID: "12345",
			giveDeviceID: "tap1",
			giveCIDR:     "10.0.0.0/16",
			giveGW:       "10.0.0.1",
			giveStatus:   common_pb.Status_DEPLOYING,

			expRes: nil,
			expErr: evmError{errors.New("Invalid MAC address"), "00:00:00:00:00,01"},
			pass:   false,
		},
		{
			giveName:     "vm1",
			giveIP:       "10.0.0.2",
			giveMAC:      "00:00:00:00:00:01",
			giveRemoteID: "12345",
			giveDeviceID: "tap1",
			giveCIDR:     "10.0.0.016",
			giveGW:       "10.0.0.1",
			giveStatus:   common_pb.Status_DEPLOYING,

			expRes: nil,
			expErr: evmError{errors.New("Invalid CIDR address"), "10.0.0.016"},
			pass:   false,
		},
		{
			giveName:     "vm1",
			giveIP:       "10.0.0.2",
			giveMAC:      "00:00:00:00:00:01",
			giveRemoteID: "12345",
			giveDeviceID: "tap1",
			giveCIDR:     "10.0.0.0/16",
			giveGW:       "10.0.01",
			giveStatus:   common_pb.Status_DEPLOYING,

			expRes: nil,
			expErr: evmError{errors.New("Invalid GW address"), "10.0.01"},
			pass:   false,
		},
	}
	errorTest := evmError{errors.New("this is an error"), "hi"}
	t.Log(errorTest)
	for _, tt := range tests {
		t.Run(tt.giveName, func(t *testing.T) {

			vm, err := NewEvm(
				tt.giveName,
				tt.giveIP,
				tt.giveMAC,
				tt.giveRemoteID,
				tt.giveDeviceID,
				tt.giveCIDR,
				tt.giveGW,
				tt.giveStatus)
			assert.Equal(t, tt.expRes, vm)
			if tt.pass {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tt.expErr.Error(), err.Error())
			}
		})
	}
}

func TestCreateMinimalPort(t *testing.T) {
	tests := []struct {
		in      *pb.InternalPortConfig
		server  *httptest.Server
		metrics metrics.Metrics
		url     string
		expResp *AlcorEvm
		expErr  error
	}{
		{
			in: &pb.InternalPortConfig{
				Name:      "vm1",
				Vpcid:     "10101010",
				Tenantid:  "1234567",
				Sg:        "12323213",
				Projectid: "1234567",
				Subnetid:  "1221",
				Cidr:      "10.0.0.0/16",
				Gw:        "10.0.0.1",
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"port":{"id":"123456789101112","mac_address":"00:00:00:00:00:01","fixed_ips":[{"ip_address":"10.0.0.2"}}`))
			})),
			metrics: &mockMetrics{
				ServiceName:     "fake",
				OpsTotalLatency: nil,
				OpsSuccess:      nil,
				OpsFail:         nil,
			},
			expResp: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				mac:      "00:00:00:00:00:01",
				remoteID: "123456789101112",
				deviceID: "tap12345678910",
				cidr:     "10.0.0.0/16",
				gw:       "10.0.0.1",
				status:   common_pb.Status_DEPLOYING,
			},
			expErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.in.Name, func(t *testing.T) {
			defer tt.server.Close()
			vm, err := CreateMinimalPort(
				tt.server.URL,
				tt.in,
				tt.metrics,
			)
			if !reflect.DeepEqual(tt.expResp, vm) {
				t.Error("Expected", tt.expResp, "got", vm)
			}
			if err != nil {
				assert.Equal(t, tt.expErr.Error(), err.Error())
			}
		})
	}
}

func TestUpdatePort(t *testing.T) {
	evm1, err := NewEvm(
		"vm1",
		"10.0.0.2",
		"00:00:00:00:00:01",
		"123456789101112",
		"tap12345678910",
		"10.0.0.0/16",
		"10.0.0.1",
		common_pb.Status_DEPLOYING,
	)
	assert.Nil(t, err)
	tests := []struct {
		in      *pb.InternalPortConfig
		server  *httptest.Server
		metrics metrics.Metrics
		url     string
		evm     Evm

		expErr error
	}{
		{
			in: &pb.InternalPortConfig{
				Name:      "vm1",
				Vpcid:     "10101010",
				Tenantid:  "1234567",
				Sg:        "12323213",
				Projectid: "1234567",
				Subnetid:  "1221",
				Cidr:      "10.0.0.0/16",
				Gw:        "10.0.0.1",
				Hostname:  "10.0.0.2",
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"port":{"id":"123456789101112","mac_address":"00:00:00:00:00:01","fixed_ips":[{"ip_address":"10.0.0.2"}}`))
			})),
			metrics: &mockMetrics{
				ServiceName:     "fake",
				OpsTotalLatency: nil,
				OpsSuccess:      nil,
				OpsFail:         nil,
			},
			evm:    evm1,
			expErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.in.Name, func(t *testing.T) {
			defer tt.server.Close()
			err := UpdatePort(
				tt.server.URL,
				tt.in,
				tt.metrics,
				tt.evm,
			)
			assert.Equal(t, tt.expErr, err)
		})
	}
}

func TestDeletePort(t *testing.T) {
	tests := []struct {
		in      *pb.InternalPortConfig
		server  *httptest.Server
		metrics metrics.Metrics
		url     string
		evm     *AlcorEvm

		expErr error
	}{
		{
			in: &pb.InternalPortConfig{
				Name:      "vm1",
				Vpcid:     "10101010",
				Tenantid:  "1234567",
				Sg:        "12323213",
				Projectid: "1234567",
				Subnetid:  "1221",
				Cidr:      "10.0.0.0/16",
				Gw:        "10.0.0.1",
				Hostname:  "10.0.0.2",
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"port":{"id":"123456789101112","mac_address":"00:00:00:00:00:01","fixed_ips":[{"ip_address":"10.0.0.2"}}`))
			})),
			metrics: &mockMetrics{
				ServiceName:     "fake",
				OpsTotalLatency: nil,
				OpsSuccess:      nil,
				OpsFail:         nil,
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				mac:      "00:00:00:00:00:01",
				remoteID: "123456789101112",
				deviceID: "tap12345678910",
				cidr:     "10.0.0.0/16",
				gw:       "10.0.0.1",
				status:   common_pb.Status_DEPLOYING,
			},
			expErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.in.Name, func(t *testing.T) {
			defer tt.server.Close()
			err := DeletePort(
				tt.server.URL,
				tt.in,
				tt.metrics,
				tt.evm,
			)
			assert.Equal(t, tt.expErr, err)
		})
	}
}

func TestCreateStandaloneDevice(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip tuntap add mode tap tap1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip tuntap add mode tap tap1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap2",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.CreateStandaloneDevice(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.CreateStandaloneDevice(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestDeleteStandaloneDevice(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip tuntap del mode tap tap1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip tuntap del mode tap tap1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap2",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.DeleteStandaloneDevice(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.DeleteStandaloneDevice(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}

}
func TestCreateNamespace(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns add vm1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns add vm1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap2",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.CreateNamespace(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.CreateNamespace(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestDeleteNamespace(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns delete vm1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns delete vm1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap2",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.DeleteNamespace(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.DeleteNamespace(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestMoveDeviceToNamespace(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip link set tap1 netns vm1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip link set tap1 netns vm1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap2",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.MoveDeviceToNetns(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.MoveDeviceToNetns(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestMoveDeviceRootNamespace(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip link set tap1 netns 1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip link set tap1 netns 1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap2",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.MoveDeviceToRootNetns(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.MoveDeviceToRootNetns(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestAssignIP(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip addr add 10.0.0.2/16 dev tap1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip addr add 10.0.0..2/16 dev tap1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.AssignIP(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.AssignIP(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestSetMTUProbing(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 sysctl -w net.ipv4.tcp_mtu_probing=2" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 sysctl -w net.ipv4.tcp_mtu_probing=2" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.SetMTUProbing(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.SetMTUProbing(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestBringLoUp(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip link set dev lo up" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip link set dev lo up" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.BringLoUp(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.BringLoUp(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestAssignMac(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip link set dev tap1 address 00:00:00:00:00:01" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
				mac:      "00:00:00:00:00:01",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip link set dev tap1 address 00:00:00:00:00:01" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.AssignMac(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.AssignMac(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestAddGateway(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip r add default via 10.0.0.1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
				mac:      "00:00:00:00:00:01",
				gw:       "10.0.0.1",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip r add default via 10.0.0.1" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
				mac:      "00:00:00:00:00:01",
				gw:       "10.0.0.1",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.AddGateway(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.AddGateway(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestBringDeviceUp(t *testing.T) {
	metrics := mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		BashExec func(cmd string) ([]byte, error)
		evm      *AlcorEvm
		expErr   error
		pass     bool
	}{
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip link set dev tap1 up" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm1",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
				mac:      "00:00:00:00:00:01",
				gw:       "10.0.0.1",
			},
			expErr: nil,
			pass:   true,
		},
		{
			BashExec: func(cmd string) ([]byte, error) {
				if cmd != "ip netns exec vm1 ip link set dev tap1 up" {
					return nil, errors.New("Bad")
				}
				return []byte("tap1"), nil
			},
			evm: &AlcorEvm{
				name:     "vm2",
				ip:       "10.0.0.2",
				deviceID: "tap1",
				cidr:     "10.0.0.0/16",
				mac:      "00:00:00:00:00:01",
				gw:       "10.0.0.1",
			},
			expErr: errors.New("Bad"),
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.deviceID, func(t *testing.T) {
			BashExec = tt.BashExec
			if tt.pass {
				err := tt.evm.BringDeviceUp(&metrics)
				assert.Nil(t, err)
			} else {
				err := tt.evm.BringDeviceUp(&metrics)
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestGetName(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
	}
	assert.Equal(t, evm.GetName(), "vm2")
}

func TestGetIP(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
	}
	assert.Equal(t, evm.GetIP(), "10.0.0.2")
}

func TestGetMac(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
	}
	assert.Equal(t, evm.GetMac(), "00:00:00:00:00:01")
}

func TestGetCidr(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
	}
	assert.Equal(t, evm.GetCidr(), "10.0.0.0/16")
}

func TestGetGw(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
	}
	assert.Equal(t, evm.GetGw(), "10.0.0.1")
}

func TestGetDeviceId(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
	}
	assert.Equal(t, evm.GetDeviceId(), "tap1")
}

func TestGetRemoteId(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
		remoteID: "123",
	}
	assert.Equal(t, evm.GetRemoteId(), "123")
}

func TestGetStatus(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
		remoteID: "123",
		status:   common_pb.Status_DEPLOYING,
	}
	assert.Equal(t, evm.GetStatus(), common_pb.Status_DEPLOYING)
}

func TestSetName(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
		remoteID: "123",
		status:   common_pb.Status_DEPLOYING,
	}
	evm.SetName("vm3")
	assert.Equal(t, evm.GetName(), "vm3")
}

func TestSetIP(t *testing.T) {
	tests := []struct {
		evm    *AlcorEvm
		ip     string
		expErr error
		pass   bool
	}{
		{
			evm:    &AlcorEvm{},
			ip:     "10.0.0.2",
			expErr: nil,
			pass:   true,
		},
		{
			evm:    &AlcorEvm{},
			ip:     "123",
			expErr: evmError{errors.New("Invalid IP Address"), "123"},
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.ip, func(t *testing.T) {
			err := tt.evm.SetIP(tt.ip)
			if tt.pass {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestSetMac(t *testing.T) {
	tests := []struct {
		evm    *AlcorEvm
		mac    string
		expErr error
		pass   bool
	}{
		{
			evm:    &AlcorEvm{},
			mac:    "00:00:00:00:00:01",
			expErr: nil,
			pass:   true,
		},
		{
			evm:    &AlcorEvm{},
			mac:    "123",
			expErr: evmError{errors.New("Invalid MAC address"), "123"},
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.ip, func(t *testing.T) {
			err := tt.evm.SetMac(tt.mac)
			if tt.pass {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestSetCidr(t *testing.T) {
	tests := []struct {
		evm    *AlcorEvm
		mac    string
		expErr error
		pass   bool
	}{
		{
			evm:    &AlcorEvm{},
			mac:    "10.0.0.0/16",
			expErr: nil,
			pass:   true,
		},
		{
			evm:    &AlcorEvm{},
			mac:    "123",
			expErr: evmError{errors.New("Invalid CIDR address"), "123"},
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.ip, func(t *testing.T) {
			err := tt.evm.SetCidr(tt.mac)
			if tt.pass {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestSetGw(t *testing.T) {
	tests := []struct {
		evm    *AlcorEvm
		mac    string
		expErr error
		pass   bool
	}{
		{
			evm:    &AlcorEvm{},
			mac:    "10.0.0.1",
			expErr: nil,
			pass:   true,
		},
		{
			evm:    &AlcorEvm{},
			mac:    "123",
			expErr: evmError{errors.New("Invalid GW address"), "123"},
			pass:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.evm.ip, func(t *testing.T) {
			err := tt.evm.SetGw(tt.mac)
			if tt.pass {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err.Error(), tt.expErr.Error())
			}
		})
	}
}

func TestSetDeviceId(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
		remoteID: "123",
		status:   common_pb.Status_DEPLOYING,
	}
	evm.SetDeviceId("123123")
	assert.Equal(t, evm.GetDeviceId(), "123123")
}

func TestSetRemoteId(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
		remoteID: "123",
		status:   common_pb.Status_DEPLOYING,
	}
	evm.SetRemoteId("abcd")
	assert.Equal(t, evm.GetRemoteId(), "abcd")
}

func TestSetStatus(t *testing.T) {
	evm := AlcorEvm{
		name:     "vm2",
		ip:       "10.0.0.2",
		deviceID: "tap1",
		cidr:     "10.0.0.0/16",
		mac:      "00:00:00:00:00:01",
		gw:       "10.0.0.1",
		remoteID: "123",
		status:   common_pb.Status_DEPLOYING,
	}
	evm.SetStatus(common_pb.Status_DONE)
	assert.Equal(t, evm.GetStatus(), common_pb.Status_DONE)
}
