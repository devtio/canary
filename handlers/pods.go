package handlers

import (
	"net/http"

	istioclient "github.com/devtio/canary/kubernetes"
	"github.com/devtio/canary/log"
	"github.com/gorilla/mux"
)

func ListPods(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	release := vars["release"]

	w.Header().Set("Access-Control-Allow-Origin", "*")
	client, err := istioclient.NewClient()

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if release != "" {
		pods, err2 := client.GetNamespacePodsByRelease(namespace, release)
		if err2 != nil {
			log.Error(err)
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		log.Info("Pods retrieved: %s", len(pods.Items))
		RespondWithJSON(w, http.StatusOK, pods)

	} else {
		pods, err2 := client.GetPods(namespace, "")
		if err2 != nil {
			log.Error(err)
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		log.Info("Pods retrieved: %s", len(pods.Items))
		RespondWithJSON(w, http.StatusOK, pods)

	}

}
