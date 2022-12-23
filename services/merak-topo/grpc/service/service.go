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

	utils.Logger.Info("Received on TopologyHandler %s", in)

	k8client, err := utils.K8sClient()
	if err != nil {
		utils.Logger.Error("create k8s client error %s", err.Error())
		return &returnMessage, err
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		utils.Logger.Error("connect to DB error %s", err1.Error())
		return &returnMessage, err1
	}

	// Operation&Return
	switch op := in.OperationType; op {

	case pb_common.OperationType_INFO:

		if in.Config.GetTopologyId() != "" {
			err_info := handler.Info(k8client, in.Config.GetTopologyId(), &returnMessage)
			if err_info != nil {
				utils.Logger.Error("INFO can not get the topology information %s", err_info.Error())
				returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
				returnMessage.ReturnMessage = "INFO fails."

			} else {
				utils.Logger.Info("INFO get the topology information")
				returnMessage.ReturnCode = pb_common.ReturnCode_OK
				returnMessage.ReturnMessage = "INFO successes."
			}

			utils.Logger.Debug("return message %v", returnMessage.ReturnMessage)
			utils.Logger.Debug("return code %v", returnMessage.ReturnCode)
			utils.Logger.Debug("return compute node %+v", returnMessage.ComputeNodes)
			utils.Logger.Debug("return host node %v", returnMessage.Hosts)
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

		/*comment gw creation function*/
		// if data_plane_cidr == "" || aca_num == 0 || aca_per_rack == 0 || rack_num == 0 || cgw_num == 0 {
		if data_plane_cidr == "" || aca_num == 0 || aca_per_rack == 0 || rack_num == 0 || ports_per_vswitch == 0 {

			utils.Logger.Info("Invalid input info for valid data plane cider, aca number, aca per rack number, rack number, ports per vswitch, and control plane gateways")

			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			returnMessage.ReturnMessage = "Must provide a valid data plane cider, aca number, aca per rack number, rack number, ports per vswitch, and control plane gateways"

			return &returnMessage, nil

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
			err_create := handler.Create(k8client, topo_id, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), uint32(cgw_num), data_plane_cidr, uint32(ports_per_vswitch), images, aca_parameters, &returnMessage)

			if err_create != nil {
				utils.Logger.Error("Fail to create topology %s", err_create.Error())
				returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
				returnMessage.ReturnMessage = "Fail to Create Topology."
			} else {
				returnMessage.ReturnCode = pb_common.ReturnCode_OK
				returnMessage.ReturnMessage = "Success to create topology"
			}
			utils.Logger.Debug("return message %v", returnMessage.ReturnMessage)
			utils.Logger.Debug("return code %v", returnMessage.ReturnCode)
			utils.Logger.Debug("return compute node %+v", returnMessage.ComputeNodes)
			utils.Logger.Debug("return host node %v", returnMessage.Hosts)

			return &returnMessage, err_create
		}

	case pb_common.OperationType_DELETE:
		// delete topology
		err := handler.Delete(k8client, in.Config.TopologyId, &returnMessage)

		//return topology message-- compute info

		if err != nil {
			utils.Logger.Error("Delete topology - fail %s", err.Error())
			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			returnMessage.ReturnMessage = "Fail to Delete Topology."
		} else {
			utils.Logger.Info("Delete topology - success")
			returnMessage.ReturnCode = pb_common.ReturnCode_OK
			returnMessage.ReturnMessage = "Success to Delete Topology"
		}

		utils.Logger.Debug("return message %v", returnMessage.ReturnMessage)
		utils.Logger.Debug("return code %v", returnMessage.ReturnCode)

	case pb_common.OperationType_UPDATE:
		// update topology
	default:
		utils.Logger.Info("Unknown Operation")
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		returnMessage.ReturnMessage = "TopologyHandler: Unknown Operation"
	}

	return &returnMessage, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	utils.Logger.Info("Received on TopologyHandler %s", in)
	var returnMessage pb.ReturnTopologyMessage
	return &returnMessage, nil
}
