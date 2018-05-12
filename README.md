```
minikube start
kubectl apply -f httpbin.yaml
go run main.go -kubeconfig ~/.kube/config
```
