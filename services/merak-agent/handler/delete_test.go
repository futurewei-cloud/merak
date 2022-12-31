package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	"github.com/futurewei-cloud/merak/services/merak-agent/evm"
	"github.com/stretchr/testify/assert"
)

func TestCaseDelete(t *testing.T) {
	metrics := &mockMetrics{
		ServiceName:     "fake",
		OpsTotalLatency: nil,
		OpsSuccess:      nil,
		OpsFail:         nil,
	}
	tests := []struct {
		bashExec func(cmd string) ([]byte, error)
		server   *httptest.Server
		in       *pb.InternalPortConfig
		expErr   error
		pass     bool
	}{
		{
			bashExec: func(cmd string) ([]byte, error) {
				return []byte("tap1"), nil
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"port":{"id":"","mac_address":"00:00:00:00:00:01","fixed_ips":[{"ip_address":"10.0.0.2"}}`))
			})),
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.in.Name, func(t *testing.T) {
			defer tt.server.Close()
			evm.BashExec = tt.bashExec
			MerakMetrics = metrics
			_, err := caseDelete(
				context.Background(),
				tt.in,
				tt.server.URL+"",
			)
			assert.Nil(t, err)
		})
	}
}
