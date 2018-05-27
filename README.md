1. get a CFAR + CFCR running somewhere
   (see [deploy](deploy) folder for a local bosh-lite deployment)

2. deploy some sample services
   ```
   kubectl apply -f httpbin.yaml
   ```

3. ???
   ```
   go run main.go -kubeconfig ~/.kube/config
   ```

todo:
- call [`UpsertTcpRouteMappings` on CF Routing API](https://godoc.org/github.com/cloudfoundry/routing-api)
- what are the lifecycle events?  when does the `ExternalPort` get allocated?
- label selector for services to expose?
