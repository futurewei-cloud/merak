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

package activities

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	"github.com/futurewei-cloud/merak/services/merak-network/entities"
	"github.com/futurewei-cloud/merak/services/merak-network/http"
	"github.com/futurewei-cloud/merak/services/merak-network/utils"
)

func RegisterNode(ctx context.Context, compute []*common_pb.InternalComputeInfo, wg *sync.WaitGroup, projectId string) (string, error) {
	log.Println("RegisterNode")
	//defer wg.Done()
	log.Printf("compute %s", compute)
	nodeInfo := entities.NodeStruct{}

	for _, host := range compute {
		log.Printf("host %s", host)
		nodeBody := entities.NodeBody{
			LocalIP:    host.DatapathIp,
			MacAddress: host.Mac,
			NodeID:     host.Id,
			NodeName:   host.Name,
			ServerPort: 50001,
			Veth:       host.Veth,
		}
		nodeInfo.Hosts = append(nodeInfo.Hosts, nodeBody)
	}
	log.Printf("nodeInfo: %s", nodeInfo)
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30007/nodes/bulk", "POST", nodeInfo, nil)
	if returnErr != nil {
		log.Printf("returnErr %s", returnErr)
		return "", returnErr
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.NodeReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("RegisterNode done")
	return "", nil
}
