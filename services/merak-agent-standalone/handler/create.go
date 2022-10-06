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
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
)

func caseCreate(ctx context.Context, in *pb.InternalPortConfig) (*pb.AgentReturnInfo, error) {

	log.Println("Create Minimal Port")
	vmInfo := pb.ReturnPortInfo{
		Ip:       "",
		Deviceid: "",
		Remoteid: "",
		Status:   common_pb.Status_ERROR,
	}

	ip := "10.0.0.2"
	mac := "aa:bb:cc:dd:ee:ff"
	portID := "NO ALCOR"
	gw := "10.0.0.1"
	tapName := "tap" + in.Name
	vmInfo = pb.ReturnPortInfo{
		Ip:       ip,
		Deviceid: tapName,
		Remoteid: portID,
		Status:   common_pb.Status_ERROR,
	}
	// Create Device
	log.Println("Adding tap " + tapName)
	cmd := exec.Command("bash", "-c", "ip tuntap add mode tap "+tapName)
	stdout, err := cmd.Output()
	if err != nil {
		log.Println("ip tuntap add failed!" + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "ovs-vsctl failed! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}
	log.Println("Creating Namespace " + in.Name)
	cmd = exec.Command("bash", "-c", "ip netns add "+in.Name)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Namespace creation failed! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Namespace creation failed! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Creating veth pair in" + in.Name + " and out" + in.Name)
	cmd = exec.Command("bash", "-c", "ip link add in"+in.Name+" type veth peer name out"+in.Name)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Inner and outer veth creation failed! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Inner and outer veth creation failed! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}
	log.Println("Moving veth in" + in.Name + " to namespace " + in.Name)
	cmd = exec.Command("bash", "-c", "ip link set in"+in.Name+" netns "+in.Name)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Move veth into namespace failed! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Move veth into namespace failed! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Assigning IP " + ip + " to veth device")
	cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ip addr add "+ip+"/"+strings.Split(in.Cidr, "/")[1]+" dev in"+in.Name)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to give inner veth IP! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to give inner veth IP! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Bringing inner veth up")
	cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ip link set dev in"+in.Name+" up")
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed bring up inner veth! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed bring up inner veth! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Setting MTU probing")
	cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" sysctl -w net.ipv4.tcp_mtu_probing=2")
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to set MTU probing! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to set MTU probing! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Bringing up outer veth")
	cmd = exec.Command("bash", "-c", "ip link set dev out"+in.Name+" up")
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to bring up outer veth!  " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to bring up outer veth!  " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Bringing up loopback")
	cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ip link set dev lo up")
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to bring up loopback! " + string(stdout) + " " + err.Error())
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to bring up loopback! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Assigning MAC " + mac + " address to veth")
	cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ip link set dev in"+in.Name+" address "+mac)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to assign mac! " + string(stdout) + " " + err.Error())
		return &pb.AgentReturnInfo{
			ReturnMessage: "Assign mac! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Adding default Gateway " + gw)
	cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ip r add default via "+gw)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed add default gw! " + gw + " " + string(stdout) + " " + err.Error())
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed add default gw! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Creating bridge device bridge" + in.Name)
	cmd = exec.Command("bash", "-c", "ip link add name bridge"+in.Name+" type bridge")
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to create bridge! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to create bridge! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Adding veth to bridge")
	cmd = exec.Command("bash", "-c", "ip link set out"+in.Name+" master bridge"+in.Name)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to add veth to bridge! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to add veth to bridge! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Adding TAP device to bridge")
	cmd = exec.Command("bash", "-c", "ip link set "+tapName+" master bridge"+in.Name)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to add tap to bridge " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to add tap to bridge " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Bringing bridge up")
	cmd = exec.Command("bash", "-c", "ip link set dev bridge"+in.Name+" up")
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to bring up bridge " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to bring up bridge " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Bringing Tap device up")
	cmd = exec.Command("bash", "-c", "ip link set dev "+tapName+" up")
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed to bring up tap device " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to bring up tap device " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}
	log.Println("Successfully created devices!")

	vmInfo.Status = common_pb.Status_DONE
	return &pb.AgentReturnInfo{
		ReturnMessage: "Create Success",
		ReturnCode:    common_pb.ReturnCode_OK,
		Port: &pb.ReturnPortInfo{
			Ip:       ip,
			Deviceid: tapName,
			Remoteid: portID,
			Status:   common_pb.Status_DONE,
		},
	}, nil
}
