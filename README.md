## Test repo

This project starts a local web service written in golang.  
It is based off a JSON example web service https://gowebexamples.com/json/

### Installation

- git clone this repo into $GO_PATH/github.com/devtio
- `cd $GO_PATH/github.com/devtio/canary`
- `glide install --strip-vendor`

### Testing locally
- `go run server.go`
- Run istio and bookinfo tutorial for testing if required
- Carry out the create route rule example of the bookinfo tutorial if required
- `curl -s http://localhost:8080/virtual-services` to test the GET virtual services method
- Test creating a virtual service that would represent the Release State for Canary `curl -s -XPOST -d'{"name":"dummy-4","releaseId":"my-fancy-release", "releaseName":"MyFancyRelease", "host":"b", "subset": "v1"}' http://localhost:8080/virtual-services/dumbo` (this represents v1 of service b being used for my-fancy-release)
- Test retrieving current state Releases `curl -s http://localhost:8080/releases/dumbo`

### Build and deploy image
- Ensure istio-system namespace is running on cluster
- `make build` from server folder
- `make minikube-docker` 
- `make k8s-deploy`

### Testing on cluster
- tbd 