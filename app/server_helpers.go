package app

import (
	"net/http"

	"github.com/ugorji/go/codec"

	log "github.com/sirupsen/logrus"
)

func respondWith(w http.ResponseWriter, code int, response interface{}) {
	if err, ok := response.(error); ok {
		log.Errorf("Error %d: %v", code, err)
		response = err.Error()
	} else if 500 <= code && code < 600 {
		log.Errorf("Non-error %d: %v", code, response)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(code)
	encoder := codec.NewEncoder(w, &codec.JsonHandle{})
	if err := encoder.Encode(response); err != nil {
		log.Errorf("Error encoding response: %v", err)
	}
}
