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

package evm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/futurewei-cloud/merak/services/common/metrics"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	constants "github.com/futurewei-cloud/merak/services/common"
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

type evmError struct {
	Err     error
	Message string
}

func (r evmError) Error() string {
	return fmt.Sprintf("%s: %v", r.Message, r.Err)
}

// Creates a new EVM with the given attributes
func NewEvm(name, ip, mac, remoteID, deviceID, cidr, gw string, status common_pb.Status) (Evm, error) {

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
func CreateMinimalPort(url string, in *pb.InternalPortConfig, m metrics.Metrics) (Evm, error) {
	log.Println("Create minimal port")

	var err error
	defer m.GetMetrics(&err)()

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
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
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
func UpdatePort(url string, in *pb.InternalPortConfig, m metrics.Metrics, evm Evm) error {
	var err error
	defer m.GetMetrics(&err)()

	updatePortBody := updatePortMain{
		updatePort{
			ProjectID:     in.Projectid,
			ID:            evm.GetRemoteId(),
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
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		log.Println("Failed send Update Port request to Alcor!", err)
		return err
	}
	log.Println("Sending update_port request to Alcor for " + evm.GetName() + " ID " + evm.GetRemoteId())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to update port to Alcor!: \n", jsonStringBody)
		return err
	}
	log.Println("VM Name: "+evm.GetName()+" PortID: "+evm.GetRemoteId()+" Response code from Alcor update-port ", resp.StatusCode)
	if resp.StatusCode != constants.HTTP_OK {
		return errors.New("Failed to update port! Response Code: " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}

// Sends a delete port request to Alcor
func DeletePort(url string, in *pb.InternalPortConfig, m metrics.Metrics, evm Evm) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Send Delete Port Request to Alcor")
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(nil))
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
	log.Println("VM Name: "+evm.GetName()+" Port ID: "+evm.GetRemoteId()+" Response code from Alcor delete-port ", resp.StatusCode)
	if resp.StatusCode != constants.HTTP_OK {
		log.Println("Alcor failed to delete, response " + strconv.Itoa(resp.StatusCode))
		return errors.New("Failed to delete port! Response Code: " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}

// Creates a new Tap device and adds it to the ovs-bridge
func (evm AlcorEvm) CreateDevice(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Adding tap " + evm.deviceID + " to br-int!")
	stdout, err := BashExec("ovs-vsctl add-port br-int " + evm.deviceID + " --  set Interface " + evm.deviceID + " type=internal")
	if err != nil {
		log.Println("ovs-vsctl failed! " + string(stdout))
		return err
	}
	return nil
}

// Creates a tap device for testing without Alcor
func (evm AlcorEvm) CreateStandaloneDevice(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Adding tap (standalone)" + evm.deviceID)
	stdout, err := BashExec("ip tuntap add mode tap " + evm.deviceID)
	if err != nil {
		log.Println("ip tuntap add failed!" + string(stdout))
		return err
	}
	return nil
}

// Deletes a Tap devices from the ovs bridge
func (evm AlcorEvm) DeleteDevice(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Deleting Tap device from br-int" + evm.deviceID)
	stdout, err := BashExec("ovs-vsctl del-port br-int " + evm.deviceID)
	if err != nil {
		log.Println("Failed to delete tap " + string(stdout))
		return err
	}
	return nil
}

// Deletes a Tap devices from the ovs bridge
func (evm AlcorEvm) DeleteStandaloneDevice(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Deleting TAP device (standalone) " + evm.deviceID)
	stdout, err := BashExec("ip tuntap del mode tap " + evm.deviceID)
	if err != nil {
		log.Println("Failed to delete tap " + string(stdout))
		return err
	}
	return nil
}

// Creates a new network namespace
func (evm AlcorEvm) CreateNamespace(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Creating Namespace " + evm.name)
	stdout, err := BashExec("ip netns add " + evm.name)
	if err != nil {
		log.Println("Namespace creation failed! " + string(stdout))
		return err
	}
	return nil
}

// Deletes a network namespace
func (evm AlcorEvm) DeleteNamespace(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Deleting Namespace")
	stdout, err := BashExec("ip netns delete " + evm.name)
	if err != nil {
		log.Println("Namespace deletion failed! " + string(stdout))
		return err
	}
	return nil
}

func (evm AlcorEvm) MoveDeviceToNetns(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Moving tap " + evm.deviceID + " to namespace " + evm.name)
	stdout, err := BashExec("ip link set " + evm.deviceID + " netns " + evm.name)
	if err != nil {
		log.Println("Move tap into namespace failed! " + string(stdout))
		return err
	}
	return nil
}

func (evm AlcorEvm) MoveDeviceToRootNetns(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Moving tap " + evm.deviceID + " to root namespace " + evm.name)
	stdout, err := BashExec("ip netns exec " + evm.name + " ip link set " + evm.deviceID + " netns 1")
	if err != nil {
		log.Println("Move tap to root namespace failed! " + string(stdout))
		return err
	}
	return nil
}

// Assigns an IP address to the inner veth-pair
func (evm AlcorEvm) AssignIP(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Assigning IP " + evm.ip + " to tap device")
	stdout, err := BashExec("ip netns exec " + evm.name + " ip addr add " + evm.ip + "/" + strings.Split(evm.cidr, "/")[1] + " dev " + evm.deviceID)
	if err != nil {
		log.Println("Failed to give tap IP! " + string(stdout))
		return err
	}
	return nil
}

// Sets MTU probing for the inner veth
func (evm AlcorEvm) SetMTUProbing(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Setting MTU probing")
	stdout, err := BashExec("ip netns exec " + evm.name + " sysctl -w net.ipv4.tcp_mtu_probing=2")
	if err != nil {
		log.Println("Failed to set MTU probing! " + string(stdout))
		return err
	}
	return nil
}

// Brings the loopback device inside the network namespace up
func (evm AlcorEvm) BringLoUp(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Bringing up loopback")
	stdout, err := BashExec("ip netns exec " + evm.name + " ip link set dev lo up")
	if err != nil {
		log.Println("Failed to bring up loopback! " + string(stdout))
		return err
	}
	return nil
}

// Assigns a mac address to the inner veth
func (evm AlcorEvm) AssignMac(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Assigning MAC " + evm.mac + " address to veth")
	stdout, err := BashExec("ip netns exec " + evm.name + " ip link set dev " + evm.deviceID + " address " + evm.mac)
	if err != nil {
		log.Println("Assign mac! " + string(stdout))
		return err
	}
	return nil
}

// Adds a gateway inside the network namespace
func (evm AlcorEvm) AddGateway(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Adding default Gateway " + evm.gw)
	stdout, err := BashExec("ip netns exec " + evm.name + " ip r add default via " + evm.gw)
	if err != nil {
		log.Println("Failed add default gw! " + string(stdout))
		return err
	}
	return nil
}

// Brings the tap device up
func (evm AlcorEvm) BringDeviceUp(m metrics.Metrics) error {
	var err error
	defer m.GetMetrics(&err)()

	log.Println("Bringing Tap device up")
	stdout, err := BashExec("ip netns exec " + evm.name + " ip link set dev " + evm.deviceID + " up")
	if err != nil {
		log.Println("Failed to bring up tap device " + string(stdout))
		return err
	}
	return nil
}

// Returns the EVM's name
func (evm AlcorEvm) GetName() string {
	return evm.name
}

// Returns the EVM's IP
func (evm AlcorEvm) GetIP() string {
	return evm.ip
}

// Returns the EVM's MAC address
func (evm AlcorEvm) GetMac() string {
	return evm.mac
}

// Returns the EVM's network cidr
func (evm AlcorEvm) GetCidr() string {
	return evm.cidr
}

// Returns the EVM's gateway address
func (evm AlcorEvm) GetGw() string {
	return evm.gw
}

// Returns the EVM's tap device name
func (evm AlcorEvm) GetDeviceId() string {
	return evm.deviceID
}

// Returns the EVM's ID in Alcor
func (evm AlcorEvm) GetRemoteId() string {
	return evm.remoteID
}

// Returns the EVM's status
func (evm AlcorEvm) GetStatus() common_pb.Status {
	return evm.status
}

// Sets the EVM's name
func (evm *AlcorEvm) SetName(name string) {
	evm.name = name
}

// Sets the EVM's IP address
func (evm *AlcorEvm) SetIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return evmError{errors.New("Invalid IP Address"), ip}
	}
	evm.ip = ip
	return nil
}

// Sets the EVM's MAC address
func (evm *AlcorEvm) SetMac(mac string) error {
	_, err := net.ParseMAC(mac)
	if err != nil {
		return evmError{errors.New("Invalid MAC address"), mac}
	}
	evm.mac = mac
	return nil
}

// Sets the EVM's cidr
func (evm *AlcorEvm) SetCidr(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return evmError{errors.New("Invalid CIDR address"), cidr}
	}
	evm.cidr = cidr
	return nil
}

// Sets the EVM's Gateway
func (evm *AlcorEvm) SetGw(gw string) error {
	if net.ParseIP(gw) == nil {
		return evmError{errors.New("Invalid GW address"), gw}
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

var BashExec = BashExecute

// Executes the given Bash command
func BashExecute(cmd string) ([]byte, error) {
	return exec.Command("Bash", "-c", cmd).Output()
}
