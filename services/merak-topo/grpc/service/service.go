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

package service

import (
	"context"
	"errors"
	"flag"
	"strings"

	pb_common "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	"github.com/futurewei-cloud/merak/services/merak-topo/handler"
	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
)

var (
	Port = flag.Int("port", constants.TOPLOGY_GRPC_SERVER_PORT, "The server port")
)

type Server struct {
	pb.MerakTopologyServiceServer
}

func (s *Server) TopologyHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	var returnMessage pb.ReturnTopologyMessage
	err_flag := 0
	errs := errors.New("merak-topo can't handle this request")
	var err_return error

	topoPrefix := in.Config.GetTopologyId()[:5]

	/*comment: create topology in the default namespace*/
	// namespace := "default"
	namespace := "merak-" + topoPrefix

	utils.Logger.Info("Received request from Scenario Manager", "request", in)

	k8client, err := utils.K8sClient()
	if err != nil {
		utils.Logger.Error("k8s client", "configuration", err.Error())
		return &returnMessage, errs
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		utils.Logger.Error("redis database", "connect to DB", err1.Error())
		return &returnMessage, errs
	}

	// Operation&Return
	switch op := in.OperationType; op {

	case pb_common.OperationType_INFO:

		if in.Config.GetTopologyId() != "" {
			err_info := handler.Info(k8client, in.Config.GetTopologyId(), &returnMessage, topoPrefix, namespace)
			if err_info != nil {
				utils.Logger.Info("topology information is not ready yet", in.Config.GetTopologyId(), err_info.Error())
				err_flag = 1
				returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
				returnMessage.ReturnMessage = "CHECK fail."
			} else {
				returnMessage.ReturnCode = pb_common.ReturnCode_OK
				returnMessage.ReturnMessage = "CHECK success."
			}

			utils.Logger.Debug("requrest CHECK details", "return code", returnMessage.ReturnCode, "return compute node", returnMessage.ComputeNodes, "return host node", returnMessage.Hosts)

		}

	case pb_common.OperationType_CREATE:

		aca_num := in.Config.GetNumberOfVhosts()
		rack_num := in.Config.GetNumberOfRacks()
		aca_per_rack := in.Config.GetVhostPerRack()
		data_plane_cidr := in.Config.GetDataPlaneCidr()
		// cgw_num := in.Config.GetNumberOfGateways()  /*comment gw creation function, set cgw_num=0*/
		cgw_num := 0
		topo_id := in.Config.GetTopologyId()
		images := in.Config.GetImages()

		ports_per_vswitch := in.Config.GetPortsPerVswitch()

		aca_parameters := ""
		service_config := in.Config.GetServices()
		for _, service := range service_config {
			if service.Name == "aca-cmd" {
				aca_ip := ""
				aca_port := ""
				for _, par := range service.Parameters {
					words := strings.Fields(par)
					if words[0] == "-a" {
						aca_ip = words[1]
					} else if words[0] == "-p" {
						aca_port = words[1]
					}
				}
				aca_parameters = aca_ip + " " + aca_port
			}
		}

		if data_plane_cidr == "" || aca_num == 0 || aca_per_rack == 0 || rack_num == 0 || ports_per_vswitch == 0 {

			utils.Logger.Error("request DEPLOY", "Invalid input info", "check data plane cider, aca number, aca per rack number, rack number, ports per vswitch, and control plane gateways")

			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			returnMessage.ReturnMessage = "Must provide a valid data plane cider, aca number, aca per rack number, rack number, ports per vswitch, and control plane gateways"

			return &returnMessage, errs

		}

		switch s := in.Config.TopologyType; s {
		case pb.TopologyType_SINGLE:
		//
		case pb.TopologyType_LINEAR:
			//
		case pb.TopologyType_MESH:
			//
		case pb.TopologyType_CUSTOM:
			//
		case pb.TopologyType_REVERSED:
			//
		default:
			// pb.TopologyType_TREE
			err_create := handler.Create(k8client, topo_id, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), uint32(cgw_num), data_plane_cidr, uint32(ports_per_vswitch), images, aca_parameters, &returnMessage, topoPrefix, namespace)

			if err_create != nil {
				utils.Logger.Error("can't deploy topology", topo_id, err_create.Error())
				err_flag = 1
				returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
				returnMessage.ReturnMessage = "DEPLOY fail."
			} else {
				utils.Logger.Info("request DEPLOY", topo_id, "success")
				returnMessage.ReturnCode = pb_common.ReturnCode_OK
				returnMessage.ReturnMessage = "DEPLOY Success"
			}

			utils.Logger.Debug("requrest DEPLOY details", "return code", returnMessage.ReturnCode, "return compute node", returnMessage.ComputeNodes, "return host node", returnMessage.Hosts)

			return &returnMessage, err_create
		}

	case pb_common.OperationType_DELETE:
		// delete topology
		err := handler.Delete(k8client, in.Config.TopologyId, &returnMessage, topoPrefix, namespace)

		//return topology message-- compute info

		if err != nil {
			utils.Logger.Error("request DELETE", in.Config.TopologyId, err.Error())
			err_flag = 1
			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			returnMessage.ReturnMessage = "DELETE fail"
		} else {
			utils.Logger.Info("request DELETE", in.Config.TopologyId, "success")
			returnMessage.ReturnCode = pb_common.ReturnCode_OK
			returnMessage.ReturnMessage = "DELETE success"
		}

		utils.Logger.Debug("request DELETE", "return message", returnMessage.ReturnMessage, "return code", returnMessage.ReturnCode)

	case pb_common.OperationType_UPDATE:
		// update topology
	default:
		utils.Logger.Info("Unknown Operation", in.Config.TopologyId, "please check the input")
		err_flag = 1
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Unknown operation, please retry"
	}

	if err_flag == 1 {
		err_return = errs
		err_flag = 0
	} else {
		err_return = nil
	}
	return &returnMessage, err_return

}
