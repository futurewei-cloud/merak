package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

var (
	ACA_IMAGE = "phudtran/aca:latest"
	OVS_IMAGE = "yanmo96/ovs_only:latest"
	GW_IMAGE  = "yanmo96/aca_build_standard:v2"
	Ctx       = context.Background()

	DEFAULT_NAMESPACE = "default"
)

func config_cnode(node database.Vnode, topo_id string) *corev1.ConfigMap {

	out, err := json.Marshal(node.Flinks)
	if err != nil {
		fmt.Printf("config_node to json marsh error %s", err)
	}

	cnode_body := string(out)
	fmt.Println(cnode_body)

	conf := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Topology",
			APIVersion: "networkop.co.uk/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: "default",
			Labels:    map[string]string{"topo": topo_id}},
		Data: map[string]string{"links": cnode_body},
	}

	return conf

}

func crd_cnode(node database.Vnode) string {

	out, err := json.Marshal(node.Flinks)
	if err != nil {
		fmt.Printf("config_node to json marsh error %s", err)
	}

	cnode_body := string(out)

	// crd_node := &apiextensionsv1beta1.CustomResourceDefinition{
	// 	TypeMeta: metav1.TypeMeta{
	// 		Kind:       "Topology",
	// 		APIVersion: "networkop.co.uk/v1beta1",
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      "topos.networkop.co.uk",
	// 		Namespace: "default",
	// 		SelfLink:  cnode_body,
	// 	},
	// 	Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
	// 		Group:   "networkop.co.uk",
	// 		Version: "version",
	// 		Scope:   apiextensionsv1beta1.NamespaceScoped,
	// 		Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
	// 			Plural:     "topos",
	// 			Singular:   "topology",
	// 			Kind:       "Topology",
	// 			ShortNames: []string{},
	// 		},
	// 		// Namespace: "default",
	// 		//
	// 		// body: cnode_body,
	// 	},
	// ApiextensionsV1beta1.CustomResourceDefinitionSpec{

	// }

	return cnode_body
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

// func getClusterConfig() *rest.Config {
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		glog.Fatal(err.Error())
// 	}
// 	return config
// }

// func getRestClient() *rest.RESTClient {

// 	cfg := getClusterConfig()

// 	gv := schema.GroupVersion{Group: "networkop.co.uk", Version: "version"}
// 	cfg.GroupVersion = &gv
// 	cfg.APIPath = "/apis/networkop.co.uk/v1beta1" // you can verify the path from step 2

// 	var Scheme = runtime.NewScheme()
// 	var Codecs = serializer.NewCodecFactory(Scheme)
// 	cfg.NegotiatedSerializer = Codecs.WithoutConversion()

// 	restClient, err := rest.RESTClientFor(cfg)

// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	return restClient
// }

func Topo_deploy(k8client *kubernetes.Clientset, topo database.TopologyData) error {

	topo_id := topo.Topology_id
	nodes := topo.Vnodes
	// group := "networkop.co.uk"
	// version := "version"
	// namespace := "default"
	// plural := "topologies"

	// kubeconfig := flag.String("kubeconfig", "~/.kube/config", "location to your kubeconfig file")

	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	return fmt.Errorf(err.Error())
	// }

	// data, err := k8client.RESTClient().Get().AbsPath("/apis/networking.k8s.io/v1beta1").DoRaw(context.TODO())

	// if err != nil {
	// 	fmt.Errorf("RESTclient error %s", err)
	// }

	for _, node := range nodes {
		// config_node := crd_cnode(node)

		//////////////// configure node

		// cmap := &corev1.ConfigMap{

		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Name: node.Name + "-config"},

		// 	Data: map[string]string{"config": ""},
		// }

		// _, err1 := k8client.CoreV1().ConfigMaps("default").Create(Ctx, cmap, metav1.CreateOptions{})
		// if err1 != nil {
		// 	return fmt.Errorf("config map create error %s", err1)
		// }

		/////////// Get CRD
		_, err := k8client.RESTClient().Get().AbsPath("/apis/networkop.co.uk/v1beta1").DoRaw(context.TODO())
		if err != nil {
			return fmt.Errorf("get restclient error %s", err)
		}

		// fmt.Println("Get Restclient data %v", data)
		// fmt.Println("Get node links %v", node.Flinks)

		///////////post to CRD

		body := make(map[string]interface{})
		body["apiVersion"] = "networkop.co.uk/vibeta1"
		body["kind"] = "Topology"

		meta := make(map[string]interface{})
		meta["name"] = node.Name
		meta["labels"] = topo_id

		body["metadata"] = meta

		spec_links := make(map[string]interface{})
		spec_links["links"] = node.Flinks

		body["spec"] = spec_links

		create_body := make(map[string]interface{})

		create_body["group"] = "networkop.co.uk"
		create_body["version"] = "v1beta1"
		create_body["namespace"] = "default"
		create_body["plural"] = "topologies"
		create_body["body"] = body

		data, err := json.Marshal(create_body)

		if err != nil {
			return fmt.Errorf("create body error %s", err)
		}

		_, err1 := k8client.RESTClient().Post().AbsPath("/apis/networkop.co.uk/v1beta1").Body(data).DoRaw(context.TODO())

		if err1 != nil {
			return fmt.Errorf("post restclient error %s", err1)
		}

		////////// Create CRD

		// clientConfig := ctrl.GetConfigOrDie()

		// apiextensionsclientSet, err := apiextensionsclientset.NewForConfig(clientConfig)

		// if err != nil {
		// 	return fmt.Errorf("crd creation error %s", err)
		// }

		// _, err1 := apiextensionsclientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Create(Ctx, crd_node, metav1.CreateOptions{})
		// if err1 == nil {
		// 	fmt.Println("topology CRD created")
		// } else {
		// 	fmt.Errorf("topology failed %s", err1)
		// }

		///////// create pods
		// newPod := &corev1.Pod{
		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Name: cm_node.Name,
		// 	},
		// 	Spec: corev1.PodSpec{
		// 		Containers: []corev1.Container{
		// 			{Name: "test1box", Image: "busybox:latest", Command: []string{"sleep", "100000"}},
		// 		},
		// 	},
		// }

		// pod, err2 := k8client.CoreV1().Pods("default").Create(Ctx, newPod, metav1.CreateOptions{})

		// if err2 != nil {
		// 	return fmt.Errorf("create pod error %s", err2)
		// }
		// fmt.Println(pod)

	}

	return nil

}

func Topo_delete(k8client *kubernetes.Clientset, topo database.TopologyData) error {

	nodes := topo.Vnodes

	for _, node := range nodes {

		err2 := k8client.CoreV1().Pods("default").Delete(Ctx, node.Name, metav1.DeleteOptions{})

		if err2 != nil {
			return fmt.Errorf("create pod error %s", err2)
		}
	}
	return nil
}

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

//
// 	configMap.Data = body

// 	if err := client.Update(ctx, configMap); err != nil {
// 		// handle error
// 	}

// 	if err := client.Delete(ctx, configMap); err != nil {
// 		// handle error
// 	}
// }
