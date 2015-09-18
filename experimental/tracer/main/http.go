package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
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

func (t *tracer) pidsForContainer(id string) ([]int, error) {
	var container docker.Container
	t.docker.WalkContainers(func(c docker.Container) {
		if c.ID() == id {
			container = c
		}
	})

	if container == nil {
		return []int{}, fmt.Errorf("Not Found")
	}

	pidTree, err := process.NewTree(process.NewWalker("/proc"))
	if err != nil {
		return []int{}, err
	}

	return pidTree.GetChildren(container.PID())
}

type Container struct {
	Id   string
	Name string
	PIDs []int
}

func (t *tracer) http(port int) {
	router := mux.NewRouter()

	router.Methods("GET").Path("/container").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pidTree, err := process.NewTree(process.NewWalker("/proc"))
		if err != nil {
			respondWith(w, http.StatusBadRequest, err.Error())
			return
		}

		containers := []Container{}
		t.docker.WalkContainers(func(container docker.Container) {
			children, _ := pidTree.GetChildren(container.PID())
			out := Container{
				Name: strings.TrimPrefix(container.Container().Name, "/"),
				Id: container.ID(),
				PIDs: children,
			}
			containers = append(containers, out)
		})

		respondWith(w, http.StatusOK, containers)
	})

	router.Methods("POST").Path("/container/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		children, err := t.pidsForContainer(id)
		if err != nil {
			respondWith(w, http.StatusBadRequest, err.Error())
			return
		}

		for _, pid := range children {
			t.ptrace.TraceProcess(pid)
		}
		w.WriteHeader(204)
	})

	router.Methods("DELETE").Path("/container/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		children, err := t.pidsForContainer(id)
		if err != nil {
			respondWith(w, http.StatusBadRequest, err.Error())
			return
		}

		for _, pid := range children {
			t.ptrace.StopTracing(pid)
		}
		w.WriteHeader(204)
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

	router.Methods("GET").Path("/traces").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondWith(w, http.StatusOK, t.store.Traces())
	})

	mime.AddExtensionType(".svg", "image/svg+xml")

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
