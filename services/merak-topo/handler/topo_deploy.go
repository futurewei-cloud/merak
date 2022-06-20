package handler

import (
	"context"
	"fmt"
	
	"strings"
	"flag"
	

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	ACA_IMAGE = "phudtran/aca:latest"
	OVS_IMAGE = "yanmo96/ovs_only:latest"
	GW_IMAGE  = "yanmo96/aca_build_standard:v2"

	DEFAULT_NAMESPACE = "default"
)

func config_cnode(node database.Vnode, topo_id string) Cnode {
	var config_node Cnode
	var meta Cmetadata
	var lab LabelStruct

	config_node.ApiVersion = "networkop.co.uk/v1beta1"
	config_node.Kind = "Topology"
	meta.Name = node.Name

	lab.Topo = topo_id

	meta.labels = lab

	config_node.Metadata = meta

	return config_node

}

func config_clink(link database.Vlink, topo_id string) ConfigClink {

	var config_clink ConfigClink

	config_clink.Peer_pod = strings.Split(link.Name, ":")[2]
	config_clink.Local_pod = strings.Split(link.Name, ":")[1]
	config_clink.Local_intf = link.Src.Intf
	config_clink.Local_ip = link.Src.Ip
	config_clink.Peer_intf = link.Dst.Intf
	config_clink.Peer_ip = link.Dst.Ip

	return config_clink
}

func Topo_deploy(topo database.TopologyData) {

	// topo_id := topo.Topology_id

	// fmt.Printf("topo_id %v", topo_id)

	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})

	config, err := kubeconfig.ClientConfig()

	if err != nil {
		panic(err)
	}

	clientset := kubernetes.NewForConfigOrDie(config)

	modeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		panic(err)
	}
	for _, n := range modeList.Items {
		fmt.Printf("K8s Cluster Node: %v \n", n.Name)
	}

	// k8s config
	

	kube_config := flag.String("kubeconfig", "~/.kube/config", "location to the kubeconfig file")
	config1, err := clientcmd.BuildConfigFromFalgs("",*kube_config)
	if err != nil{
		//handle err
	}
	fmt.Println(config1)




	// configure nodes in topology
	// var node_links []ConfigClink
	// var cm_nodes []Cnode
	// for _, link := range topo.Vlinks {
	// 	node_link := config_clink(link, topo_id)
	// 	node_links = append(node_links, node_link)
	// }

	// for _, node := range topo.Vnodes {
	// 	cm_node := config_cnode(node, topo_id)
	// 	uid := 0
	// 	var cm_node_links []Clink

	// 	fmt.Println(node.Name)
	// 	for _, nlink := range node_links {
	// 		// fmt.Println(nlink.Local_pod)
	// 		var clink Clink
	// 		if nlink.Local_pod == strings.Split(node.Name, ":")[0] {

	// 			clink.Peer_pod = nlink.Peer_pod
	// 			clink.Local_intf = nlink.Local_intf
	// 			clink.Local_ip = nlink.Local_ip
	// 			clink.Peer_intf = nlink.Peer_intf
	// 			clink.Peer_ip = nlink.Peer_ip
	// 			clink.Uid = strconv.FormatInt(int64(uid), 10)
	// 			uid = uid + 1
	// 			cm_node_links = append(cm_node_links, clink)
	// 			// fmt.Println("*******")
	// 			// fmt.Println(clink)

	// 		} else if nlink.Peer_pod == strings.Split(node.Name, ":")[0] {

	// 			clink.Peer_intf = nlink.Local_intf
	// 			clink.Peer_ip = nlink.Local_ip
	// 			clink.Local_intf = nlink.Peer_intf
	// 			clink.Local_ip = nlink.Peer_ip
	// 			clink.Uid = strconv.FormatInt(int64(uid), 10)
	// 			uid = uid + 1
	// 			cm_node_links = append(cm_node_links, clink)
	// 			// fmt.Println("&&&&&&&")
	// 			// fmt.Println(clink.Uid)

	// 		}

	// 	}
	// 	cm_node.Spec.Links = cm_node_links
	// 	fmt.Println("=============")
	// 	fmt.Println(cm_node)
	// 	cm_nodes = append(cm_nodes, cm_node)
	// }

	








	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod3",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test1box", Image: "busybox:latest", Command: []string{"sleep", "100000"}},
			},
		},
	}

	pod, err := clientset.CoreV1().Pods("default").Create(context.Background(), newPod, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
	fmt.Println(pod)

	newPod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod2",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test1box", Image: "busybox:latest", Command: []string{"sleep", "100000"}},
			},
		},
	}

	pod2, err := clientset.CoreV1().Pods("default").Create(context.Background(), newPod2, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
	fmt.Println(pod2)

}


// func config (body map[string]string ) error{
// 	configMap := &corev1.ConfigMap{
// 		Metadata: &metav1.ObjectMeta{
// 			Name:      k8s.String("my-configmap"),
// 			Namespace: k8s.String("my-namespace"),
// 		},
// 		Data: map[string]string{"hello": "world"},
// 	}

// 	if err := client.Create(ctx, configMap); err != nil {
// 		// handle error
// 	}

// 	configMap.Data = body

// 	if err := client.Update(ctx, configMap); err != nil {
// 		// handle error
// 	}

// 	if err := client.Delete(ctx, configMap); err != nil {
// 		// handle error
// 	}
// }

// cmap {
// 	'apiVersion': 'networkop.co.uk/v1beta1',
// 	'kind': 'Topology',
// 	'metadata': {
// 		'name': 'ovs-0',
// 		'labels': {'topo': 'demo_v2aca'}
// 		},
// 	'spec': {
// 		'links': [
// 			{
// 				'uid': 0,
// 				'local_intf': 'eth1',
// 				'local_ip': '',
// 				'peer_intf': 'eth1',
// 				'peer_pod': 'aca-0',
// 				'peer_ip': '10.99.0.1/24'
// 			},
// 			{
// 				'uid': 1,
// 				'local_intf': 'eth2',
// 				'local_ip': '',
// 				'peer_intf': 'eth1',
// 				'peer_pod': 'aca-1',
// 				'peer_ip': '10.99.0.2/24'
// 			},
// 			{
// 				'uid': 2,
// 				'local_intf': 'eth3',
// 				'local_ip': '',
// 				'peer_intf': 'eth1',
// 				'peer_pod': 'aca-2',
// 				'peer_ip': '10.99.0.3/24'
// 			},
// 			{
// 				'uid': 3,
// 				'local_intf': 'eth4',
// 				'local_ip': '',
// 				'peer_intf': 'eth1',
// 				'peer_pod': 'aca-3',
// 				'peer_ip': '10.99.0.4/24'
// 			}
// 		]}}
