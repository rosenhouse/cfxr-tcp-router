// inspired by https://github.com/kubernetes/sample-controller
package main

import (
	"flag"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var (
		masterURL  string
		kubeconfig string
		namespace  string
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&namespace, "namespace", "default", "The namespace to inspect for services.")
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("reading kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("new kubernetes client: %s", err.Error())
	}

	serviceNodePorts, err := getServicesWithNodePorts(namespace, kubeClient.CoreV1())
	if err != nil {
		log.Fatalf("getting services: %s", err)
	}

	nodeIPs, err := getNodesWithIPs(kubeClient.CoreV1())
	if err != nil {
		log.Fatalf("getting nodes: %s", err)
	}
	for _, nodeIP := range nodeIPs {
		for serviceName, nodePort := range serviceNodePorts {
			fmt.Printf("%s <- %s:%d\n", serviceName, nodeIP, nodePort)
		}
	}
}

func getServicesWithNodePorts(namespace string, coreV1Client corev1.CoreV1Interface) (map[string]int, error) {
	allServices, err := coreV1Client.Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	serviceNodePorts := make(map[string]int)
	for _, service := range allServices.Items {
		nodePort := service.Spec.Ports[0].NodePort
		if nodePort == 0 {
			continue
		}
		serviceNodePorts[service.Name] = int(nodePort)
	}
	return serviceNodePorts, nil
}

func getNodesWithIPs(coreV1Client corev1.CoreV1Interface) ([]string, error) {
	allNodes, err := coreV1Client.Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	allNodeIPs := []string{}
	for _, node := range allNodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Type != "InternalIP" {
				continue
			}
			allNodeIPs = append(allNodeIPs, address.Address)
		}
	}
	return allNodeIPs, nil
}
