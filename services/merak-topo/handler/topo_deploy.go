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
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"

	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/remotecommand"
)

var (
	// ACA_IMAGE = "meraksim/merak-agent:311f5f6c"
	// OVS_IMAGE = "yanmo96/ovs_only:latest"
	RYU_IP   = "ryu.merak.svc.cluster.local"
	RYU_PORT = "6653"
	Ctx      = context.Background()

	namespace        = "default"
	topologyClassGVR = schema.GroupVersionResource{
		Group:    "networkop.co.uk",
		Version:  "v1beta1",
		Resource: "topologies",
	}
)

func CreateTopologyClasses(client dynamic.Interface, name string, links []database.Vlink) error {
	rc := NewTopologyClass(name, links)

	_, err := client.Resource(topologyClassGVR).Namespace(namespace).Create(Ctx, rc, metav1.CreateOptions{})

	// log.Printf("Creating TopologyClass %s", name)
	// log.Printf("topology class details %s", rc)

	if err != nil {
		return fmt.Errorf("failed to create topologyClass %s", err)
	}

	return nil

}

func GetTopologyClasses(client dynamic.Interface, name string) error {

	_, err := client.Resource(topologyClassGVR).Namespace(namespace).Get(Ctx, name, metav1.GetOptions{})

	log.Printf("Get TopologyClass %s", name)

	if err != nil {
		return fmt.Errorf("failed to create topologyClass %s", err)
	}

	// fmt.Printf("Get %v topology data: %v", name, data)

	return nil

}

func DeleteTopologyClasses(client dynamic.Interface, name string) error {

	err := client.Resource(topologyClassGVR).Namespace(namespace).Delete(Ctx, name, metav1.DeleteOptions{})

	log.Printf("Delete TopologyClass %s", name)

	if err != nil {
		return fmt.Errorf("failed to create topologyClass %s", err)
	}

	return nil

}

func NewTopologyClass(name string, links []database.Vlink) *unstructured.Unstructured {
	var clinks []map[string]interface{}
	for _, link := range links {
		config_clink := map[string]interface{}{
			"uid":        link.Uid,
			"peer_pod":   link.Peer_pod,
			"local_intf": link.Local_intf,
			"local_ip":   link.Local_ip,
			"peer_intf":  link.Peer_intf,
			"peer_ip":    link.Peer_ip,
		}
		clinks = append(clinks, config_clink)
	}

	out := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Topology",
			"apiVersion": "networkop.co.uk/v1beta1",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"links": clinks,
			},
		},
	}
	return out
}

func Topo_deploy(k8client *kubernetes.Clientset, aca_image string, ovs_image string, topo database.TopologyData) error {
	/*comment gw creation function*/
	// var k8snodes []string

	nodes := topo.Vnodes

	config := ctrl.GetConfigOrDie()
	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("fails to create dynamic client %s", err)
	}

	/*comment gw creation function*/
	// k_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	// if err1 != nil {
	// 	return fmt.Errorf("fails to query k8s nodes info %s", err1)
	// }

	// for _, s := range k_nodes.Items {
	// 	if s.Spec.Taints == nil {
	// 		k8snodes = append(k8snodes, s.Name)
	// 	}
	// }

	for _, node := range nodes {

		// Create topology class

		err := CreateTopologyClasses(dclient, node.Name, node.Flinks)

		if err != nil {
			return fmt.Errorf("failed to create runtime class %s", err)
		}

		interface_num := len(node.Nics) + 1

		init_container := corev1.Container{
			Name:            "init-" + node.Name,
			Image:           "networkop/init-wait:latest",
			ImagePullPolicy: "IfNotPresent",
			Args:            []string{strconv.Itoa(interface_num), "0"},
		}

		init_containers := []corev1.Container{}
		init_containers = append(init_containers, init_container)

		var grace_period = int64(0)

		//// create pods
		var newPod *corev1.Pod
		l := make(map[string]string)
		l["App"] = node.Name
		l["Topo"] = "topology"

		anno := make(map[string]string)
		anno["linkerd.io/inject"] = "enabled"

		var sc corev1.SecurityContext
		pri := true
		sc.Privileged = &pri
		allow_pri := true
		sc.AllowPrivilegeEscalation = &allow_pri
		var capab corev1.Capabilities

		capab.Add = append(capab.Add, "NET_ADMIN")
		capab.Add = append(capab.Add, "SYS_TIME")
		sc.Capabilities = &capab

		var tol []corev1.Toleration
		var t1 corev1.Toleration
		var t2 corev1.Toleration

		var sec int64
		t1.Key = "node.kubernetes.io/not-ready"
		t2.Key = "node.kubernetes.io/unreachable"
		t1.Effect = "NoExecute"
		t2.Effect = "NoExecute"
		sec = 6000
		t1.TolerationSeconds = &sec
		t2.TolerationSeconds = &sec

		tol = append(tol, t1)
		tol = append(tol, t2)

		var aff corev1.Affinity

		if strings.Contains(node.Name, "vhost") {
			l["Type"] = "vhost"
			newPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   node.Name,
					Labels: l,
				},
				Spec: corev1.PodSpec{
					InitContainers: init_containers,
					Containers: []corev1.Container{
						{
							Name:            "vhost",
							Image:           aca_image,
							ImagePullPolicy: "IfNotPresent",
							// Args:            []string{},
							Command:         []string{"/bin/sh", "-c", "/merak-bin/merak-agent 172.31.115.6 30014"},
							SecurityContext: &sc,
						},
					},

					Affinity:                      &aff,
					RestartPolicy:                 "OnFailure",
					TerminationGracePeriodSeconds: &grace_period,
					Tolerations:                   tol,
				},
			}
		} else if strings.Contains(node.Name, "rack") || strings.Contains(node.Name, "vs") || strings.Contains(node.Name, "core") {

			ovs_set, err0 := ovs_config(topo, node.Name, RYU_IP, RYU_PORT)
			if err0 != nil {
				return fmt.Errorf("fails to get ovs switch controller info %s", err0)
			}

			l["Type"] = "vswitch"

			newPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   node.Name,
					Labels: l,
				},
				Spec: corev1.PodSpec{
					InitContainers: init_containers,
					Containers: []corev1.Container{
						{
							Name:            "vswitch",
							Image:           ovs_image,
							ImagePullPolicy: "IfNotPresent",
							Args:            []string{"service rsyslog restart; /etc/init.d/openvswitch-switch restart; " + ovs_set + "sleep infinity"},
							Command:         []string{"/bin/sh", "-c"},
							SecurityContext: &sc,
						},
					},
					RestartPolicy:                 "OnFailure",
					TerminationGracePeriodSeconds: &grace_period,
					Tolerations:                   tol,
				},
			}
			/*comment gw creation function*/
			// } else if strings.Contains(node.Name, "cgw") {
			// 	l["Type"] = "configgw"

			// 	log.Printf("assign cgw to node %v", k8snodes[0])

			// 	newPod = &corev1.Pod{
			// 		ObjectMeta: metav1.ObjectMeta{
			// 			Name:   node.Name,
			// 			Labels: l,
			// 		},
			// 		Spec: corev1.PodSpec{
			// 			InitContainers: init_containers,
			// 			Containers: []corev1.Container{
			// 				{
			// 					Name:            "cgw",
			// 					Image:           GW_IMAGE,
			// 					ImagePullPolicy: "IfNotPresent",
			// 					Command:         []string{"sleep", "100000"},
			// 					SecurityContext: &sc,
			// 				},
			// 			},
			// 			RestartPolicy:                 "OnFailure",
			// 			NodeName:                      k8snodes[0],
			// 			TerminationGracePeriodSeconds: &grace_period,
			// 			Tolerations:                   tol,
			// 		},
			// 	}
			// 	if len(k8snodes) > 1 {
			// 		k8snodes = k8snodes[1:]
			// 		log.Printf("unassigned nodes %v", k8snodes)

			// 	}

		} else {
			return errors.New("no image for this device, please upload the image before create topology")
		}

		_, err_create := k8client.CoreV1().Pods("default").Create(Ctx, newPod, metav1.CreateOptions{})

		if err_create != nil {
			return fmt.Errorf("create pod error %s", err_create)
		} else {
			err_db := database.SetValue(topo.Topology_id+":"+node.Name, newPod)
			if err_db != nil {
				log.Fatalf("fail: save topology in DB %s", err_db)
			}

		}

	}

	return nil

}

