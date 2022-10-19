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
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/tidwall/gjson"
)

type port struct {
	Port minimalPort `json:"port"`
}

type minimalPort struct {
	AdminState bool                `json:"admin_state_up"`
	DeviceID   string              `json:"device_id"`
	NetworkID  string              `json:"network_id"`
	SG         []string            `json:"security_groups"`
	FixIPs     []map[string]string `json:"fixed_ips"`
	TenantID   string              `json:"tenant_id"`
}

type updatePortMain struct {
	Port updatePort `json:"port"`
}

type updatePort struct {
	ProjectID     string `json:"project_id"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	NetworkID     string `json:"network_id"`
	TenantID      string `json:"tenant_id"`
	AdminState    bool   `json:"admin_state_up"`
	VethName      string `json:"veth_name"`
	DeviceID      string `json:"device_id"`
	DeviceOwner   string `json:"device_owner"`
	FastPath      bool   `json:"fast_path"`
	BindingHostID string `json:"binding:host_id"`
}

func caseCreate(ctx context.Context, in *pb.InternalPortConfig) (*pb.AgentReturnInfo, error) {

	log.Println("Create Minimal Port")
	vmInfo := pb.ReturnPortInfo{
		Ip:       "",
		Deviceid: "",
		Remoteid: "",
		Status:   common_pb.Status_ERROR,
	}
	minimalPortBody := port{
		Port: minimalPort{
			AdminState: true,
			DeviceID:   in.Name,
			NetworkID:  in.Vpcid,
			SG:         []string{in.Sg},
			FixIPs:     []map[string]string{{"subnet_id": in.Subnetid}},
			TenantID:   in.Tenantid,
		},
	}

	body, err := json.Marshal(minimalPortBody)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to marshal json!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}
	log.Println("Sending body to Alcor: \n", string(body[:]))
	resp, err := http.Post("http://"+RemoteServer+":"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to send create minimal port to Alcor!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}
	log.Println("VM Name: "+in.Name+" Response code from Alcor", resp.StatusCode)
	if resp.StatusCode != constants.HTTP_CREATE_SUCCESS {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to create minimal port! Response Code: " + strconv.Itoa(resp.StatusCode),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, errors.New("Failed to create minimal port! Response Code: " + strconv.Itoa(resp.StatusCode))
	}
	respBodyByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to parse response",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}
	respBody := string(respBodyByte[:])
	log.Println("Reponse Body: ", string(respBody))
	// Parse response from Alcor
	ip := gjson.Get(string(respBody), "port.fixed_ips.0.ip_address").Str
	mac := gjson.Get(string(respBody), "port.mac_address").Str
	portID := gjson.Get(string(respBody), "port.id").Str
	if constants.ALCOR_PORT_ID_SUBSTRING_LENGTH >= len(portID) {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Port ID from Alcor too short",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	tapName := "tap" + portID[:11]
	vmInfo = pb.ReturnPortInfo{
		Ip:       ip,
		Deviceid: tapName,
		Remoteid: portID,
		Status:   common_pb.Status_ERROR,
	}
	// Create Device
	log.Println("Adding tap " + tapName + " to br-int!")
	cmd := exec.Command("bash", "-c", "ovs-vsctl add-port br-int "+tapName+" --  set Interface "+tapName+" type=internal")
	stdout, err := cmd.Output()
	if err != nil {
		log.Println("ovs-vsctl failed! " + string(stdout))
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
		log.Println("Failed to bring up loopback! " + string(stdout))
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
		log.Println("Assign mac! " + string(stdout))
		return &pb.AgentReturnInfo{
			ReturnMessage: "Assign mac! " + string(stdout),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Adding default Gateway " + in.Gw)
	cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ip r add default via "+in.Gw)
	stdout, err = cmd.Output()
	if err != nil {
		log.Println("Failed add default gw! " + string(stdout))
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

	updatePortBody := updatePortMain{
		updatePort{
			ProjectID:     in.Projectid,
			ID:            portID,
			Name:          in.Name,
			Description:   "",
			NetworkID:     in.Vpcid,
			TenantID:      in.Tenantid,
			AdminState:    true,
			VethName:      "in" + in.Name,
			DeviceID:      in.Name,
			DeviceOwner:   "compute:nova",
			FastPath:      true,
			BindingHostID: in.Hostname,
		},
	}

	body, err = json.Marshal(updatePortBody)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to marshal json!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}
	jsonStringBody := string(body[:])
	log.Println("Creating update_port request with body: \n", jsonStringBody)
	req, err := http.NewRequest(http.MethodPut, "http://"+RemoteServer+":"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports/"+portID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		log.Println("Failed send Update Port request to Alcor!", err)
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed send Update Port request to Alcor!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}

	log.Println("Sending update_port request to Alcor for " + in.Name + " ID " + portID)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Println("Failed to update port to Alcor!: \n", jsonStringBody)
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed update port!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, err
	}
	log.Println("VM Name: "+in.Name+" PortID: "+portID+" Response code from Alcor update-port ", resp.StatusCode)
	if resp.StatusCode != constants.HTTP_OK {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to update_port! Response Code: " + strconv.Itoa(resp.StatusCode),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          &vmInfo,
		}, errors.New("Failed to update port! Response Code: " + strconv.Itoa(resp.StatusCode))
	}
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
