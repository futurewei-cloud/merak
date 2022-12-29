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
	"strconv"
	"strings"
	"time"

	constants "github.com/futurewei-cloud/merak/services/common"
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
	SDN_IP   = "sdn-controller.merak.svc.cluster.local"
	SDN_PORT = "6653"
	Ctx      = context.Background()

	topologyClassGVR = schema.GroupVersionResource{
		Group:    "networkop.co.uk",
		Version:  "v1beta1",
		Resource: "topologies",
	}
)

func CreateTopologyClasses(client dynamic.Interface, name string, links []database.Vlink, namespace string) error {
	rc := NewTopologyClass(name, links, namespace)

	_, err := client.Resource(topologyClassGVR).Namespace(namespace).Create(Ctx, rc, metav1.CreateOptions{})

	if err != nil {
		utils.Logger.Error("can't create topologyClass", "create topology class error", err.Error(), "namespace", namespace, "vnode", name)
		return err
	}

	return nil

}

func GetTopologyClasses(client dynamic.Interface, name string, namespace string) error {

	_, err := client.Resource(topologyClassGVR).Namespace(namespace).Get(Ctx, name, metav1.GetOptions{})

	if err != nil {
		utils.Logger.Error("can't get topology class data", "get topologyClass error", err.Error(), "namespace", namespace, "vnode", name)
		return err
	}

	return nil

}

/*Comment: unused function when deleting with namespace*/
// func DeleteTopologyClasses(client dynamic.Interface, name string, namespace string) error {

// 	err := client.Resource(topologyClassGVR).Namespace(namespace).Delete(Ctx, name, metav1.DeleteOptions{})

// 	utils.Logger.Debug("Delete TopologyClass %s", name)

// 	if err != nil {
// 		utils.Logger.Error("failed to create topologyClass", err.Error())
// 		return err
// 	}

// 	return nil

// }

func NewTopologyClass(name string, links []database.Vlink, namespace string) *unstructured.Unstructured {
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

func Topo_deploy(k8client *kubernetes.Clientset, aca_image string, ovs_image string, topo database.TopologyData, aca_parameters string, namespace string) error {
	/*comment gw creation function*/
	// var k8snodes []string

	errs := errors.New("topology deployment fails")
	errs_flag := 0

	nodes := topo.Vnodes

	config := ctrl.GetConfigOrDie()
	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		utils.Logger.Error("fails to create k8s client in topology deployment", "create dynamic client error", err.Error())
		return err
	}

	var vhost_pods_config []*corev1.Pod
	var rack_pods_config []*corev1.Pod
	var vs_pods_config []*corev1.Pod

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

	start_time := time.Now()

	for _, node := range nodes {

		// Create topology class

		err := CreateTopologyClasses(dclient, node.Name, node.Flinks, namespace)

		if err != nil {
			utils.Logger.Error("can't create topology class", "meshnet-cni", err.Error(), "vnode name", node.Name, "namespace", namespace)
			errs_flag = 1
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
		sec = 600000000000
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
							Command:         []string{"/bin/sh", "-c", "/merak-bin/merak-agent " + aca_parameters},
							SecurityContext: &sc,
							Ports: []corev1.ContainerPort{
								{ContainerPort: constants.AGENT_GRPC_SERVER_PORT},
								{ContainerPort: constants.AGENT_PROMETHEUS_PORT},
							},
						},
					},

					Affinity:                      &aff,
					RestartPolicy:                 "OnFailure",
					TerminationGracePeriodSeconds: &grace_period,
					Tolerations:                   tol,
				},
			}

			vhost_pods_config = append(vhost_pods_config, newPod)

		} else if strings.Contains(node.Name, "rack") {

			ovs_set, err0 := ovs_config(topo, node.Name, SDN_IP, SDN_PORT)
			if err0 != nil {
				utils.Logger.Error("fails to configure ovs", " ovs switch controller info error", err0.Error(), "vnode", node.Name)
				errs_flag = 1
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

			rack_pods_config = append(rack_pods_config, newPod)

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

		} else if strings.Contains(node.Name, "vs") || strings.Contains(node.Name, "core") {

			ovs_set, err0 := ovs_config(topo, node.Name, SDN_IP, SDN_PORT)
			if err0 != nil {
				utils.Logger.Error("fails to configure ovs", "ovs switch controller info error", err0.Error(), "vnode", node.Name)
				errs_flag = 1
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
			vs_pods_config = append(vs_pods_config, newPod)

		} else {
			utils.Logger.Error("device type in topology has not been defined yet", "device type", "not defined")
			errs_flag = 1
		}

	}

	elaps0 := time.Since(start_time)
	start0 := time.Now()

	utils.Logger.Info("request DEPLOY", "create topology crd data in K8s (in second)", elaps0)

	for _, newPod := range vs_pods_config {

		_, err_create := k8client.CoreV1().Pods(namespace).Create(Ctx, newPod, metav1.CreateOptions{})
		if err_create != nil {
			utils.Logger.Error("request DEPLOY", "create pod in k8s cluster", err_create.Error(), "namespace", namespace, "pod", newPod.Name)
			errs_flag = 1
		} else {
			err_db := database.SetValue(topo.Topology_id[:5]+":"+newPod.Name, newPod)
			if err_db != nil {
				utils.Logger.Error("request DEPLOY", "can't save topology in DB", err_db.Error(), "topologyid_pod", topo.Topology_id+"_"+newPod.Name)
				errs_flag = 1
			}

		}
	}

	for _, newPod := range rack_pods_config {

		_, err_create := k8client.CoreV1().Pods(namespace).Create(Ctx, newPod, metav1.CreateOptions{})
		if err_create != nil {
			utils.Logger.Error("can't create pod", "error", err_create.Error(), "pod", newPod.Name, "namespace", namespace)
			errs_flag = 1
		} else {
			err_db := database.SetValue(topo.Topology_id[:5]+":"+newPod.Name, newPod)
			if err_db != nil {
				utils.Logger.Error("can't save topology in DB", "error", err_db.Error(), "topologyid_pod", topo.Topology_id+"_"+newPod.Name)
				errs_flag = 1
			}

		}
	}

	for _, newPod := range vhost_pods_config {

		_, err_create := k8client.CoreV1().Pods(namespace).Create(Ctx, newPod, metav1.CreateOptions{})
		if err_create != nil {
			utils.Logger.Error("can't create pod", "create pod error", err_create.Error(), "namespace", namespace, "topologyid_pod", topo.Topology_id+"_"+newPod.Name)
			errs_flag = 1
		} else {
			err_db := database.SetValue(topo.Topology_id[:5]+":"+newPod.Name, newPod)
			if err_db != nil {
				utils.Logger.Error("can't save topology in DB", "save topology in DB error", err_db.Error())
				errs_flag = 1
			}

		}
	}

	elaps1 := time.Since(start0)

	utils.Logger.Info("request DEPLOY", "create pod in K8s (in second)", elaps1)

	if errs_flag == 1 {
		return errs
	} else {
		return nil
	}

}

