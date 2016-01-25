package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

func dot(w io.Writer, t render.RenderableNodes) {
	fmt.Fprintf(w, "digraph G {\n")
	fmt.Fprintf(w, "\toutputorder=edgesfirst;\n")
	fmt.Fprintf(w, "\toverlap=scale;\n")
	fmt.Fprintf(w, "\tnode [style=filled];\n")
	fmt.Fprintf(w, "\t\n")
	t.ForEach(func(rn render.RenderableNode) {
		label := rn.LabelMajor
		if len(label) > 20 {
			label = label[:20] + "..."
		}
		fmt.Fprintf(w, "\t%q [label=%q];\n", rn.ID, label)
		for _, other := range rn.Adjacency {
			fmt.Fprintf(w, "\t%q -> %q;\n", rn.ID, other)
		}
		fmt.Fprintf(w, "\t\n")
	})
	fmt.Fprintf(w, "}\n")
}

func handleHTML(rpt report.Report) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest("HTML", r)
		format := `<html><head></head><body><center><img src="/svg?%s" style="margin:5%;"/></center></body></html>`
		fmt.Fprintf(w, format, r.URL.RawQuery)
	}
}

func handleDot(rpt report.Report) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest("Dot", r)

		r.ParseForm()
		var (
			engine   = getDefault(r.Form, "engine", "twopi")
			topology = getDefault(r.Form, "topology", "containers")
		)
		log.Printf("engine=%s topology=%s", engine, topology)

		t, err := renderTo(rpt, topology)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("render %s to %d node(s)", topology, t.Size())

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		dot(w, t)
	}
}

func handleSVG(rpt report.Report) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest("SVG", r)

		r.ParseForm()
		var (
			engine   = getDefault(r.Form, "engine", "twopi")
			topology = getDefault(r.Form, "topology", "containers")
		)
		log.Printf("engine=%s topology=%s", engine, topology)

		t, err := renderTo(rpt, topology)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("render %s to %d node(s)", topology, t.Size())

		vizcmd, err := exec.LookPath(engine)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer
		dot(&buf, t)

		w.Header().Set("Content-Type", "image/svg+xml")
		if err := (&exec.Cmd{
			Path:   vizcmd,
			Args:   []string{vizcmd, "-Tsvg"},
			Stdin:  &buf,
			Stdout: w,
			Stderr: os.Stderr,
		}).Run(); err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func logRequest(what string, r *http.Request) {
	log.Printf("> %s %s %s %s", what, r.RemoteAddr, r.Method, r.URL.String())
}

func getDefault(v url.Values, key, def string) string {
	if val := v.Get(key); val != "" {
		return val
	}
	return def
}
