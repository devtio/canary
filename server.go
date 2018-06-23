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

	r.HandleFunc("/namespaces", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		fmt.Println("New client instantiated")
		if r.Method == "GET" {
			// GET namespaces
			namespaces, err2 := client.GetNamespaces()
			if err2 != nil {
				fmt.Errorf("Error occurred in getting namespaces", err)
				return
			}
			fmt.Println("Namespaces retrieved: ", len(namespaces.Items))
			json.NewEncoder(w).Encode(namespaces)
		}

	})

	r.HandleFunc("/virtual-services", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		fmt.Println("New client instantiated")
		if r.Method == "GET" {
			// GET virtual services
			virtualServices, err2 := client.GetVirtualServices("default", "")
			if err2 != nil {
				fmt.Errorf("Error occurred in getting virtual services", err)
				return
			}
			fmt.Println("Virtual services retrieved: ", len(virtualServices))
			json.NewEncoder(w).Encode(virtualServices)
		}
	})

	r.HandleFunc("/virtual-services/{namespace}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		fmt.Println("New client instantiated")
		fmt.Println("Getting Virtual Services for namespace: ", namespace)
		if r.Method == "GET" {
			// GET virtual services
			virtualServices, err2 := client.GetVirtualServices(namespace, "")
			if err2 != nil {
				fmt.Errorf("Error occurred in getting virtual services", err)
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