func ovs_config(topo database.TopologyData, node_name string, sdn_ip string, sdn_port string) (string, error) {

	nodes := topo.Vnodes
	ovs_set := "ovs-vsctl add-br br0; ovs-vsctl set-controller br0 tcp:" + sdn_ip + ":" + sdn_port + "; "

	for _, node := range nodes {
		if node.Name == node_name {
			for _, n := range node.Nics {
				ovs_set = ovs_set + "ovs-vsctl add-port br0 " + n.Intf + "; "
			}

		}

	}

	return ovs_set, nil

}

func Pod_query(k8client *kubernetes.Clientset, pod *corev1.Pod, cmd []string, namespace string) (string, error) {

	err := errors.New("fails to query pods in K8s cluster")

	req := k8client.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).Namespace(namespace).SubResource("exec") // .Param("container", containerName)
	scheme := runtime.NewScheme()
	err1 := corev1.AddToScheme(scheme)
	if err1 != nil {
		utils.Logger.Error("pod query fail: ", "addtoscheme", err1.Error())
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

	config, err2 := utils.K8sConfig()
	if err2 != nil {
		utils.Logger.Error("fail: k8sconfig", "k8sconfig error", err2.Error())
	}

	exec, err3 := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err3 != nil {
		utils.Logger.Error("fail: newspdyexecutor", "k8sconfig POST error", err3.Error())
	}
	var stdout, stderr bytes.Buffer
	err4 := exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err4 != nil {
		utils.Logger.Error("fail: stream", "stream error", err4.Error())
	}

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return stdout.String(), err
	} else {
		return stdout.String(), nil
	}

}

func Topo_delete(k8client *kubernetes.Clientset, topo database.TopologyData, namespace string) error {

	err_del_db := database.DeleteAllValuesWithKeyPrefix(topo.Topology_id[:5])

	if err_del_db != nil {
		utils.Logger.Error("can't delete topology info in DB", "topology delete in DB error", err_del_db.Error())
		return err_del_db
	}

	k8client.CoreV1().Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{})

	return nil
}

// save topology to redis
func Topo_save(topo database.TopologyData) error {
	// check pod status
	topo_id := topo.Topology_id

	err_db := database.SetValue(topo_id[:5], topo)
	if err_db != nil {
		utils.Logger.Error("fail to save topology in DB", "topology save in DB error", err_db.Error())
		return err_db
	}
	return nil
}
