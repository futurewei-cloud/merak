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

package tests

import (
	"log"
	"testing"

	pb_common "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	"github.com/futurewei-cloud/merak/services/merak-topo/handler"
	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
)

func TestTopologyCreate(t *testing.T) {

	returnMessage := pb.ReturnTopologyMessage{
		ReturnCode:    pb_common.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
	aca_num := 48
	rack_num := 4
	aca_per_rack := 12
	data_plane_cidr := "10.200.0.0/16"
	topo_id := "1topo"
	cgw_num := 6
	ports_per_vswitch := 2
	var imgs []*pb.InternalTopologyImage
	var aca_parameters string
	topoPrefix := "1topo"
	namespace := "default"

	k8client, err := utils.K8sClient()
	if err != nil {
		log.Printf("create k8s client error %s", err)
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		log.Printf("connect to DB error %s", err1)
	}

	err2 := handler.Create(k8client, topo_id, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), uint32(cgw_num), data_plane_cidr, uint32(ports_per_vswitch), imgs, aca_parameters, &returnMessage, topoPrefix, namespace)
	if err2 != nil {
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Fail to Create Topology."

	} else {
		returnMessage.ReturnCode = pb_common.ReturnCode_OK
		returnMessage.ReturnMessage = "Topology Deployed."

	}

	log.Printf("///// CREATE Return Message //// %v", &returnMessage)

}
