package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
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

	service, err := clientset.CoreV1().Services("merak").Get(context.Background(), "scenario-manager-service", v1.GetOptions{})
	if err != nil {
		log.Printf("error getting services: %v\n", err)
		os.Exit(1)
	}
	set := labels.Set(service.Spec.Selector)
	listOptions := v1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, _ := clientset.CoreV1().Pods("merak").List(context.Background(), listOptions)
	nodename := pods.Items[len(pods.Items)-1].Spec.NodeName
	node, _ := clientset.CoreV1().Nodes().Get(context.Background(), nodename, v1.GetOptions{})
	nodePortAddress := node.Status.Addresses[0].Address
	nodePort := service.Spec.Ports[len(service.Spec.Ports)-1].NodePort
	fmt.Printf("Found Scenario-Manager\nservice ip %s\nservice port %d\n", nodePortAddress, nodePort)

	topoConfig := entities.TopologyConfig{
		Name:             "Top1",
		NumberOfGateways: 0,
		NumberOfVhosts:   5,
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
	resp, _ := http.Post("http://"+nodePortAddress+":"+strconv.Itoa(int(nodePort))+"/api/topologies", "application/json", bytes.NewBuffer(body))
	log.Println(resp)
}
