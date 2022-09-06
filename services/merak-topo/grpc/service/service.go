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
	"fmt"
	"log"

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
	log.Printf("Received on TopologyHandler %s", in)

	var returnMessage pb.ReturnTopologyMessage

	k8client, err := utils.K8sClient()
	if err != nil {
		return nil, fmt.Errorf("create k8s client error %s", err.Error())
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		fmt.Printf("connect to DB error %s", err1)
	}

	// Operation&Return
	switch op := in.OperationType; op {

	case pb_common.OperationType_INFO:

		if in.Config.GetTopologyId() != "" {
			err_info := handler.Info(k8client, in.Config.GetTopologyId(), &returnMessage)
			if err_info != nil {
				returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
				returnMessage.ReturnMessage = "INFO fails."

			} else {
				returnMessage.ReturnCode = pb_common.ReturnCode_OK
				returnMessage.ReturnMessage = "INFO successes."
			}

			log.Printf("return message %v", returnMessage.ReturnMessage)
			log.Printf("return code %v", returnMessage.ReturnCode)
			log.Printf("return compute node %+v", returnMessage.ComputeNodes)
			log.Printf("return host node %v", returnMessage.Hosts)
		}

	case pb_common.OperationType_CREATE:

		aca_num := in.Config.GetNumberOfVhosts()
		rack_num := in.Config.GetNumberOfRacks()
		aca_per_rack := in.Config.GetVhostPerRack()
		data_plane_cidr := in.Config.GetDataPlaneCidr()
		// cgw_num := in.Config.GetNumberOfGateways()  /*comment gw creation function, set cgw_num=0*/
		cgw_num := 0
		topo_id := in.Config.GetTopologyId()

		ports_per_vswitch := in.Config.GetPortsPerVswitch()

		/*comment gw creation function*/
		// if data_plane_cidr == "" || aca_num == 0 || aca_per_rack == 0 || rack_num == 0 || cgw_num == 0 {
		if data_plane_cidr == "" || aca_num == 0 || aca_per_rack == 0 || rack_num == 0 {

			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			returnMessage.ReturnMessage = "Must provide a valid data plane cider, aca number, aca per rack number, rack number, and control plane gateways"

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
			// err_create := handler.Create(k8client, topo_id, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), uint32(cgw_num), data_plane_cidr, &returnMessage)

			err_create := handler.Create(k8client, topo_id, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), uint32(cgw_num), data_plane_cidr, uint32(ports_per_vswitch), &returnMessage)

			if err_create != nil {
				returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
				returnMessage.ReturnMessage = "Fail to Create Topology."
			} else {
				returnMessage.ReturnCode = pb_common.ReturnCode_OK
				returnMessage.ReturnMessage = "Success to create topology"
			}
			log.Printf("return message %v", returnMessage.ReturnMessage)
			log.Printf("return code %v", returnMessage.ReturnCode)
			log.Printf("return compute node %+v", returnMessage.ComputeNodes)
			log.Printf("return host node %v", returnMessage.Hosts)

			return &returnMessage, err_create
		}

	case pb_common.OperationType_DELETE:
		// delete topology
		err := handler.Delete(k8client, in.Config.TopologyId)

		//return topology message-- compute info

		if err != nil {
			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			returnMessage.ReturnMessage = "Fail to Delete Topology."
		} else {
			returnMessage.ReturnCode = pb_common.ReturnCode_OK
			returnMessage.ReturnMessage = "Success to Delete Topology"
		}

		log.Printf("return message %v", returnMessage.ReturnMessage)
		log.Printf("return code %v", returnMessage.ReturnCode)

	case pb_common.OperationType_UPDATE:
		// update topology
	default:
		log.Println("Unknown Operation")
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		returnMessage.ReturnMessage = "TopologyHandler: Unknown Operation"
	}

	return &returnMessage, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	log.Printf("Received on TopologyHandler %s", in)
	var returnMessage pb.ReturnTopologyMessage
	return &returnMessage, nil
}
