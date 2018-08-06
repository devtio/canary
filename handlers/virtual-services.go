package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	istioclient "github.com/devtio/canary/kubernetes"
	"github.com/devtio/canary/log"
	models "github.com/devtio/canary/models"
	"github.com/gorilla/mux"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListVirtualServices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	client, err := istioclient.NewClient()

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	virtualServices, err2 := client.GetVirtualServices(namespace, "")
	if err2 != nil {
		log.Error(err)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Info("Virtualservices retrieved: %s", len(virtualServices))
	RespondWithJSON(w, http.StatusOK, virtualServices)

}

func CreateVirtualService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	client, err := istioclient.NewClient()

	// POST new virtual service
	var vsd models.VirtualServiceDTO
	json.NewDecoder(r.Body).Decode(&vsd)
	fmt.Println("Attempting to create virtual service")

	res, err := createVirtualService(*client, vsd, namespace)
	if err != nil {
		fmt.Fprintf(w, "Error: ", err.Error())
	} else {
		fmt.Fprintf(w, res)
	}
}

func createVirtualService(client istioclient.IstioClient, vsd models.VirtualServiceDTO, namespace string) (string, error) {
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
			fmt.Println("Error: ", rrErr)
		}
	}()
	wg.Wait()
	fmt.Println("Error: ", rrErr)
	if rrErr != nil {
		fmt.Println("Error: ", rrErr)
	}
	fmt.Println("Virtual service created: ", vsd.Name, vsd.Host, vsd.ReleaseID, vsd.ReleaseName, namespace)
	return result, rrErr
}
