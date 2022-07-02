package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/tidwall/gjson"
)

var (
	Port = flag.Int("port", constants.AGENT_GRPC_SERVER_PORT, "The server port")
)

type Server struct {
	pb.UnimplementedMerakAgentServiceServer
}

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

func (s *Server) PortHandler(ctx context.Context, in *pb.InternalPortConfig) (*pb.ReturnMessage, error) {
	log.Println("Received on PortHandler", in)

	// Parse input
	switch op := in.OperationType; op {
	case pb.OperationType_INFO:
		log.Println("Info Unimplemented")
		return &pb.ReturnMessage{
			ReturnMessage: "Info Unimplemented",
			ReturnCode:    pb.ReturnCode_FAILED,
		}, errors.New("info unimplemented")

	case pb.OperationType_CREATE:
		log.Println("Operation Create")

		log.Println("Create Minimal Port")

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
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to marshal json!",
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Sending body to Alcor: \n", string(body[:]))
		resp, err := http.Post("http://"+constants.ALCOR_ADDRESS+":"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports", "application/json", bytes.NewBuffer(body))
		if err != nil {
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to send create minimal port to Alcor!",
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Response code from Alcor", resp.StatusCode)

		respBodyByte, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to parse response",
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		respBody := string(respBodyByte[:])
		log.Println("Reponse Body: ", string(respBody))
		// Parse response from Alcor
		ip := gjson.Get(string(respBody), "port.fixed_ips.0.ip_address").Str
		mac := gjson.Get(string(respBody), "port.mac_address").Str
		portID := gjson.Get(string(respBody), "port.id").Str
		if constants.ALCOR_PORT_ID_SUBSTRING_LENGTH >= len(portID) {
			return &pb.ReturnMessage{
				ReturnMessage: "Port ID from Alcor too short",
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		tapName := "tap" + gjson.Get(string(respBody), "port.id").Str[:11]

		// Create Device
		log.Println("OVS setup")
		cmd := exec.Command("bash", "-c", "ovs-vsctl add-port br-int "+tapName+" --  set Interface "+tapName+" type=internal")
		stdout, err := cmd.Output()
		if err != nil {
			log.Println("ovs-vsctl failed! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "ovs-vsctl failed! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Creating Namespace")
		cmd = exec.Command("bash", "-c", "ip netns add "+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Namespace creation failed! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Namespace creation failed! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Creating veth pair")
		cmd = exec.Command("bash", "-c", "ip link add in"+in.Name+" type veth peer name out"+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Inner and outer veth creation failed! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Inner and outer veth creation failed! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Moving veth to namespace")
		cmd = exec.Command("bash", "-c", "ip link set in"+in.Name+" netns "+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Move veth into namespace failed! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Move veth into namespace failed! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Assigning IP address to veth device")
		cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ip addr add "+ip+"/"+strings.Split(in.Cidr, "/")[1]+" dev in"+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed to give inner veth IP! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to give inner veth IP! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Bringing inner veth up")
		cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ip link set dev in"+in.Name+" up")
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed bring up inner veth! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed bring up inner veth! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Setting MTU probing")
		cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" sysctl -w net.ipv4.tcp_mtu_probing=2")
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed to set MTU probing! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to set MTU probing! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Bringing up outer veth")
		cmd = exec.Command("bash", "-c", "ip link set dev out"+in.Name+" up")
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed to bring up outer veth!  " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to bring up outer veth!  " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Bringing up loopback")
		cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ifconfig lo up")
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed to bring up loopback! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to bring up loopback! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Assigning MAC address to veth")
		cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" ifconfig in"+in.Name+" hw ether "+mac)
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Assign mac! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Assign mac! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Adding default Gateway")
		cmd = exec.Command("bash", "-c", "ip netns exec "+in.Name+" route add default gw "+in.Gw)
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed add default gw! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed add default gw! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Creating bridge device")
		cmd = exec.Command("bash", "-c", "ip link add name bridge"+in.Name+" type bridge")
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed to ceate bridge! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to ceate bridge! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Adding veth to bridge")
		cmd = exec.Command("bash", "-c", "ip link set out"+in.Name+" master bridge"+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed to add veth to bridge! " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to add veth to bridge! " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Adding TAP device to bridge")
		cmd = exec.Command("bash", "-c", "ip link set "+tapName+" master bridge"+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed to add tap to bridge " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to add tap to bridge " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Bringing bridge up")
		cmd = exec.Command("bash", "-c", "ip link set dev bridge"+in.Name+" up")
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed to bring up bridge " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to bring up bridge " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}

		log.Println("Bringing Tap device up")
		cmd = exec.Command("bash", "-c", "ip link set dev "+tapName+" up")
		stdout, err = cmd.Output()
		if err != nil {
			log.Println("Failed to bring up tap device " + string(stdout))
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to bring up tap device " + string(stdout),
				ReturnCode:    pb.ReturnCode_FAILED,
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
				BindingHostID: in.Name,
			},
		}

		body, err = json.Marshal(updatePortBody)
		if err != nil {
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to marshal json!",
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		jsonStringBody := string(body[:])
		log.Println("Sending body to Alcor: \n", jsonStringBody)
		req, err := http.NewRequest(http.MethodPut, "http://"+constants.ALCOR_ADDRESS+":"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports/"+portID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		if err != nil {
			return &pb.ReturnMessage{
				ReturnMessage: "Failed send Update Port request to Alcor!",
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			log.Println("Failed to update port to Alcor!: \n", jsonStringBody)
			return &pb.ReturnMessage{
				ReturnMessage: "Failed update port!",
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		respBodyByte, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return &pb.ReturnMessage{
				ReturnMessage: "Failed to parse response",
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Response code from Alcor", resp.StatusCode)
		log.Println("Response from Alcor Update Port: ", string(respBodyByte[:]))
		return &pb.ReturnMessage{
			ReturnMessage: "Create Success",
			ReturnCode:    pb.ReturnCode_OK,
		}, nil

	case pb.OperationType_UPDATE:

		log.Println("Update Unimplemented")
		return &pb.ReturnMessage{
			ReturnMessage: "Update Unimplemented",
			ReturnCode:    pb.ReturnCode_FAILED,
		}, errors.New("update unimplemented")

	case pb.OperationType_DELETE:

		log.Println("Delete Unimplemented")
		return &pb.ReturnMessage{
			ReturnMessage: "Delete Unimplemented",
			ReturnCode:    pb.ReturnCode_FAILED,
		}, errors.New("delete unimplemented")

	default:
		log.Println("Unknown Operation")
		return &pb.ReturnMessage{
			ReturnMessage: "Unknown Operation",
			ReturnCode:    pb.ReturnCode_FAILED,
		}, errors.New("unknown operation")
	}
}
