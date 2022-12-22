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

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"github.com/tidwall/gjson"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	numVhost := 10

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	kubeConfigPath := filepath.Join(homeDir, ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Printf("error getting Kubernetes config: %v\n", err)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Printf("error getting Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}

	scenarioNodePortAddress, scenarioNodePort := getServiceAddress(clientset, "merak", "scenario-manager-service")
	scenarioConfigID := config(clientset, scenarioNodePortAddress, scenarioNodePort, uint(numVhost))

	fmt.Printf("Deploying topology with %d vhosts\n", numVhost)
	deployTopo(scenarioConfigID, scenarioNodePortAddress, scenarioNodePort, numVhost)
	fmt.Printf("Deploying Network\n")
	deployNetwork(scenarioConfigID, scenarioNodePortAddress, scenarioNodePort)
	fmt.Printf("Deploying Compute\n")
	deployCompute(scenarioConfigID, scenarioNodePortAddress, scenarioNodePort)
}

func config(clientset *kubernetes.Clientset,
	scenarioNodePortAddress string,
	scenarioNodePort int32,
	numVhost uint) string {

	topoConfig := entities.TopologyConfig{
		Name:             "Top1",
		NumberOfGateways: 0,
		NumberOfVhosts:   numVhost,
		NumberOfRacks:    2,
		TopoType:         "FATTREE",
		VhostsPerRack:    5,
		PortsPerVSwitch:  5,
		DataPlaneCidr:    "10.200.0.0/16",
		Images: []entities.Image{
			{
				Args:     []string{"string"},
				Cmd:      []string{"string"},
				Id:       "image-1",
				Name:     "ACA",
				Registry: "meraksim/merak-agent:dev",
			},
			{
				Args:     []string{"string"},
				Cmd:      []string{"string"},
				Id:       "image-2",
				Name:     "OVS",
				Registry: "yanmo96/ovs_only:latest",
			},
		},
		VLinks: []entities.VLink{
			{
				From: "string",
				Name: "string",
				To:   "string",
			},
		},
		VNodes: []entities.VNode{
			{
				Name: "string",
				Nics: []entities.Nic{
					{
						Ip:   "string",
						Name: "string",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(topoConfig)
	resp, _ := http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/topologies", "application/json", bytes.NewBuffer(body))
	respBodyByte, _ := io.ReadAll(resp.Body)
	respBody := string(respBodyByte[:])
	topoConfigID := gjson.Get(respBody, "data.id").Str

	netConfig := entities.NetworkConfig{
		Gateways: []entities.Gateway{
			{
				Ips:  []string{"string"},
				Name: "string",
			},
		},
		Name:                   "network-config-1",
		NumberOfSecurityGroups: 1,
		NumberOfVPCS:           1,
		NumberOfSubnetPerVpc:   1,
		Routers: []entities.Router{
			{
				Name: "string",
				SubnetGateways: []string{
					"string",
				},
			},
		},
		SecurityGroups: []entities.SecurityGroup{
			{
				ApplyTo:   []string{"string"},
				Name:      "sg-1",
				ProjectId: "123456789",
				Rules: []entities.Rule{
					{
						Description:    "string",
						Direction:      "string",
						EtherType:      "string",
						Name:           "string",
						PortRange:      "string",
						Protocol:       "string",
						RemoteGroupId:  "string",
						RemoteIpPrefix: "string",
					},
				},
				TenantId: "123456789",
			},
		},
		Vpcs: []entities.VPCInfo{
			{
				TenantId:        "123456789",
				ProjectId:       "123456789",
				VpcId:           "10.1.0.0/16",
				NumberOfSubnets: 1,
				VpcCidr:         "10.1.0.0/16",
				SubnetInfo: []entities.SubnetInfo{
					{
						SubnetCidr:    "10.1.0.0/16",
						SubnetGateway: "10.1.0.1",
						NumberOfVMs:   10,
					},
				},
			},
		},
	}

	body, _ = json.Marshal(netConfig)
	resp, _ = http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/network-config", "application/json", bytes.NewBuffer(body))

	respBodyByte, _ = io.ReadAll(resp.Body)
	respBody = string(respBodyByte[:])
	netConfigID := gjson.Get(respBody, "data.id").Str

	vpcManagerNodePortAddress, vpcNodePort := getServiceAddress(clientset, "default", "vpcmanager-service")
	nodeManagerNodePortAddress, nodeNodePort := getServiceAddress(clientset, "default", "nodemanager-service")
	ncmPortAddress, ncmPort := getServiceAddress(clientset, "default", "netwconfigmanager-service")
	serviceConfig := entities.ServiceConfig{
		Name: "service-config-1",
		Services: []entities.Service{
			{
				Cmd:  "alcorIp",
				Name: "alcor cluster's ip address",
				Parameters: []string{
					"-X POST",
					"-H 'Content-Type: application/json'",
					"-H 'Accept: */*'",
				},
				ReturnCode: []uint32{
					200,
					201,
				},
				ReturnString: []string{""},
				Url:          vpcManagerNodePortAddress,
				WhenToRun:    "INIT",
				WhereToRun:   "NETWORK",
			},
			{
				Cmd:  "curl",
				Name: "alcor-createDefaultTable",
				Parameters: []string{
					"-X POST",
					"-H 'Content-Type: application/json'",
					"-H 'Accept: */*'",
				},
				ReturnCode: []uint32{
					200,
					201,
				},
				ReturnString: []string{""},
				Url:          fmt.Sprintf("http://%s:%d/segments/createDefaultTable", vpcManagerNodePortAddress, vpcNodePort),
				WhenToRun:    "INIT",
				WhereToRun:   "INIT",
			},
			{
				Cmd:  "curl",
				Name: "alcor-nodemanager",
				Parameters: []string{
					"-X POST",
					"-H 'Content-Type: application/json'",
					"-H 'Accept: */*'",
					"-d '{ \"host_info\": [ { \"host_dvr_mac\": \"string\", \"local_ip\": \"string\", \"mac_address\": \"string\", \"ncm_id\": \"string\", \"ncm_uri\": \"string\", \"node_id\": \"string\", \"node_name\": \"string\", \"server_port\": 0, \"veth\": \"string\" } ]}'",
				},
				ReturnCode: []uint32{
					200,
					201,
				},
				ReturnString: []string{""},
				Url:          fmt.Sprintf("http://%s:%d/nodes/bulk", nodeManagerNodePortAddress, nodeNodePort),
				WhenToRun:    "AFTER:alcor-createDefaultTable",
				WhereToRun:   "NETWORK",
			},
			{
				Cmd:  "/root/alcor-control-agent/build/bin/AlcorControlAgent",
				Name: "aca-cmd",
				Parameters: []string{
					"-d",
					fmt.Sprintf("-a %s", ncmPortAddress),
					fmt.Sprintf("-p %d", ncmPort),
				},
				ReturnCode: []uint32{
					0,
				},
				ReturnString: []string{""},
				Url:          "",
				WhenToRun:    "INIT",
				WhereToRun:   "AGENT",
			},
		},
	}

	body, _ = json.Marshal(serviceConfig)
	resp, _ = http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/service-config", "application/json", bytes.NewBuffer(body))

	respBodyByte, _ = io.ReadAll(resp.Body)
	respBody = string(respBodyByte[:])
	serviceConfigID := gjson.Get(respBody, "data.id").Str

	computeConfig := entities.ComputeConfig{
		Name:                 "compute-config-1",
		NumberOfComputeNodes: numVhost,
		NumberOfPortPerVm:    1,
		NumberOfVmPerVpc:     10,
		Scheduler:            "SEQUENTIAL",
		VmDeployType:         "UNIFORM",
		VPCInfo: []entities.VPCInfo{
			{
				TenantId:        "123456789",
				ProjectId:       "123456789",
				VpcCidr:         "10.1.0.0/16",
				NumberOfSubnets: 1,
				SubnetInfo: []entities.SubnetInfo{
					{
						SubnetCidr:    "10.1.0.0/16",
						SubnetGateway: "10.1.0.1",
						NumberOfVMs:   10,
					},
				},
			},
		},
	}

	body, _ = json.Marshal(computeConfig)
	resp, _ = http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/compute-config", "application/json", bytes.NewBuffer(body))

	respBodyByte, _ = io.ReadAll(resp.Body)
	respBody = string(respBodyByte[:])
	computeConfigID := gjson.Get(respBody, "data.id").Str

	testConfig := entities.TestConfig{
		Name: "test-config-1",
		Tests: []entities.Test{
			{
				Cmd:        "ping",
				Id:         "string",
				Name:       "pingall",
				Parameters: []string{"string"},
				Script:     "github.com/merak/test/scripts/script1.sh",
				WhereToRun: "AGENT",
				WhenToRun:  "after:VM_CREATED",
			},
		},
	}

	body, _ = json.Marshal(testConfig)
	resp, _ = http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/test-config", "application/json", bytes.NewBuffer(body))

	respBodyByte, _ = io.ReadAll(resp.Body)
	respBody = string(respBodyByte[:])
	testConfigID := gjson.Get(respBody, "data.id").Str

	scenarioConfig := entities.Scenario{
		Name:          "scenario-test-1",
		ProjectId:     "123456789",
		TopologyId:    topoConfigID,
		ServiceConfId: serviceConfigID,
		NetworkConfId: netConfigID,
		ComputeConfId: computeConfigID,
		TestConfId:    testConfigID,
	}

	body, _ = json.Marshal(scenarioConfig)
	resp, _ = http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/scenarios", "application/json", bytes.NewBuffer(body))

	respBodyByte, _ = io.ReadAll(resp.Body)
	respBody = string(respBodyByte[:])
	scenarioConfigID := gjson.Get(respBody, "data.id").Str
	return scenarioConfigID
}

func getServiceAddress(clientset *kubernetes.Clientset, namespace string, serviceName string) (string, int32) {
	service, err := clientset.CoreV1().Services(namespace).Get(context.Background(), serviceName, v1.GetOptions{})
	if err != nil {
		log.Printf("error getting services: %v\n", err)
		os.Exit(1)
	}
	set := labels.Set(service.Spec.Selector)
	listOptions := v1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, _ := clientset.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	nodename := pods.Items[len(pods.Items)-1].Spec.NodeName
	node, _ := clientset.CoreV1().Nodes().Get(context.Background(), nodename, v1.GetOptions{})
	nodePortAddress := node.Status.Addresses[0].Address
	nodePort := service.Spec.Ports[len(service.Spec.Ports)-1].NodePort
	return nodePortAddress, nodePort
}

func deployTopo(scenarioConfigID string, scenarioNodePortAddress string, scenarioNodePort int32, expReady int) {
	deployTopo := entities.ScenarioAction{
		ScenarioId: scenarioConfigID,
		Service: entities.ServiceAction{
			Action:      "DEPLOY",
			ServiceName: "topology",
		},
	}

	body, _ := json.Marshal(deployTopo)
	http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/scenarios/actions", "application/json", bytes.NewBuffer(body))
	checkTopo(scenarioConfigID, scenarioNodePortAddress, scenarioNodePort, expReady)
}

func checkTopo(scenarioConfigID string, scenarioNodePortAddress string, scenarioNodePort int32, expReady int) {
	deployTopo := entities.ScenarioAction{
		ScenarioId: scenarioConfigID,
		Service: entities.ServiceAction{
			Action:      "CHECK",
			ServiceName: "topology",
		},
	}
	body, _ := json.Marshal(deployTopo)

	fmt.Println("Waiting for vhost to be ready.")
	for {
		resp, _ := http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/scenarios/actions", "application/json", bytes.NewBuffer(body))
		respBodyByte, _ := io.ReadAll(resp.Body)
		respBody := string(respBodyByte[:])
		readyNodes, _ := strconv.Atoi(strings.TrimSpace(strings.Split(strings.Split(gjson.Get(respBody, "message").Str, ",")[1], ":")[1]))
		if readyNodes == expReady {
			break
		}
		time.Sleep(time.Second * 2)
		time.Sleep(time.Second * 2)
		fmt.Printf("%d of %d vHosts ready\n", readyNodes, expReady)
	}
	fmt.Println()
}

func deployNetwork(scenarioConfigID string, scenarioNodePortAddress string, scenarioNodePort int32) {
	deployNetwork := entities.ScenarioAction{
		ScenarioId: scenarioConfigID,
		Service: entities.ServiceAction{
			Action:      "DEPLOY",
			ServiceName: "network",
		},
	}

	body, _ := json.Marshal(deployNetwork)
	http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/scenarios/actions", "application/json", bytes.NewBuffer(body))
	// checkNetwork(scenarioConfigID, scenarioNodePortAddress, scenarioNodePort)
}

func checkNetwork(scenarioConfigID string, scenarioNodePortAddress string, scenarioNodePort int32) {
	checkNetwork := entities.ScenarioAction{
		ScenarioId: scenarioConfigID,
		Service: entities.ServiceAction{
			Action:      "CHECK",
			ServiceName: "network",
		},
	}
	body, _ := json.Marshal(checkNetwork)
	resp, _ := http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/scenarios/actions", "application/json", bytes.NewBuffer(body))
	respBodyByte, _ := io.ReadAll(resp.Body)
	respBody := string(respBodyByte[:])
	fmt.Println(respBody)
}

func deployCompute(scenarioConfigID string, scenarioNodePortAddress string, scenarioNodePort int32) {
	deployCompute := entities.ScenarioAction{
		ScenarioId: scenarioConfigID,
		Service: entities.ServiceAction{
			Action:      "DEPLOY",
			ServiceName: "compute",
		},
	}

	body, _ := json.Marshal(deployCompute)
	http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/scenarios/actions", "application/json", bytes.NewBuffer(body))
	checkCompute(scenarioConfigID, scenarioNodePortAddress, scenarioNodePort)
}

func checkCompute(scenarioConfigID string, scenarioNodePortAddress string, scenarioNodePort int32) {
	checkCompute := entities.ScenarioAction{
		ScenarioId: scenarioConfigID,
		Service: entities.ServiceAction{
			Action:      "CHECK",
			ServiceName: "compute",
		},
	}

	var expReady int
	var ready int
	fmt.Println("Waiting for EVM to be ready.")
	for {
		body, _ := json.Marshal(checkCompute)
		resp, _ := http.Post("http://"+scenarioNodePortAddress+":"+strconv.Itoa(int(scenarioNodePort))+"/api/scenarios/actions", "application/json", bytes.NewBuffer(body))
		respBodyByte, _ := io.ReadAll(resp.Body)
		respBody := string(respBodyByte[:])
		ready, _ = strconv.Atoi(strings.TrimSpace(strings.Split(gjson.Get(respBody, "data.vms.0.id").Str, " ")[0]))
		expReady, _ = strconv.Atoi(strings.TrimSpace(strings.Split(gjson.Get(respBody, "data.vms.0.id").Str, " ")[3]))
		if ready == expReady {
			break
		}
		time.Sleep(time.Second * 2)
		fmt.Printf("%d of %d VMs ready\n", ready, expReady)
	}
	fmt.Println("All VMs Ready!!")
}
