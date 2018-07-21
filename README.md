# Devtio Canary #

Devtio 'Canary' allows you to setup and run Canary releases of your applications on Kubernetes.
You can target any change to specific users or groups, gradually roll out the change and monitor it's progress. If there are issues, you can rollback immediately.

## Test repo ##

This project starts a local web service written in golang.  
It is based off a JSON example web service https://gowebexamples.com/json/

### Installation ###

- git clone this repo into $GO_PATH/github.com/devtio
- `cd $GO_PATH/github.com/devtio/canary`
- `glide install --strip-vendor`

### Testing locally ###
- Ensure istio (>= v0.8) is running on the cluster
- Give permissions `kubectl create clusterrolebinding cluster-system-anonymous --clusterrole=cluster-admin --user=system:anonymous`
- Run the [devtio/dummy project](https://github.com/devtio/dummy) to create the dummy namespace with example resources i.e. `kubectl apply -f samples/dummy/setup.yaml`
- Run canary locally `go run server.go`

#### GET Releases Test
- `curl -s http://localhost:8080/releases/dummy` to test the GET releases method
- The output should look something like this: `{"release1":{"id":"release1","name":"release1","gateway":{"hosts":["dummy.xx.xxx.xxx.xxx.nip.io"]},"apps":[{"hosts":["a"],"labels":{"app":"a","version":"v2"}},{"hosts":["b"],"labels":{"app":"b","version":"v2"}}]}}`

#### GET TrafficSegments Test
- `curl -s http://localhost:8080/traffic-segments/dummy` to test the GET releases method
- The output should look something like this `[{"id":"release1","name":"release1","match":{"headers":{"x-client-id":{"exact":"fancy"}}}}]`

### Build and deploy image to minikube ###
- Ensure istio-system namespace is running on cluster
- `make build` from server folder
- `make minikube-docker` 
- `make k8s-deploy`

### Build and deploy image to GCP ###
- Ensure istio-system namespace is running on cluster
- `export GCR_ID=gcr.io/project_id/image` e.g. gcr.io/innate-lacing-206112/canary
- `export GATEWAY_URL=xxxxx` eg. export GATEWAY_URL=35.189.51.63
- `make build` from server folder
- `make gcr-docker` 
- `kubectl create ns canary`
- `make k8s-deploy`
- `curl canary.$GATEWAY_URL.nip.io/health` test via health endpoint
- Getting other endpoints to work is WIP.. working on how to get the in-cluster client working


### Testing on cluster ###
- tbd 
