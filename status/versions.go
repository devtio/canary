package status

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/devtio/canary/config"
	"github.com/devtio/canary/kubernetes"
	"k8s.io/api/core/v1"
	kube "k8s.io/client-go/kubernetes"
)

type serviceProduct func() (*ProductInfo, error)

var (
	expVersion = regexp.MustCompile("[0-9]\\.[0-9]\\.[0-9]")
)

func getVersions() {
	products := []serviceProduct{
		istioVersion,
		prometheusVersion,
		kubernetesVersion,
	}
	for _, prod := range products {
		getVersionProduct(prod)
	}
}

func getVersionProduct(serviceProduct serviceProduct) {
	productInfo, err := serviceProduct()
	if err == nil {
		info.Products = append(info.Products, *productInfo)
	}
}

func istioVersion() (*ProductInfo, error) {
	product := ProductInfo{}
	istioConfig := config.Get().Products.Istio
	resp, err := http.Get(istioConfig.UrlServiceVersion)
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			product.Name = "Istio"
			version := expVersion.FindStringSubmatch(string(body))
			if product.Version = "Unknown"; version != nil {
				product.Version = version[0]
			}
			return &product, nil
		}
	}
	return nil, err
}

type p8sResponseVersion struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
}

func prometheusVersion() (*ProductInfo, error) {
	product := ProductInfo{}
	prometheusV := new(p8sResponseVersion)
	prometheusUrl := config.Get().Products.PrometheusServiceURL
	resp, err := http.Get(prometheusUrl + "/version")
	if err == nil {
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&prometheusV)
		if err == nil {
			product.Name = "Prometheus"
			product.Version = prometheusV.Version
			return &product, nil
		}
	}
	return nil, err
}

const (
	// These constants are tweaks to the k8s client I think once are set up they won't change so no need to put them on the config
	// Default QPS and Burst are quite low and those are not designed for a backend that should perform several
	// queries to build an inventory of entities from a k8s backend.
	// Other k8s clients have increased these values to a similar values.
	k8sQPS   = 100
	k8sBurst = 200
)

func kubernetesVersion() (*ProductInfo, error) {
	product := ProductInfo{}
	config, err := kubernetes.ConfigClient()
	if err == nil {
		config.QPS = k8sQPS
		config.Burst = k8sBurst
		k8s, err := kube.NewForConfig(config)
		if err == nil {
			serverVersion, err := k8s.Discovery().ServerVersion()
			if err == nil {
				product.Name = "Kubernetes"
				product.Version = serverVersion.GitVersion
				return &product, nil
			}
		}
	}
	return nil, err
}
func getService(namespace string, service string) (*v1.ServiceSpec, error) {
	client, err := kubernetes.NewClient()
	if err != nil {
		return nil, err
	}
	details, err := client.GetServiceDetails(namespace, service)
	if err != nil {
		return nil, err
	}
	return &details.Service.Spec, nil
}
