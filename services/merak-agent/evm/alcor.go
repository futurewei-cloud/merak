package evm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
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

var _ Evm = (*AlcorEvm)(nil)

type AlcorEvm struct {
	name     string
	ip       string
	mac      string
	cidr     string
	gw       string
	deviceID string
	remoteID string
	status   common_pb.Status
}

// Creates a new EVM with the given attributes
func NewEvm(name, ip, mac, remoteID, deviceID, cidr, gw string, status common_pb.Status) (*AlcorEvm, error) {
	evm := &AlcorEvm{}

	evm.SetName(name)
	evm.SetRemoteId(remoteID)
	evm.SetDeviceId(deviceID)
	evm.SetStatus(status)

	err := evm.SetIP(ip)
	if err != nil {
		return nil, err
	}
	err = evm.SetMac(mac)
	if err != nil {
		return nil, err
	}
	err = evm.SetCidr(cidr)
	if err != nil {
		return nil, err
	}
	err = evm.SetGw(gw)
	if err != nil {

		return nil, err
	}
	log.Println("Successfully recieved EVM info ", evm)
	return evm, nil
}

// Sends a create minimal port request to alcor
func CreateMinimalPort(in *pb.InternalPortConfig, remoteServer string) (*AlcorEvm, error) {
	log.Println("Create minimal port")
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
		return nil, err
	}

	log.Println("Sending body to Alcor: \n", string(body[:]))
	resp, err := http.Post("http://"+remoteServer+":"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	log.Println("VM Name: "+in.Name+" Response code from Alcor", resp.StatusCode)
	if resp.StatusCode != constants.HTTP_CREATE_SUCCESS {
		return nil, errors.New("Failed to create minimal port! Response Code: " + strconv.Itoa(resp.StatusCode))
	}
	respBodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respBody := string(respBodyByte[:])
	log.Println("Reponse Body: ", string(respBody))
	// Parse response from Alcor
	ip := gjson.Get(string(respBody), "port.fixed_ips.0.ip_address").Str
	mac := gjson.Get(string(respBody), "port.mac_address").Str
	portID := gjson.Get(string(respBody), "port.id").Str
	if constants.ALCOR_PORT_ID_SUBSTRING_LENGTH >= len(portID) {
		return nil, err
	}
	tapName := "tap" + portID[:11]

	evm, err := NewEvm(in.Name, ip, mac, portID, tapName, in.Cidr, in.Gw, common_pb.Status_DEPLOYING)
	if err != nil {
		return nil, err
	}
	return evm, nil
}

