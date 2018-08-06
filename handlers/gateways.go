package handlers

import (
	"net/http"

	istioclient "github.com/devtio/canary/kubernetes"
	"github.com/devtio/canary/log"
	"github.com/gorilla/mux"
)

func ListGateways(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	client, err := istioclient.NewClient()

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	gateways, err2 := client.GetGateways(namespace)
	if err2 != nil {
		log.Error(err)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Info("Gateways retrieved: %s", len(gateways))
	RespondWithJSON(w, http.StatusOK, gateways)
}
