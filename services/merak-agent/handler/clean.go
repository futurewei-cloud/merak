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
	"log"
	"os/exec"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
)

func caseClean(ctx context.Context, in *pb.InternalPortConfig) (*pb.AgentReturnInfo, error) {

	log.Println("Deleting all namespace")
	cmd := exec.Command("bash", "-c", "for name in $(ip netns ls); do ip netns delete $name; done")
	stdout, err := cmd.Output()
	if err != nil {
		log.Println("Namespace deletion failed! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Namespace deletion failed! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	log.Println("Deleting all tap devices")
	cmd = exec.Command("bash", "-c", "for name in $(ovs-vsctl list-ports br-int | grep tap); do ovs-vsctl del-port br-int $name; done")
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to delete all tap devices " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to delete tap " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	//ip l | grep -Po "bridgev\d+:"
	log.Println("Deleting all bridge devices")
	cmd = exec.Command("bash", "-c", "for name in $( ip l | grep -Po 'bridgev\\d+:'); do ip l delete ${name::-1} ; done")
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Bridge deletion failed! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Bridge deletion failed! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}
	return &pb.AgentReturnInfo{
		ReturnMessage: "Delete Success!",
		ReturnCode:    common_pb.ReturnCode_OK,
	}, nil

}
