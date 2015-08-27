package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

	"github.com/weaveworks/scope/report"
)

func handleJSON(tpy report.Topology) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(getView(r).eval(tpy.Copy()).NodeMetadatas); err != nil {
			log.Print(err)
			return
		}
	}
}

func handleDOT(tpy report.Topology) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		dot(w, getView(r).eval(tpy.Copy()))
	}
}

func handleSVG(tpy report.Topology) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command(engine(r), "-Tsvg")

		buff := &bytes.Buffer{}
		dot(buff, getView(r).eval(tpy.Copy()))
		cmd.Stdin = buff

		cmd.Stdout = w
		stderr, err := cmd.StderrPipe()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		if err := cmd.Run(); err != nil {
			stderrB, _ := ioutil.ReadAll(stderr)
			log.Printf("stderr: %s, %q", err, stderrB)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><head>\n")
	//fmt.Fprintf(w, `<meta http-equiv="refresh" content="10">`+"\n")
	fmt.Fprintf(w, "</head><body>\n")
	fmt.Fprintf(w, `<center><img src="/svg?%s" width="100%%" height="95%%"></center>`+"\n", r.URL.Query().Encode())
	fmt.Fprintf(w, "</body></html>\n")
}

func getView(r *http.Request) view {
	var expressions []expression
	expressions = append(expressions, getExpressionsFromBody(r)...)
	expressions = append(expressions, getExpressionsFromQuery(r)...)
	log.Printf("Serving view with %d expression(s)", len(expressions))
	return view(expressions)
}

func getExpressionsFromQuery(r *http.Request) []expression {
	strs := []string{}
	if err := r.ParseForm(); err != nil {
		log.Printf("Get expressions from query: %v", err)
		return []expression{}
	}
	for _, str := range r.Form["expr"] {
		log.Printf("Query expression: %s", str)
		strs = append(strs, str)
	}
	return parseExpressions(strs)
}

func getExpressionsFromBody(r *http.Request) []expression {
	strs := []string{}
	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		log.Printf("Body expression: %s", scanner.Text())
		strs = append(strs, scanner.Text())
	}
	return parseExpressions(strs)
}

func engine(r *http.Request) string {
	engine := r.FormValue("engine")
	if engine == "" {
		engine = "dot"
	}
	return engine
}
