package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	istioclient "github.com/devtio/canary/kubernetes"
	models "github.com/devtio/canary/models"
	"github.com/gorilla/mux"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListReleases(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	client, err := istioclient.NewClient()

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

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

		// get the object spec
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

func CreateRelease(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	client, err := istioclient.NewClient()

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Println("Decoding request body into Release Object")
	var release models.Release
	json.NewDecoder(r.Body).Decode(&release)
	fmt.Println("Decoded release: ", release)

	// We need to modify the existing gateway virtualservice
	// So for every http route in the existing virtualservice, if the destination host service has a version defined in our release object, then we change the subset to equal that version?
	// and append the devtio release header

	// first get virtual services
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

		// get the object spec
		spec := vs.GetSpec()
		hosts := spec["hosts"].([]interface{})
		_, gatewaysPresent := spec["gateways"].([]interface{})
		var hostsStringArray []string
		for _, host := range hosts {
			hostsStringArray = append(hostsStringArray, host.(string))
		}
		https, httpsPresent := spec["http"].([]interface{})
		// first get the release label
		if httpsPresent {
			appsToAddToGatewayVirtualService := map[string]string{}
			appsToAddToHostVirtualService := map[string]string{}
			for _, http := range https {
				httpMap, isHTTPMap := http.(map[string]interface{})
				if isHTTPMap {
					// read route section
					app := ""
					routes, routesPresent := httpMap["route"].([]interface{})
					if routesPresent {
						for _, route := range routes {
							routeMap, isRouteMap := route.(map[string]interface{})
							if isRouteMap {
								destination, destinationPresent := routeMap["destination"].(map[string]interface{})
								if destinationPresent {
									app, _ = destination["host"].(string)
								}
							}
						}
					}
					for _, appInRelease := range release.Apps {
						appInReleaseID, appLabelPresent := appInRelease.Labels["app"]
						versionInRelease := appInRelease.Labels["version"]
						if appLabelPresent && appInReleaseID == app {
							// gateway-bound vs has top level gateways in spec
							if gatewaysPresent && sameStringSlice(hostsStringArray, release.Gateway.Hosts) {
								fmt.Println("Need to add http rule for app: ", app, " for gateway-bound virtualservice")
								appsToAddToGatewayVirtualService[app] = versionInRelease
							}

							// other virtual services - create a match rule for each virtualservice for host that is mentioned in release
							for _, appInHosts := range hostsStringArray {
								if appInHosts == appInReleaseID {
									fmt.Println("Need to add match devtio rule for app: ", appInReleaseID, " for virtualservice bound to ", appInReleaseID)
									appsToAddToHostVirtualService[appInReleaseID] = versionInRelease
								}
							}
						}
					}
				}
			}
			// Now modify the virtual services accordingly
			// Gateway-bound VS
			for a, v := range appsToAddToGatewayVirtualService {
				newHTTPRule := map[string]interface{}{
					"route": []map[string]interface{}{
						{
							"destination": map[string]string{
								"host":   a,
								"subset": v,
							},
						},
					},
				}
				https = append(https, newHTTPRule)
			}
			// Host-bound VS
			for a, v := range appsToAddToHostVirtualService {
				newHTTPRule := map[string]interface{}{
					"match": []map[string]interface{}{
						{
							"headers": map[string]interface{}{
								"devtio": map[string]string{
									"exact": release.ID,
								},
							},
						},
					},
					"route": []map[string]interface{}{
						{
							"destination": map[string]string{
								"host":   a,
								"subset": v,
							},
						},
					},
				}
				https = append(https, newHTTPRule)
			}

		}
		spec["http"] = https
		res, err := putVirtualService(*client, meta, spec, namespace)
		if err != nil {
			fmt.Println(w, "Error: ", err.Error())
		} else {
			fmt.Println(w, res)
		}
	}
}

func putVirtualService(client istioclient.IstioClient, meta meta_v1.ObjectMeta, spec map[string]interface{}, namespace string) (string, error) {
	result := "Error"
	virtualService := (&istioclient.VirtualService{
		ObjectMeta: meta,
		Spec:       spec,
	}).DeepCopyIstioObject()
	var rrErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, rrErr := client.PutVirtualService(namespace, virtualService); rrErr == nil {
			result = "Success"
			fmt.Println("Error: ", rrErr)
		}
	}()
	wg.Wait()
	if rrErr != nil {
		fmt.Println("Error: ", rrErr)
	} else {
		fmt.Println("Virtual service modified successfully: ", virtualService)
	}
	return result, rrErr
}
