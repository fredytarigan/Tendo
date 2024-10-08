package k8s

import (
	"fmt"

	"github.com/fredytarigan/Tendo/pkg/tendo/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func BuildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func GetKubernetesConfig(kubeconfig string) kubernetes.Clientset {
	config, err := BuildConfig(kubeconfig)
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("unable to build kubernetes client config with error: %s", err))
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("unable to setup kubernetes client with error: %s", err))
	}

	return *clientSet
}