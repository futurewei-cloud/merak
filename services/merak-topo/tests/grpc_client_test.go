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
	"context"
	"flag"
	"strconv"
	"strings"
	"testing"
	"time"

	pb_common "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:40052", "the address to connect to")
	ctx  = context.Background()
)

//grpc test
func TestGrpcClient(t *testing.T) {
	var topology_address strings.Builder
	topology_address.WriteString(constants.TOPLOGY_GRPC_SERVER_ADDRESS)
	topology_address.WriteString(":")
	topology_address.WriteString(strconv.Itoa(constants.TOPLOGY_GRPC_SERVER_PORT))

	node0 := pb.InternalVNodeInfo{
		OperationType: pb_common.OperationType_CREATE,
		Id:            "0",
		Name:          "proj1-topo1-vnode0",
		Type:          pb.VNodeType_VHOST,
		Vnics:         []*pb.InternalVNicInfo{},
	}
	node1 := pb.InternalVNodeInfo{
		OperationType: pb_common.OperationType_CREATE,
		Id:            "1",
		Name:          "proj1-topo1-vnode1",
		Type:          pb.VNodeType_VSWITCH,
		Vnics:         []*pb.InternalVNicInfo{},
	}

	link0 := pb.InternalVLinkInfo{
		OperationType: pb_common.OperationType_CREATE,
		Id:            "0",
		Name:          "link_0",
		Src:           "10.0.0.1",
		Dst:           "10.0.0.2",
	}

	link1 := pb.InternalVLinkInfo{
		OperationType: pb_common.OperationType_CREATE,
		Id:            "1",
		Name:          "link_1",
		Src:           "10.0.0.2",
		Dst:           "10.0.0.3",
	}

	topologyConfig_c1 := pb.InternalTopologyConfiguration{
		FormatVersion:  1,
		RevisionNumber: 1,
		RequestId:      "proj1-topo1-info-test",
		TopologyId:     "proj1-topo1",
		MessageType:    pb_common.MessageType_FULL,
		Vnodes:         []*pb.InternalVNodeInfo{&node0, &node1},
		Vlinks:         []*pb.InternalVLinkInfo{&link0, &link1},
		ExtraInfo:      &pb.InternalTopologyExtraInfo{Info: "info test"},
	}

	topologyConfig_c2 := pb.InternalTopologyConfiguration{
		FormatVersion:  1,
		RevisionNumber: 1,
		RequestId:      "proj1-topo2-create-test",
		TopologyId:     "proj1-topo2",
		MessageType:    pb_common.MessageType_FULL,
		Vnodes:         []*pb.InternalVNodeInfo{},
		Vlinks:         []*pb.InternalVLinkInfo{},
		ExtraInfo:      &pb.InternalTopologyExtraInfo{Info: "create test"},
	}
	// Test cases for INFO, CREATE, DELETE, UPDATE
	topology_info := pb.InternalTopologyInfo{
		OperationType: pb_common.OperationType_INFO,
		Config:        &topologyConfig_c1,
	}

	topology_create := pb.InternalTopologyInfo{
		OperationType: pb_common.OperationType_CREATE,
		Config:        &topologyConfig_c2,
	}

	conn, _ := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()

	c := pb.NewMerakTopologyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err1 := c.TopologyHandler(ctx, &topology_create)
	_, err2 := c.TopologyHandler(ctx, &topology_info)

	assert.Equal(t, err1, err2)
}
