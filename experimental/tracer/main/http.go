package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	dockerClient "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/probe/docker"
)

func respondWith(w http.ResponseWriter, code int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error handling http request: %v", err.Error())
	}
}

func (t *tracer) http(port int) {
	router := mux.NewRouter()

	router.Methods("GET").Path("/containers").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		containers := []*dockerClient.Container{}
		t.docker.WalkContainers(func(container docker.Container) {
			containers = append(containers, container.Container())
		})

		respondWith(w, http.StatusOK, containers)
	})

	router.Methods("GET").Path("/pid").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondWith(w, http.StatusOK, t.ptrace.AttachedPIDs())
	})

	router.Methods("POST").Path("/pid/{pid:\\d+}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pid, err := strconv.Atoi(mux.Vars(r)["pid"])
		if err != nil {
			respondWith(w, http.StatusBadRequest, err.Error())
			return
		}

		t.ptrace.TraceProcess(pid)
		w.WriteHeader(204)
	})

	router.Methods("DELETE").Path("/pid/{pid:\\d+}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pid, err := strconv.Atoi(mux.Vars(r)["pid"])
		if err != nil {
			respondWith(w, http.StatusBadRequest, err.Error())
			return
		}

		t.ptrace.StopTracing(pid)
		w.WriteHeader(204)
	})

	router.Methods("GET").Path("/trace").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondWith(w, http.StatusOK, t.store.Traces())
	})

	router.Methods("GET").PathPrefix("/").Handler(http.FileServer(FS(false))) // everything else is static

	log.Printf("Launching HTTP API on port %d", port)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Printf("Unable to create http listener: %v", err)
	}
}
