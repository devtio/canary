package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	istioclient "github.com/devtio/canary/kubernetes"
	models "github.com/devtio/canary/models"
	"github.com/gorilla/mux"
)

func ListTrafficSegments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	w.Header().Set("Access-Control-Allow-Origin", "*")
	client, err := istioclient.NewClient()

	if err != nil {
		fmt.Errorf("Error occurred", err)
		return
	}

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

func CreateTrafficSegment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	// releaseId, releaseIdPresent := vars["releaseId"]
	client, err := istioclient.NewClient()

	// POST new virtual service
	var vsd models.VirtualServiceDTO
	json.NewDecoder(r.Body).Decode(&vsd)
	fmt.Println("Attempting to create traffic segment service")

	res, err := createVirtualService(*client, vsd, namespace)
	if err != nil {
		fmt.Fprintf(w, "Error: ", err.Error())
	} else {
		fmt.Fprintf(w, res)
	}
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
