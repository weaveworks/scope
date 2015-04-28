package main

import (
	"encoding/json"
	"net/http"
)

func respondWith(w http.ResponseWriter, code int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}
