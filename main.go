// inspired by https://github.com/kubernetes/sample-controller
package main

import (
	"flag"
	"fmt"
	"log"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	routing_api_models "code.cloudfoundry.org/routing-api/models"
	uaa_client "code.cloudfoundry.org/uaa-go-client"
	uaa_config "code.cloudfoundry.org/uaa-go-client/config"
	routing_api "github.com/cloudfoundry/routing-api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var (
		masterURL       string
		kubeconfig      string
		namespace       string
		uaaClientName   string
		uaaClientSecret string
		uaaURL          string
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&namespace, "namespace", "default", "The namespace to inspect for services.")
	flag.StringVar(&uaaClientName, "uaa-client-name", "routing_api_client", "UAA client name for connecting to the Routing API")
	flag.StringVar(&uaaClientSecret, "uaa-client-secret", "", "UAA client secret for connecting to the Routing API.  Try 'bosh int <(credhub get -n /bosh-lite/cf/uaa_clients_routing_api_client_secret) --path /value'")
	flag.StringVar(&uaaURL, "uaa-url", "https://uaa.bosh-lite.com", "UAA API URL")
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("reading kubeconfig: %s", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("new kubernetes client: %s", err.Error())
	}

	uaaConfig := &uaa_config.Config{
		ClientName:       uaaClientName,
		ClientSecret:     uaaClientSecret,
		UaaEndpoint:      uaaURL,
		SkipVerification: true, // fixme!
	}

	logger := lager.NewLogger("some-logger")
	clock := clock.NewClock()
	uaaClient, err := uaa_client.NewClient(logger, uaaConfig, clock)
	if err != nil {
		log.Fatalf("uaa client: %s", err)
	}

	fmt.Printf("Connecting to: %s ...\n", uaaConfig.UaaEndpoint)

	token, err := uaaClient.FetchToken(true)
	if err != nil {
		log.Fatalf("uaa fetch token: %s", err)
	}

	fmt.Printf("Token: %#v\n", token)

	routingApiClient := routing_api.NewClient("https://api.bosh-lite.com", true)
	routingApiClient.SetToken(token.AccessToken)
	err = routingApiClient.CreateRouterGroup(routing_api_models.RouterGroup{
		Name:            "cfcr-tcp",
		Type:            routing_api_models.RouterGroup_TCP,
		ReservablePorts: routing_api_models.ReservablePorts("2000-3000"),
	})
	if err != nil {
		log.Fatalf("creating router group: %s", err)
	}
	myRouterGroup, err := routingApiClient.RouterGroupWithName("cfcr-tcp")
	if err != nil {
		log.Fatalf("getting that router group: %s", err)
	}

	serviceNodePorts, err := getServicesWithNodePorts(namespace, kubeClient.CoreV1())
	if err != nil {
		log.Fatalf("getting services: %s", err)
	}

	nodeIPs, err := getNodesWithIPs(kubeClient.CoreV1())
	if err != nil {
		log.Fatalf("getting nodes: %s", err)
	}

	ttl := 60

	var tcpRouteMappings []routing_api_models.TcpRouteMapping
	for _, nodeIP := range nodeIPs {
		for serviceName, nodePort := range serviceNodePorts {
			fmt.Printf("%s <- %s:%d\n", serviceName, nodeIP, nodePort)
			tcpRouteMappings = append(tcpRouteMappings, routing_api_models.TcpRouteMapping{
				TcpMappingEntity: routing_api_models.TcpMappingEntity{
					RouterGroupGuid: myRouterGroup.Guid,
					HostPort:        uint16(nodePort),
					HostIP:          nodeIP,
					ExternalPort:    2001,
					TTL:             &ttl,
				},
			})
		}
	}

	err = routingApiClient.UpsertTcpRouteMappings(tcpRouteMappings)
	if err != nil {
		log.Fatalf("upserting route mappings: %s", err)
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
