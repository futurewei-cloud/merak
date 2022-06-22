package utils

import (
	// "flag"
	"fmt"
	// "path/filepath"

	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	// "k8s.io/client-go/tools/clientcmd"
)

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

	config := ctrl.GetConfigOrDie()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return clientset, err

}
