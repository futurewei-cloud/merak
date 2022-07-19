package utils

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
)

const K8SAPIURL string = "https://10.213.43.224:6443"
const K8SCONFIGPATH string = "/etc/kubernetes/admin.conf"

func K8sClient() (*kubernetes.Clientset, error) {

	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, "etc/kubernetes", "admin.conf"), "")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()

	// kubeconfig := flag.String("kubeconfig", filepath.Join("/root", "etc/kubernetes", "admin.conf"), "")
	// flag.Parse()

	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	return nil, fmt.Errorf(err.Error())
	// }

	// config := ctrl.GetConfigOrDie()

	config, err_config := K8sConfig()
	if err_config != nil {
		return nil, fmt.Errorf(err_config.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return clientset, err

}

func K8sConfig() (*rest.Config, error) {

	config, err := clientcmd.BuildConfigFromFlags(K8SAPIURL, K8SCONFIGPATH)
	if err != nil {
		return nil, err
	}

	return config, nil

}