func ovs_config(topo database.TopologyData, node_name string, ryu_ip string, ryu_port string) (string, error) {

	nodes := topo.Vnodes
	ovs_set := "ovs-vsctl add-br br0; ovs-vsctl set-controller br0 tcp:" + ryu_ip + ":" + ryu_port + "; "

	for _, node := range nodes {
		if node.Name == node_name {
			for _, n := range node.Nics {
				ovs_set = ovs_set + "ovs-vsctl add-port br0 " + n.Intf + "; "
			}

		}

	}

	log.Println(ovs_set)

	return ovs_set, nil

}

func Pod_query(k8client *kubernetes.Clientset, pod *corev1.Pod, cmd []string) (string, error) {

	req := k8client.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).Namespace("default").SubResource("exec") // .Param("container", containerName)
	scheme := runtime.NewScheme()
	if err1 := corev1.AddToScheme(scheme); err1 != nil {
		return " ", fmt.Errorf("fail: addtoscheme %s", err1.Error())
	}
	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Stdin:  false,
		Stdout: true,
		Stderr: true,
		TTY:    false,

		Container: pod.Spec.Containers[0].Name,
		Command:   cmd,
	}, parameterCodec)

	config, err_config := utils.K8sConfig()
	if err_config != nil {
		return " ", fmt.Errorf("fail: k8sconfig %s", err_config.Error())
	}

	exec, err2 := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err2 != nil {
		return " ", fmt.Errorf("fail: newspdyexecutor %s", err2.Error())
	}
	var stdout, stderr bytes.Buffer
	err3 := exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err3 != nil {
		return " ", fmt.Errorf("fail: stream %s", err3.Error())
	}

	return stdout.String(), nil

}

func Topo_delete(k8client *kubernetes.Clientset, topo database.TopologyData) error {

	config := ctrl.GetConfigOrDie()
	dclient, err := dynamic.NewForConfig(config)

	if err != nil {
		return fmt.Errorf("failed to create dynamic client %s", err)
	}

	err_del_db := database.DeleteAllValuesWithKeyPrefix(topo.Topology_id)

	if err_del_db != nil {
		return fmt.Errorf("failed to delete topology info %s", err_del_db)
	}

	for _, node := range topo.Vnodes {

		err_del := k8client.CoreV1().Pods("default").Delete(Ctx, node.Name, metav1.DeleteOptions{})

		if err_del != nil {
			return fmt.Errorf("delete pod container error %s", err_del)
		}

		err_del_t := DeleteTopologyClasses(dclient, node.Name)
		if err_del_t != nil {
			return fmt.Errorf("delete pod topology error %s", err_del_t)
		}

	}
	return nil
}

// save topology to redis
func Topo_save(topo database.TopologyData) error {
	// check pod status
	topo_id := topo.Topology_id

	err_db := database.SetValue(topo_id, topo)
	if err_db != nil {
		return fmt.Errorf("fail to save in db %s", err_db)
	}
	return nil
}
