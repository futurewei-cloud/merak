package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
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
	ACA_IMAGE = "phudtran/merak-agent:dev"
	OVS_IMAGE = "yanmo96/ovs_only:latest"
	GW_IMAGE  = "yanmo96/aca_build_standard:v2"
	Ctx       = context.Background()

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
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"links": links,
			},
		},
	}
}

func Topo_deploy(k8client *kubernetes.Clientset, topo database.TopologyData) error {

	nodes := topo.Vnodes

	// dynamic client

	config := ctrl.GetConfigOrDie()
	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client %s", err)
	}

	for _, node := range nodes {

		// Create topology class

		err := CreateTopologyClasses(dclient, node.Name, node.Flinks)

		if err != nil {
			return fmt.Errorf("failed to create runtime class %s", err)
		}

		//// create pods
		var newPod *corev1.Pod
		l := make(map[string]string)

		if strings.Contains(node.Name, "vhost") {
			l["Type"] = "vhost"
			newPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   node.Name,
					Labels: l,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "vhost", Image: ACA_IMAGE, Command: []string{"sleep", "100000"}},
					},
				},
			}
		} else if strings.Contains(node.Name, "vswitch") || strings.Contains(node.Name, "tor") {
			l["Type"] = "vswitch"
			var sc corev1.SecurityContext
			pri := true
			sc.Privileged = &pri
			allow_pri := true
			sc.AllowPrivilegeEscalation = &allow_pri
			var capab corev1.Capabilities

			capab.Add = append(capab.Add, "NET_ADMIN")
			capab.Add = append(capab.Add, "SYS_TIME")
			sc.Capabilities = &capab

			newPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   node.Name,
					Labels: l,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "vswitch",
							Image: OVS_IMAGE,
							Args:  []string{"service rsyslog restart; /etc/init.d/openvswitch-switch restart; sleep infinity"},
							// add ovs setup commands
							Command:         []string{"/bin/sh", "-c"},
							SecurityContext: &sc,
						},
					},
				},
			}
		} else if strings.Contains(node.Name, "cgw") {
			l["Type"] = "configgw"

			newPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   node.Name,
					Labels: l,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "cgw", Image: GW_IMAGE, Command: []string{"sleep", "100000"}},
					},
				},
			}

		} else {
			return errors.New("no image for this device, please upload the image before create topology")
		}

		_, err = k8client.CoreV1().Pods("default").Create(Ctx, newPod, metav1.CreateOptions{})

		if err != nil {
			return fmt.Errorf("create pod error %s", err)
		} else {
			err = database.SetValue(node.Name+"-pod", newPod)
			if err != nil {
				log.Fatalf("fail: save topology in DB %s", err)
			}

		}

	}

	return nil

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

	log.Printf("Output from pod: %v", stdout.String())
	log.Printf("Error from pod: %v", stderr.String())

	return stdout.String(), nil

}

func Topo_delete(k8client *kubernetes.Clientset, topo database.TopologyData) error {

	config := ctrl.GetConfigOrDie()
	dclient, err := dynamic.NewForConfig(config)

	if err != nil {
		return fmt.Errorf("failed to create dynamic client %s", err)
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
func Topo_save(k8client *kubernetes.Clientset, topo database.TopologyData) error {
	// check pod status
	topo_id := topo.Topology_id
	pods_status, err := Topo_pod_check(k8client, topo)
	if err != nil {
		return fmt.Errorf("fail to check pods status %s", err)
	}
	fmt.Println(pods_status)

	err_db := database.SetValue(topo_id, topo)
	if err_db != nil {
		return fmt.Errorf("fail to save in db %s", err_db)
	}
	return nil
}

func Topo_pod_check(k8client *kubernetes.Clientset, topo database.TopologyData) (bool, error) {

	s := true

	nodes := topo.Vnodes

	for _, node := range nodes {

		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})

		if err != nil {
			return false, fmt.Errorf("get pod error %s", err)
		}
		if res.Status.Phase != "Running" {
			s = false
		}
	}
	return s, nil
}

// 		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})
// out, err := k8client.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
