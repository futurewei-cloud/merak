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
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	constants "github.com/futurewei-cloud/merak/services/common"
)

func caseDelete(ctx context.Context, in *pb.InternalPortConfig) (*pb.AgentReturnInfo, error) {
	log.Println("Send Delete Port Request to Alcor")

	req, err := http.NewRequest(http.MethodDelete, "http://"+constants.ALCOR_ADDRESS+":"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports/"+in.Remoteid, bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		log.Println("Failed send Delete Port request to Alcor!", err)
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed send Delete Port request to Alcor!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	log.Println("Sending delete port request to Alcor", req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to delete port to Alcor!")
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed Delete port!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}
	log.Println("Response code from Alcor", resp.StatusCode)
	if resp.StatusCode != constants.HTTP_OK {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to Delete Port ! Response Code: " + strconv.Itoa(resp.StatusCode),
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, errors.New("Failed to delete port! Response Code: " + strconv.Itoa(resp.StatusCode))
	}

	log.Println("Deleting Namespace")
	cmd := exec.Command("bash", "-c", "ip netns delete "+in.Name)
	stdout, err := cmd.Output()
	if err != nil {
		log.Println("Namespace deletion failed! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Namespace deletion failed! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}
	log.Println("Deleting bridge device")
	cmd = exec.Command("bash", "-c", "ip link delete bridge"+in.Name)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Bridge deletion failed! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Bridge deletion failed! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	tapName := "tap" + in.Remoteid[:11]
	log.Println("Deleting TAP device " + tapName)
	cmd = exec.Command("bash", "-c", "ovs-vsctl del-port br-int "+tapName)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to delete tap " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to delete tap " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	return &pb.AgentReturnInfo{
		ReturnMessage: "Delete Success!",
		ReturnCode:    common_pb.ReturnCode_OK,
	}, nil

}
