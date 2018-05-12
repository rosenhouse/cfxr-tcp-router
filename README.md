```
minikube start
kubectl apply -f httpbin.yaml
go run main.go -kubeconfig ~/.kube/config
```

todo:
- call [`UpsertTcpRouteMappings` on CF Routing API](https://godoc.org/github.com/cloudfoundry/routing-api)
- what are the lifecycle events?  when does the `ExternalPort` get allocated?
- label selector for services to expose?
