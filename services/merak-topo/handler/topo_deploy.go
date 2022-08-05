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
	// logrus "github.com/sirupsen/logrus"

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
	// ACA_IMAGE = "phudtran/aca:latest"
	// ACA_IMAGE = "yanmo96/aca_build_standard:v3"
	ACA_IMAGE = "phudtran/merak-agent:dev"
	// ACA_IMAGE = "cjchung4849/aca:dev-268"
	// ACA_IMAGE = "cjchung4849/aca:p.268"
	OVS_IMAGE = "yanmo96/ovs_only:latest"
	GW_IMAGE  = "yanmo96/aca_build_standard:v2"
	// GW_IMAGE = "yanmo96/ym-gateway:latest"
	// RYU_IP   = "10.213.43.111"
	RYU_IP   = "ryu.default.svc.cluster.local"
	RYU_PORT = "6653"
	Ctx      = context.Background()

	namespace        = "default"
	topologyClassGVR = schema.GroupVersionResource{
		Group:    "networkop.co.uk",
		Version:  "v1beta1",
		Resource: "topologies",
	}
)

func CreateTopologyClasses(client dynamic.Interface, name string, links []map[string]interface{}) error {
	rc := NewTopologyClass(name, links)

	_, err := client.Resource(topologyClassGVR).Namespace(namespace).Create(Ctx, rc, metav1.CreateOptions{})

	log.Printf("Creating TopologyClass %s", name)

	if err != nil {
		return fmt.Errorf("failed to create topologyClass %s", err)
	}

	return nil

}

func GetTopologyClasses(client dynamic.Interface, name string) error {

	data, err := client.Resource(topologyClassGVR).Namespace(namespace).Get(Ctx, name, metav1.GetOptions{})

	log.Printf("Get TopologyClass %s", name)

	if err != nil {
		return fmt.Errorf("failed to create topologyClass %s", err)
	}

	fmt.Printf("Get %v topology data: %v", name, data)

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

func NewTopologyClass(name string, links []map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Topology",
			"apiVersion": "networkop.co.uk/v1beta1",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"links": links,
			},
		},
	}
}

func Topo_deploy(k8client *kubernetes.Clientset, topo database.TopologyData) error {

	var k8snodes []string

	nodes := topo.Vnodes

	// dynamic client

	config := ctrl.GetConfigOrDie()
	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("fails to create dynamic client %s", err)
	}

	k_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err1 != nil {
		return fmt.Errorf("fails to query k8s nodes info %s", err1)
	}

	for _, s := range k_nodes.Items {
		if s.Spec.Taints == nil {
			k8snodes = append(k8snodes, s.Name)
		}
	}

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
					// Annotations: anno,
				},
				Spec: corev1.PodSpec{
					InitContainers: init_containers,
					Containers: []corev1.Container{
						{
							Name:            "vhost",
							Image:           ACA_IMAGE,
							ImagePullPolicy: "IfNotPresent",
							// Args:            []string{"service rsyslog restart; /etc/init.d/openvswitch-switch restart; sleep infinity"},
							// Command:         []string{"/bin/sh", "-c"},
							SecurityContext: &sc,
						},
					},

					Affinity:                      &aff,
					RestartPolicy:                 "OnFailure",
					TerminationGracePeriodSeconds: &grace_period,
					Tolerations:                   tol,
				},
			}
		} else if strings.Contains(node.Name, "vswitch") || strings.Contains(node.Name, "core") || strings.Contains(node.Name, "ovs") {

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
							Image:           OVS_IMAGE,
							ImagePullPolicy: "IfNotPresent",
							Args:            []string{"service rsyslog restart; /etc/init.d/openvswitch-switch restart; " + ovs_set + "sleep infinity"},
							Command:         []string{"/bin/sh", "-c"},
							SecurityContext: &sc,
						},
					},
					RestartPolicy: "OnFailure",
					// Affinity: ,
					TerminationGracePeriodSeconds: &grace_period,
					Tolerations:                   tol,
				},
			}
		} else if strings.Contains(node.Name, "cgw") {
			l["Type"] = "configgw"

			log.Printf("assign cgw to node %v", k8snodes[0])

			newPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   node.Name,
					Labels: l,
				},
				Spec: corev1.PodSpec{
					InitContainers: init_containers,
					Containers: []corev1.Container{
						{
							Name:            "cgw",
							Image:           GW_IMAGE,
							ImagePullPolicy: "IfNotPresent",
							Command:         []string{"sleep", "100000"},
							SecurityContext: &sc,
						},
					},
					RestartPolicy:                 "OnFailure",
					NodeName:                      k8snodes[0],
					TerminationGracePeriodSeconds: &grace_period,
					Tolerations:                   tol,
				},
			}
			if len(k8snodes) > 1 {
				k8snodes = k8snodes[1:]
				log.Printf("unassigned nodes %v", k8snodes)

			}

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
		// Container: containerName,
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

	// log.Printf("Output from pod: %v", stdout.String())
	// log.Printf("Error from pod: %v", stderr.String())

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

		// err_del_pod := database.Del(topo.Topology_id + "-" + node.Name + "-pod")
		// if err_del_pod != nil {
		// 	return fmt.Errorf("failed to delete pod config %s", err_del_pod)
		// }

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

// func Topo_pod_check(k8client *kubernetes.Clientset, topo database.TopologyData) (*corev1.PodStatus, error) {

// 	s := true

// 	nodes := topo.Vnodes

// 	for _, node := range nodes {

// 		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})

// 		if err != nil {
// 			return nil, fmt.Errorf("get pod error %s", err)
// 		}
// 		if res.Status.Phase != "Running" {
// 			s = false
// 		}
// 	}
// 	return s, nil
// }

// 		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})
// out, err := k8client.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
