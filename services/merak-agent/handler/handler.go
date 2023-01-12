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
package handler

import (
	"context"
	"errors"
	"os"
	"strconv"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/common/logger"
	"github.com/futurewei-cloud/merak/services/common/metrics"
	merakEvm "github.com/futurewei-cloud/merak/services/merak-agent/evm"
	"github.com/prometheus/client_golang/prometheus"
)

type Server struct {
	pb.UnimplementedMerakAgentServiceServer
}

var (
	RemoteServer       string
	MerakMetrics       metrics.Metrics
	PrometheusRegistry *prometheus.Registry
)

var MerakLogger *logger.MerakLog

func (s *Server) PortHandler(ctx context.Context, in *pb.InternalPortConfig) (*pb.AgentReturnInfo, error) {
	MerakLogger.Info("Received on PortHandler", in)
	MerakLogger.Info("Received on PortHandler OP", in.OperationType)
	switch op := in.OperationType; op {
	case common_pb.OperationType_PRECREATE:
		MerakLogger.Info("Operation Create Minimal Port")
		createMinimalPortUrl := "http://" + RemoteServer + ":" + strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT) + "/project/" + in.Projectid + "/ports"
		return caseCreateMinimal(ctx, in, createMinimalPortUrl)

	case common_pb.OperationType_CREATE:
		MerakLogger.Info("Operation Create")
		updatePortUrl := "http://" + RemoteServer + ":" + strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT) + "/project/" + in.Projectid + "/ports/"
		return caseCreate(ctx, in, updatePortUrl)

	case common_pb.OperationType_UPDATE:
		MerakLogger.Info("Update Unimplemented")
		return &pb.AgentReturnInfo{
			ReturnMessage: "Update Unimplemented",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          nil,
		}, errors.New("update unimplemented")

	case common_pb.OperationType_DELETE:
		MerakLogger.Info("Operation Delete")
		deletePortUrl := "http://" + RemoteServer + ":" + strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT) + "/project/" + in.Projectid + "/ports/"
		return caseDelete(ctx, in, deletePortUrl)

	default:
		MerakLogger.Info("Unknown Operation")
		return &pb.AgentReturnInfo{
			ReturnMessage: "Unknown Operation",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, errors.New("unknown operation")
	}
}

func (s *Server) BulkPortAdd(ctx context.Context, in *pb.BulkPorts) (*pb.AgentReturnInfo, error) {
	_, ok := os.LookupEnv(constants.MODE_ENV)
	if !ok {
		MerakLogger.Info("Operation Bulk Port add")
		return &pb.AgentReturnInfo{}, merakEvm.Ovsdbbulk(in.Tapnames, MerakMetrics)
	} else {
		return &pb.AgentReturnInfo{}, nil
	}
}
