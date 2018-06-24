// test.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	istioclient "github.com/devtio/canary/kubernetes"
	"github.com/gorilla/mux"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type User struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Age       int    `json:"age"`
}

type VirtualServiceDTO struct {
	Name           string `json:name`
	Release        string `json.release`
	NewVersion     string `json.newVersion`
	DefaultVersion string `json.defaultVersion`
}

// Command line arguments
var (
	argConfigFile = flag.String("config", "", "Path to the YAML configuration file. If not specified, environment variables will be used for configuration.")
)

func main() {

	r := mux.NewRouter()

	// GET all namespaces
	r.HandleFunc("/namespaces", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		fmt.Println("Getting all namespaces")
		if r.Method == "GET" {
			// GET namespaces
			namespaces, err2 := client.GetNamespaces()
			if err2 != nil {
				fmt.Errorf("Error occurred in getting namespaces", err2)
				return
			}
			fmt.Println("Namespaces retrieved: ", len(namespaces.Items))
			json.NewEncoder(w).Encode(namespaces)
		}

	})

	// GET virtual services in the default namespace
	r.HandleFunc("/virtual-services", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		fmt.Println("Getting virtual services in default namespace")
		if r.Method == "GET" {
			// GET virtual services
			virtualServices, err2 := client.GetVirtualServices("default", "")
			if err2 != nil {
				fmt.Errorf("Error occurred in getting virtual services", err2)
				return
			}
			fmt.Println("Virtual services retrieved: ", len(virtualServices))
			json.NewEncoder(w).Encode(virtualServices)
		}
	})

	// GET or POST virtual services by namespace
	r.HandleFunc("/virtual-services/{namespace}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		fmt.Println("Getting Virtual Services for namespace: ", namespace)
		if r.Method == "GET" {
			// GET virtual services
			virtualServices, err2 := client.GetVirtualServices(namespace, "")
			if err2 != nil {
				fmt.Errorf("Error occurred in getting virtual services", err2)
				return
			}
			fmt.Println("Virtual services retrieved: ", len(virtualServices))
			json.NewEncoder(w).Encode(virtualServices)
		} else if r.Method == "POST" {
			// POST new virtual service
			result := "Error"
			var vsd VirtualServiceDTO
			json.NewDecoder(r.Body).Decode(&vsd)
			fmt.Println("Attempting to create virtual service")

			virtualService := (&istioclient.VirtualService{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: vsd.Name,
				},
				Spec: map[string]interface{}{
					"hosts": []interface{}{
						"reviews",
					},
				},
			}).DeepCopyIstioObject()

			var rrErr error
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, rrErr := client.CreateVirtualService(namespace, virtualService); rrErr == nil {
					result = "Success"
					fmt.Errorf("Error: ", rrErr)
				}
			}()
			wg.Wait()
			fmt.Errorf("Error: ", rrErr)
			if rrErr != nil {
				fmt.Errorf("Error: ", rrErr)
				return
			}
			fmt.Fprintf(w, "Virtual service created - Name: %s, Release: %s, New Version: %s, Default: %s, In namespace: %s", vsd.Name, vsd.Release, vsd.NewVersion, vsd.DefaultVersion, namespace)
		}

	})

	// get pods by namespace
	r.HandleFunc("/pods/{namespace}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		fmt.Println("Getting Pods for namespace: ", namespace)
		if r.Method == "GET" {
			// GET virtual services
			pods, err2 := client.GetNamespacePods(namespace)
			if err2 != nil {
				fmt.Errorf("Error occurred in getting pods", err2)
				return
			}
			fmt.Println("Pods retrieved: ", len(pods.Items))
			json.NewEncoder(w).Encode(pods)
		}
	})

	// get pods by release label and namespace
	r.HandleFunc("/pods/{namespace}/{release}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		release := vars["release"]
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		fmt.Println("Getting Pods for namespace: ", namespace, ", and release: ", release)
		fmt.Println("method: ", r.Method)
		if r.Method == "GET" {
			// GET pods
			fmt.Println("calling client", client)
			pods, err2 := client.GetNamespacePodsByRelease(namespace, release)
			if err2 != nil {
				fmt.Errorf("Error occurred in getting pods", err2)
				return
			}
			fmt.Println("Pods retrieved: ", len(pods.Items))
			json.NewEncoder(w).Encode(pods)
		}
	})

	// Health endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == "GET" {
			results := make(map[string]string)
			results["status"] = "OK"

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)

			json.NewEncoder(w).Encode(results)
		}
	})

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8080",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
