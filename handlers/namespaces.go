package handlers

import (
	"net/http"

	istioclient "github.com/devtio/canary/kubernetes"
	"github.com/devtio/canary/log"
)

// ServiceList is the API handler to fetch the list of services in a given namespace
func ListNamespaces(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	client, err := istioclient.NewClient()

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Info("Getting all namespaces")
	namespaces, err2 := client.GetNamespaces()

	if err2 != nil {
		log.Error(err)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Info("Namespaces retrieved: %s", len(namespaces.Items))
	RespondWithJSON(w, http.StatusOK, namespaces)
}
