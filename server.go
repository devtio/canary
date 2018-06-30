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
	"github.com/devtio/canary/services/models"
	"github.com/gorilla/mux"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type User struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Age       int    `json:"age"`
}

type VirtualServiceDTO struct {
	Name        string `json:name`
	Host        string `json.host`
	Subset      string `json.subset`
	ReleaseID   string `json.releaseId`
	ReleaseName string `json.releaseName`
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
	// GET request parameter is namespace
	// For POST request, data should be in following format:
	// {
	// 	name: "a",
	// 	host: "a",
	// 	header: "x-client-id",
	// 	release: "my-fancy-release"
	// }
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
			var vsd VirtualServiceDTO
			json.NewDecoder(r.Body).Decode(&vsd)
			fmt.Println("Attempting to create virtual service")

			res, err := createVirtualService(*client, vsd, namespace)
			if err != nil {
				fmt.Fprintf(w, "Error: ", err.Error())
			} else {
				fmt.Fprintf(w, res)
			}
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

	// endpoint to GET and POST releases
	r.HandleFunc("/releases/{namespace}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		if r.Method == "POST" {
			// create a Virtual Service from the release
			fmt.Println("Decoding request body into Release Object")
			var release models.Release
			json.NewDecoder(r.Body).Decode(&release)
			fmt.Println("Decoded release: ", release)

			for _, label := range release.PodLabels {
				vsd := VirtualServiceDTO{
					Name:        release.Name + "-" + label["app"],
					ReleaseName: release.Name,
					ReleaseID:   release.ID,
					Host:        label["app"],
					Subset:      label["version"],
				}
				fmt.Println("Attempting to create virtual service ", vsd, namespace)
				res, err := createVirtualService(*client, vsd, namespace)
				if err != nil {
					fmt.Fprintf(w, "Error: ", err.Error())
				} else {
					fmt.Fprintf(w, res)
				}
			}

		} else if r.Method == "GET" {
			// first get virtual services
			fmt.Println("Calling GET virtual services")
			virtualServices, err2 := client.GetVirtualServices(namespace, "")
			if err2 != nil {
				fmt.Errorf("Error occurred in getting virtual services", err2)
				return
			}
			fmt.Println("Virtual services retrieved: ", len(virtualServices))
			releases := []models.Release{}
			// iterate over virtual services
			for _, vs := range virtualServices {
				// get metadata
				meta := vs.GetObjectMeta()
				labels := meta.GetLabels()
				releaseID, releaseIDPresent := labels["releaseId"]
				releaseName, releaseNamePresent := labels["releaseName"]

				// get the obeject spec
				spec := vs.GetSpec()
				// very long nested structure to get the required fields from the virtual service spec
				hosts := spec["hosts"].([]interface{})
				https, httpsPresent := spec["http"].([]interface{})
				// first get the release label
				if httpsPresent {
					for _, http := range https {
						httpMap, isHTTPMap := http.(map[string]interface{})
						if isHTTPMap {
							matches, matchesPresent := httpMap["match"].([]interface{})
							if matchesPresent {
								for _, match := range matches {
									matchMap, isMatchMap := match.(map[string]interface{})
									if isMatchMap {
										sourceLabels, sourceLabelsPresent := matchMap["sourceLabels"].(map[string]interface{})
										if sourceLabelsPresent {
											release, ok := sourceLabels["release"].(string)
											if ok {
												fmt.Println("Release label is: ", release)

												// now get host and subset
												routes, routesPresent := httpMap["route"].([]interface{})
												if routesPresent {
													for _, route := range routes {
														routeMap, isRouteMap := route.(map[string]interface{})
														if isRouteMap {
															destination, destinationPresent := routeMap["destination"].(map[string]interface{})
															if destinationPresent {
																host, ok := destination["host"].(string)
																if ok {
																	fmt.Println("host label is: ", host)
																}
																subset, ok2 := destination["subset"].(string)
																if ok2 {
																	fmt.Println("subset label is: ", subset)
																}

																// create a Release model if host matches host
																for _, _host := range hosts {
																	if _host == host {
																		// matches, go check release id and name
																		if releaseIDPresent && releaseNamePresent {
																			labels := []models.Labels{map[string]string{
																				"app":     host,
																				"version": subset,
																			}}
																			existing := false
																			// add label to existing release if present
																			for i := range releases {
																				if releases[i].ID == releaseID {
																					releases[i].PodLabels = append(releases[i].PodLabels, labels[0])
																					existing = true
																				}
																			}
																			// else add to a new release
																			if !existing {
																				release := models.Release{
																					ID:        releaseID,
																					Name:      releaseName,
																					PodLabels: labels,
																				}
																				releases = append(releases, release)
																			}
																		}
																	}
																}
															}
														}
													}
												}

											}
										}
									}
								}
							}
						}
					}
				}

			}
			json.NewEncoder(w).Encode(releases)
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

func createVirtualService(client istioclient.IstioClient, vsd VirtualServiceDTO, namespace string) (string, error) {
	result := "Error"
	virtualService := (&istioclient.VirtualService{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: vsd.Name,
			Labels: map[string]string{
				"releaseId":   vsd.ReleaseID,
				"releaseName": vsd.ReleaseName,
			},
		},
		Spec: map[string]interface{}{
			"hosts": []interface{}{
				vsd.Host,
			},
			"http": []map[string]interface{}{
				{
					"match": []map[string]interface{}{
						{
							"sourceLabels": map[string]string{
								"release": vsd.ReleaseID,
							},
						},
					},
					"route": []map[string]interface{}{
						{
							"destination": map[string]string{
								"host":   vsd.Host,
								"subset": vsd.Subset,
							},
						},
					},
				},
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
	}
	fmt.Println("Virtual service created: ", vsd.Name, vsd.Host, vsd.ReleaseID, vsd.ReleaseName, namespace)
	return result, rrErr
}
