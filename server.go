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

	// GET istio gateways by namespace
	r.HandleFunc("/gateways/{namespace}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Println("Error occurred", err)
			return
		}

		if r.Method == "GET" {
			fmt.Println("Getting Gateways for namespace: ", namespace)
			gateways, err2 := client.GetGateways(namespace)
			if err2 != nil {
				fmt.Println("Error occurred in getting gateways", err2)
				return
			}
			fmt.Println("Gateways retrieved: ", len(gateways))
			json.NewEncoder(w).Encode(gateways)
		}
	})

	// TODO: GET and POST traffic segments to map traffic segment objects with Gateway + VS combo

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
			// var release models.Release
			// json.NewDecoder(r.Body).Decode(&release)
			// fmt.Println("Decoded release: ", release)

			// for _, label := range release.PodLabels {
			// 	vsd := VirtualServiceDTO{
			// 		Name:        release.Name + "-" + label["app"],
			// 		ReleaseName: release.Name,
			// 		ReleaseID:   release.ID,
			// 		Host:        label["app"],
			// 		Subset:      label["version"],
			// 	}
			// 	fmt.Println("Attempting to create virtual service ", vsd, namespace)
			// 	res, err := createVirtualService(*client, vsd, namespace)
			// 	if err != nil {
			// 		fmt.Fprintf(w, "Error: ", err.Error())
			// 	} else {
			// 		fmt.Fprintf(w, res)
			// 	}
			// }

		} else if r.Method == "GET" {
			// first get virtual services
			fmt.Println("Calling GET virtual services for namespace ", namespace)
			virtualServices, err2 := client.GetVirtualServices(namespace, "")
			if err2 != nil {
				fmt.Println("Error occurred in getting virtual services", err2)
				return
			}
			fmt.Println("Virtual services retrieved: ", len(virtualServices))
			releases := map[string]models.Release{}
			// iterate over virtual services
			for _, vs := range virtualServices {
				// get metadata
				meta := vs.GetObjectMeta()
				labels := meta.GetLabels()

				isManagedByDevtio, isManagedByDevtioPresent := labels["io.devtio.canary/managed"]

				if !isManagedByDevtioPresent || isManagedByDevtio == "false" {
					fmt.Println("VirtualService is not managed by devtio, skipping...")
					continue
				}

				// get the obeject spec
				spec := vs.GetSpec()
				// very long nested structure to get the required fields from the virtual service spec
				hosts := spec["hosts"].([]interface{})
				var hostsStringArray []string
				for _, host := range hosts {
					hostsStringArray = append(hostsStringArray, host.(string))
				}
				https, httpsPresent := spec["http"].([]interface{})
				// first get the release label
				if httpsPresent {
					for _, http := range https {
						httpMap, isHTTPMap := http.(map[string]interface{})
						if isHTTPMap {
							// get top-level hosts
							appendHeaders, appendHeadersPresent := httpMap["appendHeaders"].(map[string]interface{})
							if appendHeadersPresent {
								devtioHeader, devtioHeaderPresent := appendHeaders["devtio"].(string)
								if devtioHeaderPresent {
									fmt.Println("Top level hosts of release: ", devtioHeader, " are: ", hosts)
									// add to release map to return
									release, releasePresent := releases[devtioHeader]
									fmt.Println(releasePresent)
									if releasePresent {
										release.Gateway.Hosts = hostsStringArray
									} else {
										release = models.Release{
											ID:   devtioHeader,
											Name: devtioHeader,
											Gateway: models.Gateway{
												Hosts: hostsStringArray,
											},
										}
									}
									releases[devtioHeader] = release
								}
							}

							// read route section
							app, version := "", ""
							routes, routesPresent := httpMap["route"].([]interface{})
							if routesPresent {
								for _, route := range routes {
									routeMap, isRouteMap := route.(map[string]interface{})
									if isRouteMap {
										destination, destinationPresent := routeMap["destination"].(map[string]interface{})
										if destinationPresent {
											app, _ = destination["host"].(string)
											version, _ = destination["subset"].(string)
										}
									}
								}
							}

							// read matches section
							matches, matchesPresent := httpMap["match"].([]interface{})
							if matchesPresent {
								for _, match := range matches {
									matchMap, isMatchMap := match.(map[string]interface{})
									if isMatchMap {
										headers, headersPresent := matchMap["headers"].(map[string]interface{})
										if headersPresent {
											devtio, devtioPresent := headers["devtio"].(map[string]interface{})
											if devtioPresent {
												headerExact, headerExactPresent := devtio["exact"].(string)
												if headerExactPresent {
													fmt.Println("Found release, id: ", headerExact)
													fmt.Println("Hosts: ", hosts)
													// get labels
													fmt.Println("app: ", app, ", version: ", version)
													// add to release map to return
													release, releasePresent := releases[headerExact]
													app := models.App{
														Hosts: hostsStringArray,
														Labels: models.Labels{
															"app":     app,
															"version": version,
														},
													}

													if releasePresent {
														release.Apps = append(release.Apps, app)
													} else {
														release = models.Release{
															ID:   headerExact,
															Name: headerExact,
															Apps: []models.App{app},
														}
													}
													releases[headerExact] = release
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

	// endpoint to GET and POST releases
	r.HandleFunc("/traffic-segments/{namespace}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		w.Header().Set("Access-Control-Allow-Origin", "*")
		client, err := istioclient.NewClient()

		if err != nil {
			fmt.Errorf("Error occurred", err)
			return
		}

		if r.Method == "POST" {
			// todo
		} else if r.Method == "GET" {
			trafficSegments := []models.TrafficSegment{}
			// first get gateways
			fmt.Println("Calling GET gateways for namespace ", namespace)
			gateways, err2 := client.GetGateways(namespace)
			if err2 != nil {
				fmt.Println("Error occurred in getting gateways", err2)
				return
			}
			fmt.Println("Gateways retrieved: ", len(gateways))
			gatewayHosts := map[string][]string{}
			// iterate over gateways
			for _, gateway := range gateways {
				// get metadata
				spec := gateway.GetSpec()
				meta := gateway.GetObjectMeta()
				labels := meta.GetLabels()
				name := meta.GetName()
				isManagedByDevtio, isManagedByDevtioPresent := labels["io.devtio.canary/managed"]
				if !isManagedByDevtioPresent || isManagedByDevtio == "false" {
					fmt.Println("Gateway is not managed by devtio, skipping...")
					continue
				}
				servers, serversPresent := spec["servers"].([]interface{})

				if serversPresent {
					for _, server := range servers {
						serverMap, isServerMap := server.(map[string]interface{})
						if isServerMap {
							hosts, hostsPresent := serverMap["hosts"].([]interface{})
							if hostsPresent {
								for _, host := range hosts {
									hostsArray, isHostsArray := gatewayHosts[name]
									if isHostsArray {
										hostsArray = append(hostsArray, host.(string))
									} else {
										hostsArray = []string{host.(string)}
									}
									gatewayHosts[name] = hostsArray
								}
							}
						}
					}
				}
			}
			fmt.Println(gatewayHosts)

			// get virtual services that match the host/gateway
			fmt.Println("Calling GET virtual services for namespace ", namespace)
			virtualServices, err2 := client.GetVirtualServices(namespace, "")
			if err2 != nil {
				fmt.Println("Error occurred in getting virtual services", err2)
				return
			}
			fmt.Println("Virtual services retrieved: ", len(virtualServices))
			// iterate over virtual services
			for _, vs := range virtualServices {
				// get metadata
				meta := vs.GetObjectMeta()
				labels := meta.GetLabels()
				isManagedByDevtio, isManagedByDevtioPresent := labels["io.devtio.canary/managed"]

				if !isManagedByDevtioPresent || isManagedByDevtio == "false" {
					fmt.Println("VirtualService is not managed by devtio, skipping...")
					continue
				}
				spec := vs.GetSpec()
				// get hosts
				hosts := spec["hosts"].([]interface{})
				var hostsStringArray []string
				for _, host := range hosts {
					hostsStringArray = append(hostsStringArray, host.(string))
				}
				// get the obeject spec
				// get gateways
				gateways, gatewaysPresent := spec["gateways"].([]interface{})
				if gatewaysPresent {
					for _, gateway := range gateways {
						hostsArrayFromGateway := gatewayHosts[gateway.(string)]
						// check hosts match
						if sameStringSlice(hostsArrayFromGateway, hostsStringArray) {
							https, httpsPresent := spec["http"].([]interface{})
							// first get the release label
							if httpsPresent {
								for _, http := range https {
									httpMap, isHTTPMap := http.(map[string]interface{})
									if isHTTPMap {
										// get top-level hosts
										appendHeaders, appendHeadersPresent := httpMap["appendHeaders"].(map[string]interface{})
										if appendHeadersPresent {
											devtioHeader, devtioHeaderPresent := appendHeaders["devtio"].(string)
											if devtioHeaderPresent {
												// read matches section
												matches, matchesPresent := httpMap["match"].([]interface{})
												if matchesPresent {
													for _, match := range matches {
														matchMap, isMatchMap := match.(map[string]interface{})
														if isMatchMap {
															headers, headersPresent := matchMap["headers"].(map[string]interface{})
															if headersPresent {
																headersMap := map[string]*models.StringMatch{}
																for headerKey, headerValue := range headers {
																	stringMatchMap, isStringMatchMap := headerValue.(map[string]interface{})
																	if isStringMatchMap {
																		stringMatch := models.StringMatch{}
																		stringMatchExact, isStringMatchExactPresent := stringMatchMap["exact"]
																		stringMatchRegex, isStringMatchRegexPresent := stringMatchMap["regex"]
																		stringMatchPrefix, isStringMatchPrefixPresent := stringMatchMap["prefix"]
																		if isStringMatchExactPresent {
																			stringMatch.Exact = stringMatchExact.(string)
																		}
																		if isStringMatchRegexPresent {
																			stringMatch.Regex = stringMatchRegex.(string)
																		}
																		if isStringMatchPrefixPresent {
																			stringMatch.Prefix = stringMatchPrefix.(string)
																		}
																		headersMap[headerKey] = &stringMatch
																		match := models.HttpMatch{
																			Headers: headersMap,
																		}
																		trafficSegment := models.TrafficSegment{
																			ID:    devtioHeader,
																			Name:  devtioHeader,
																			Match: &match,
																		}
																		trafficSegments = append(trafficSegments, trafficSegment)
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
						} else {
							fmt.Println("Sorry, hosts don't match between Gateway and VirtualService")
						}
					}
				}

			}

			json.NewEncoder(w).Encode(trafficSegments)
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
		Addr:    "0.0.0.0:8090",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func sameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	if len(diff) == 0 {
		return true
	}
	return false
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