// Sends an update port request to Alcor
func (evm *AlcorEvm) UpdatePort(in *pb.InternalPortConfig, remoteServer string) error {
	updatePortBody := updatePortMain{
		updatePort{
			ProjectID:     in.Projectid,
			ID:            evm.remoteID,
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

	body, err := json.Marshal(updatePortBody)
	if err != nil {
		return err
	}
	jsonStringBody := string(body[:])
	log.Println("Creating update_port request with body: \n", jsonStringBody)
	req, err := http.NewRequest(http.MethodPut, "http://"+remoteServer+":"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports/"+evm.remoteID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		log.Println("Failed send Update Port request to Alcor!", err)
		return err
	}

	log.Println("Sending update_port request to Alcor for " + evm.name + " ID " + evm.remoteID)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to update port to Alcor!: \n", jsonStringBody)
		return err
	}
	log.Println("VM Name: "+evm.name+" PortID: "+evm.remoteID+" Response code from Alcor update-port ", resp.StatusCode)
	if resp.StatusCode != constants.HTTP_OK {
		return errors.New("Failed to update port! Response Code: " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}

// Sends a delete port request to Alcor
func (evm *AlcorEvm) DeletePort(in *pb.InternalPortConfig, remoteServer string) error {
	log.Println("Send Delete Port Request to Alcor")
	req, err := http.NewRequest(http.MethodDelete, "http://"+remoteServer+":"+strconv.Itoa(constants.ALCOR_PORT_MANAGER_PORT)+"/project/"+in.Projectid+"/ports/"+evm.remoteID, bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		log.Println("Failed send Delete Port request to Alcor!", err)
		return err
	}

	log.Println("Sending delete port request to Alcor", req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to delete port to Alcor!")
		return err
	}
	log.Println("VM Name: "+evm.name+" Port ID: "+evm.remoteID+" Response code from Alcor delete-port ", resp.StatusCode)
	if resp.StatusCode != constants.HTTP_OK {
		log.Println("Alcor failed to delete, response " + strconv.Itoa(resp.StatusCode))
		return errors.New("Failed to delete port! Response Code: " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}

// Creates a new Tap device and adds it to the ovs-bridge
func (evm *AlcorEvm) CreateDevice() error {
	log.Println("Adding tap " + evm.deviceID + " to br-int!")
	stdout, err := bashExec("ovs-vsctl add-port br-int " + evm.deviceID + " --  set Interface " + evm.deviceID + " type=internal")
	if err != nil {
		log.Println("ovs-vsctl failed! " + string(stdout))
		return err
	}
	return nil
}

// Creates a tap device for testing without Alcor
func (evm *AlcorEvm) CreateStandaloneDevice() error {
	log.Println("Adding tap (standalone)" + evm.deviceID)
	stdout, err := bashExec("ip tuntap add mode tap " + evm.deviceID)
	if err != nil {
		log.Println("ip tuntap add failed!" + string(stdout))
		return err
	}
	return nil
}

// Deletes a Tap devices from the ovs bridge
func (evm *AlcorEvm) DeleteDevice() error {
	log.Println("Deleting Tap device from br-int" + evm.deviceID)
	stdout, err := bashExec("ovs-vsctl del-port br-int " + evm.deviceID)
	if err != nil {
		log.Println("Failed to delete tap " + string(stdout))
		return err
	}
	return nil
}

// Deletes a Tap devices from the ovs bridge
func (evm *AlcorEvm) DeleteStandaloneDevice() error {
	log.Println("Deleting TAP device (standalone) " + evm.deviceID)
	stdout, err := bashExec("ip tuntap del mode tap " + evm.deviceID)
	if err != nil {
		log.Println("Failed to delete tap " + string(stdout))
		return err
	}
	return nil
}

// Creates a new network namespace
func (evm *AlcorEvm) CreateNamespace() error {
	log.Println("Creating Namespace " + evm.name)
	stdout, err := bashExec("ip netns add " + evm.name)
	if err != nil {
		log.Println("Namespace creation failed! " + string(stdout))
		return err
	}
	return nil
}

// Deletes a network namespace
func (evm *AlcorEvm) DeleteNamespace() error {
	log.Println("Deleting Namespace")
	stdout, err := bashExec("ip netns delete " + evm.name)
	if err != nil {
		log.Println("Namespace deletion failed! " + string(stdout))
		return err
	}
	return nil
}

// Creates a new veth pair
func (evm *AlcorEvm) CreateVethPair() error {
	log.Println("Creating veth pair in" + evm.name + " and out" + evm.name)
	stdout, err := bashExec("ip link add in" + evm.name + " type veth peer name out" + evm.name)
	if err != nil {
		log.Println("Inner and outer veth creation failed! " + string(stdout))
		return err
	}
	return nil
}

// Moves one end of the veth-pair to the network namespace
func (evm *AlcorEvm) MoveVethToNamespace() error {
	log.Println("Moving veth in" + evm.name + " to namespace " + evm.name)
	stdout, err := bashExec("ip link set in" + evm.name + " netns " + evm.name)
	if err != nil {
		log.Println("Move veth into namespace failed! " + string(stdout))
		return err
	}
	return nil
}

// Assigns an IP address to the inner veth-pair
func (evm *AlcorEvm) AssignIP() error {
	log.Println("Assigning IP " + evm.ip + " to veth device")
	stdout, err := bashExec("ip netns exec " + evm.name + " ip addr add " + evm.ip + "/" + strings.Split(evm.cidr, "/")[1] + " dev in" + evm.name)
	if err != nil {
		log.Println("Failed to give inner veth IP! " + string(stdout))
		return err
	}
	return nil
}

// Brings the inner veth up
func (evm *AlcorEvm) BringInnerVethUp() error {
	log.Println("Bringing inner veth up")
	stdout, err := bashExec("ip netns exec " + evm.name + " ip link set dev in" + evm.name + " up")
	if err != nil {
		log.Println("Failed bring up inner veth! " + string(stdout))
		return err
	}
	return nil
}

// Sets MTU probing for the inner veth
func (evm *AlcorEvm) SetMTUProbing() error {
	log.Println("Setting MTU probing")
	stdout, err := bashExec("ip netns exec " + evm.name + " sysctl -w net.ipv4.tcp_mtu_probing=2")
	if err != nil {
		log.Println("Failed to set MTU probing! " + string(stdout))
		return err
	}
	return nil
}

// Brings the outer veth up
func (evm *AlcorEvm) BringOuterVethUp() error {
	log.Println("Bringing up outer veth")
	stdout, err := bashExec("ip link set dev out" + evm.name + " up")
	if err != nil {
		log.Println("Failed to bring up outer veth!  " + string(stdout))
		return err
	}
	return nil
}

// Brings the loopback device inside the network namespace up
func (evm *AlcorEvm) BringLoUp() error {
	log.Println("Bringing up loopback")
	stdout, err := bashExec("ip netns exec " + evm.name + " ip link set dev lo up")
	if err != nil {
		log.Println("Failed to bring up loopback! " + string(stdout))
		return err
	}
	return nil
}

// Assigns a mac address to the inner veth
func (evm *AlcorEvm) AssignMac() error {
	log.Println("Assigning MAC " + evm.mac + " address to veth")
	stdout, err := bashExec("ip netns exec " + evm.name + " ip link set dev in" + evm.name + " address " + evm.mac)
	if err != nil {
		log.Println("Assign mac! " + string(stdout))
		return err
	}
	return nil
}

// Adds a gateway inside the network namespace
func (evm *AlcorEvm) AddGateway() error {
	log.Println("Adding default Gateway " + evm.gw)
	stdout, err := bashExec("ip netns exec " + evm.name + " ip r add default via " + evm.gw)
	if err != nil {
		log.Println("Failed add default gw! " + string(stdout))
		return err
	}
	return nil
}

// Deletes a linux bridge
func (evm *AlcorEvm) DeleteBridge() error {
	log.Println("Deleting bridge device")
	stdout, err := bashExec("ip link delete bridge" + evm.name)
	if err != nil {
		log.Println("Bridge deletion failed! " + string(stdout))
		return err
	}
	return nil
}

// Creates a linux bridge
func (evm *AlcorEvm) CreateBridge() error {
	log.Println("Creating bridge device bridge" + evm.name)
	stdout, err := bashExec("ip link add name bridge" + evm.name + " type bridge")
	if err != nil {
		log.Println("Failed to create bridge! " + string(stdout))
		return err
	}
	return nil
}

// Adds the outer veth to a linux bridge
func (evm *AlcorEvm) AddVethToBridge() error {
	log.Println("Adding veth to bridge")
	stdout, err := bashExec("ip link set out" + evm.name + " master bridge" + evm.name)
	if err != nil {
		log.Println("Failed to add veth to bridge! " + string(stdout))
		return err
	}
	return nil
}

// Adds a tap device to the linux bridge
func (evm *AlcorEvm) AddDeviceToBridge() error {
	log.Println("Adding TAP device to bridge " + evm.deviceID)
	stdout, err := bashExec("ip link set " + evm.deviceID + " master bridge" + evm.name)
	if err != nil {
		log.Println("Failed to add tap to bridge " + string(stdout))
		return err
	}
	return nil
}

// Brings the linux bridge up
func (evm *AlcorEvm) BringBridgeUp() error {
	log.Println("Bringing bridge up")
	stdout, err := bashExec("ip link set dev bridge" + evm.name + " up")
	if err != nil {
		log.Println("Failed to bring up bridge " + string(stdout))
		return err
	}
	return nil
}

// Brings the tap device up
func (evm *AlcorEvm) BringDeviceUp() error {
	log.Println("Bringing Tap device up")
	stdout, err := bashExec("ip link set dev " + evm.deviceID + " up")
	if err != nil {
		log.Println("Failed to bring up tap device " + string(stdout))
		return err
	}
	return nil
}

// Returns the EVM's name
func (evm *AlcorEvm) GetName() string {
	return evm.name
}

// Returns the EVM's IP
func (evm *AlcorEvm) GetIP() string {
	return evm.ip
}

// Returns the EVM's MAC address
func (evm *AlcorEvm) GetMac() string {
	return evm.mac
}

// Returns the EVM's network cidr
func (evm *AlcorEvm) GetCidr() string {
	return evm.cidr
}

// Returns the EVM's gateway address
func (evm *AlcorEvm) GetGw() string {
	return evm.gw
}

// Returns the EVM's tap device name
func (evm *AlcorEvm) GetDeviceId() string {
	return evm.deviceID
}

// Returns the EVM's ID in Alcor
func (evm *AlcorEvm) GetRemoteId() string {
	return evm.remoteID
}

// Returns the EVM's status
func (evm *AlcorEvm) GetStatus() common_pb.Status {
	return evm.status
}

// Sets the EVM's name
func (evm *AlcorEvm) SetName(name string) {
	evm.name = name
}

// Sets the EVM's IP address
func (evm *AlcorEvm) SetIP(ip string) error {
	if net.ParseIP(ip) == nil {
		log.Fatalf("Invalid IP address %s\n", ip)
	}
	evm.ip = ip
	return nil
}

// Sets the EVM's MAC address
func (evm *AlcorEvm) SetMac(mac string) error {
	_, err := net.ParseMAC(mac)
	if err != nil {
		log.Fatalf("Invalid MAC address %s\n", mac)
	}
	evm.mac = mac
	return nil
}

// Sets the EVM's cidr
func (evm *AlcorEvm) SetCidr(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatalf("Invalid CIDR %s\n", cidr)
	}
	evm.cidr = cidr
	return nil
}

// Sets the EVM's Gateway
func (evm *AlcorEvm) SetGw(gw string) error {
	if net.ParseIP(gw) == nil {
		log.Fatalf("Invalid GW address %s\n", gw)
	}
	evm.gw = gw
	return nil
}

// Sets the EVM's tap device name
func (evm *AlcorEvm) SetDeviceId(id string) {
	evm.deviceID = id
}

// Sets the EVM's ID from Alcor
func (evm *AlcorEvm) SetRemoteId(id string) {
	evm.remoteID = id
}

// Sets the EVM's status
func (evm *AlcorEvm) SetStatus(status common_pb.Status) {
	evm.status = status
}

// Executes the given bash command
func bashExec(cmd string) ([]byte, error) {
	return exec.Command("bash", "-c", cmd).Output()
}
