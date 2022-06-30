package handler

import (
	"bytes"
	"context"
	"encoding/json"
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
	Port          = flag.Int("port", constants.AGENT_GRPC_SERVER_PORT, "The server port")
	returnMessage = pb.ReturnMessage{
		ReturnCode:    pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

type Server struct {
	pb.UnimplementedMerakAgentServiceServer
}

type minimalPort struct {
	AdminState bool              `json:"admin_state_up"`
	DeviceID   string            `json:"device_id"`
	NetworkID  string            `json:"network_id"`
	SG         []string          `json:"security_groups"`
	FixIPs     map[string]string `json:"fixed_ips"`
	TenantID   string            `json:"tenant_id"`
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
		SetReturnMessage("Info Unimplemented", pb.ReturnCode_FAILED)
		return &returnMessage, nil

	case pb.OperationType_CREATE:
		log.Println("Operation Create")

		log.Println("Create Minimal Port")
		minimalPortBody := minimalPort{
			AdminState: true,
			DeviceID:   in.Name,
			NetworkID:  in.Vpcid,
			SG:         []string{in.Sg},
			FixIPs:     map[string]string{"subnet_id": in.Subnetid},
			TenantID:   in.Tenantid,
		}
		body, err := json.Marshal(minimalPortBody)
		if err != nil {
			SetReturnMessage("Failed to marshal json!", pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		log.Println("Sending body to Alcor: \n", body)
		resp, err := http.Post("http://10.213.43.251:"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports", "application/json", bytes.NewBuffer(body))
		if err != nil {
			SetReturnMessage("Failed to send create minimal port to Alcor!", pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		// var result map[string]interface{}
		// json.Unmarshal([]byte(respBody), &result)
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			SetReturnMessage("Failed to parse response", pb.ReturnCode_FAILED)
			return &returnMessage, err
		}
		// ip = json_response["port"]["fixed_ips"][0]["ip_address"]
		ip := gjson.Get(string(respBody), "port.fixed_ips.0.ip_address").Str
		// mac = json_response["port"]["mac_address"]
		mac := gjson.Get(string(respBody), "port.mac_address").Str
		// tap_name = "tap" + json_response["port"]["id"][:11]
		tapName := "tap" + gjson.Get(string(respBody), "port.id").Str[:11]
		// port_id = json_response["port"]["id"]
		portID := gjson.Get(string(respBody), "port.id").Str
		// Create devices
		cmd := exec.Command("ovs-vsctl", "add-port", "br-int", tapName, "--", "set", "Interface", tapName, "type=internal")
		stdout, err := cmd.Output()
		if err != nil {
			SetReturnMessage("ovs-vsctl failed! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}
		cmd = exec.Command("ip", "netns", "add", in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Namespace creation failed! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}
		cmd = exec.Command("ip", "link", "add", "in"+in.Name, "type", "veth", "peer", "name", "out"+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Inner and outer veth creation failed! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}
		cmd = exec.Command("ip", "link", "set", "in"+in.Name, "netns", in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Move veth into namespace failed! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "netns", "exec", in.Name, "ip", "addr", "addr", "add", ip+"/"+strings.Split(in.Cidr, "/")[1], "dev", "in"+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to give inner veth IP! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "netns", "exec", in.Name, "ip", "addr", "addr", "add", ip+"/"+strings.Split(in.Cidr, "/")[1], "dev", "in"+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to give inner veth IP! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "netns", "exec", in.Name, "ip", "link", "set", "dev", in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed bring up inner veth! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "netns", "exec", in.Name, "sysctl", "-w", "net.ipv4.tcp_mtu_probing=2")
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to set MTU probing! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "link", "set", "dev", "out"+in.Name, "up")
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to bring up outer veth! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "netns", "exec", in.Name, "ifconfig", "lo", "up")
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to bring up loopback! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "netns", "exec", in.Name, "ifconfig", "in"+in.Name, "hw", "ether", mac)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to set MTU probing! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "netns", "exec", in.Name, "route", "add", "add", "default", "gw", in.Gw)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to set MTU probing! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "link", "add", "name", "bridge"+in.Name, "type", "bridge")
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to set MTU probing! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "link", "set", "out"+in.Name, "master", "bridge"+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to set MTU probing! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "link", "set", tapName, "master", "bridge"+in.Name)
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to set MTU probing! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "link", "set", "dev", "bridge"+in.Name, "up")
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to set MTU probing! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		cmd = exec.Command("ip", "link", "set", "dev", tapName, "up")
		stdout, err = cmd.Output()
		if err != nil {
			SetReturnMessage("Failed to set MTU probing! "+string(stdout), pb.ReturnCode_FAILED)
			return &returnMessage, err
		}
		log.Println("Successfully created devices!")

		updatePortBody := updatePort{
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
		}

		body, err = json.Marshal(updatePortBody)
		if err != nil {
			SetReturnMessage("Failed to marshal json!", pb.ReturnCode_FAILED)
			return &returnMessage, err
		}

		log.Println("Sending body to Alcor: \n", body)
		resp, err = http.Post("http://10.213.43.251:"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports", "application/json", bytes.NewBuffer(body))
		if err != nil {
			SetReturnMessage("Failed send Update Port request to Alcor!", pb.ReturnCode_FAILED)
			return &returnMessage, err
		}
		log.Println(resp)
		SetReturnMessage("Create Success!", pb.ReturnCode_OK)
		return &returnMessage, nil

	case pb.OperationType_UPDATE:

		log.Println("Update Unimplemented")
		SetReturnMessage("Update Unimplemented", pb.ReturnCode_FAILED)
		return &returnMessage, nil

	case pb.OperationType_DELETE:

		log.Println("Delete Unimplemented")
		SetReturnMessage("Delete Unimplemented", pb.ReturnCode_FAILED)
		return &returnMessage, nil

	default:
		log.Println("Unknown Operation")
		SetReturnMessage("PortHandler: Unknown Operation", pb.ReturnCode_FAILED)
		return &returnMessage, nil
	}
}

func SetReturnMessage(returnString string, returnCode pb.ReturnCode) {
	returnMessage.ReturnCode = returnCode
	returnMessage.ReturnMessage = returnString
}
