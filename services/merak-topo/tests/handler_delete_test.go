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

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	"github.com/futurewei-cloud/merak/services/merak-topo/handler"
	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
)

func TestTopologyDelete(t *testing.T) {

	returnMessage := pb.ReturnTopologyMessage{
		ReturnCode:    common_pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}

	topo_id := "1topo"

	k8client, err := utils.K8sClient()
	if err != nil {
		log.Printf("create k8s client error %s", err)
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		log.Printf("connect to DB error %s", err1)
	}

	err4 := handler.Delete(k8client, topo_id)
	if err4 != nil {
		returnMessage.ReturnCode = common_pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Fail to Delete Topology."
	} else {
		returnMessage.ReturnCode = common_pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Topology Delete Done."
	}

	log.Printf("///// DELETE Return Message //// %v", &returnMessage)

}
