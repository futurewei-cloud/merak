package handler

import (
	"context"
	"encoding/json"
	"fmt"

	// "flag"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

var (
	ACA_IMAGE = "phudtran/aca:latest"
	OVS_IMAGE = "yanmo96/ovs_only:latest"
	GW_IMAGE  = "yanmo96/aca_build_standard:v2"

	DEFAULT_NAMESPACE = "default"
)

func config_cnode(node database.Vnode, topo_id string) *corev1.ConfigMap {

	// config_node := make(map[string]interface{})
	// meta := make(map[string]interface{})
	// label := make(map[string]string)
	// spec := make(map[string]string)

	// config_node["apiversion"] = "networkop.co.uk/v1beta1"
	// config_node["kind"] = "Topology"

	// meta["name"] = node.Name
	// label["topo"] = topo_id
	// meta["labels"] = label
	// config_node["metadata"] = meta

	out, err := json.Marshal(node.Flinks)
	if err != nil {
		fmt.Printf("config_node to json marsh error %s", err)
	}

	cnode_body := string(out)

	// config_node["spec"] = spec

	conf := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   node.Name,
			Labels: map[string]string{"topo": topo_id}},
		Data: map[string]string{"links": cnode_body},
	}

	return conf

}

// node -- full links--- how to set this one???
// func upload_to_k8s (node database.Vnode, topo_id string) error {
// 	config_node := config_cnode(node, topo_id)
//
// 	out, err := json.Marshal(config_node)
//     if err != nil {
//         panic (err)
//     }

//    cnode_body= string(out)

// }

func Topo_deploy(k8client *kubernetes.Clientset, topo database.TopologyData) error {

	topo_id := topo.Topology_id
	nodes := topo.Vnodes
	// group := "networkop.co.uk"
	// version := "version"
	// namespace := "default"
	// plural := "topologies"

	for _, node := range nodes {
		cm_node := config_cnode(node, topo_id)

		// fmt.Printf("===========create namespace for node %v \n", node.Name)

		// fmt.Println(cm_node)

		// out, err := json.Marshal(cm_node)
		// if err != nil {
		// 	panic(err)
		// }

		// cnode_body := string(out)

		// _, _, err :=k8client.RESTClient(). CreateNamespacedCustomObject(context.Background(), group, version, namespace, plural, cm_node)
		// if err != nil {
		// 	panic(err)
		// }

		nsName := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mdefault",
			},
		}

		_, err := k8client.CoreV1().Namespaces().Create(context.Background(), nsName, metav1.CreateOptions{})

		if err != nil {
			return fmt.Errorf("Create k8s namespace error %s", err)
		}

		_, err1 := k8client.CoreV1().ConfigMaps("mdefault").Create(context.Background(), cm_node, metav1.CreateOptions{})

		if err1 != nil {
			return fmt.Errorf("Create k8s namespace error %s", err1)
		}

		newPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: cm_node.Name,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "test1box", Image: "busybox:latest", Command: []string{"sleep", "100000"}},
				},
			},
		}

		pod, err2 := k8client.CoreV1().Pods("default").Create(context.Background(), newPod, metav1.CreateOptions{})

		if err != nil {
			return fmt.Errorf("create pod error %s", err2)
		}
		fmt.Println(pod)

	}

	return nil

}

// for _, device in range devices{
// 	err:= uploade_to_k8s(device, topo_id)
// 	if err != nil{
// 		//handle err
// 	}
// }

//

// 	newPod2 := &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "test-pod2",
// 		},
// 		Spec: corev1.PodSpec{
// 			Containers: []corev1.Container{
// 				{Name: "test1box", Image: "busybox:latest", Command: []string{"sleep", "100000"}},
// 			},
// 		},
// 	}

// 	pod2, err := clientset.CoreV1().Pods("default").Create(context.Background(), newPod2, metav1.CreateOptions{})

// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(pod2)

// }

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
