package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/alicebob/cello/report"
)

func handleTXT(r Reporter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		dot(w, r.Report().Process.RenderBy(mapFunc(req), classView(req)))
	}
}

func handleSVG(r Reporter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cmd := exec.Command(engine(req), "-Tsvg")

		wc, err := cmd.StdinPipe()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cmd.Stdout = w

		dot(wc, r.Report().Process.RenderBy(mapFunc(req), classView(req)))
		wc.Close()

		w.Header().Set("Content-Type", "image/svg+xml")
		if err := cmd.Run(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><head>\n")
	fmt.Fprintf(w, `<meta http-equiv="refresh" content="3">`+"\n")
	fmt.Fprintf(w, "</head><body>\n")
	fmt.Fprintf(w, `<center><img src="/svg?%s" width="100%%" height="95%%"></center>`+"\n", r.URL.Query().Encode())
	fmt.Fprintf(w, "</body></html>\n")
}

func dot(w io.Writer, m map[string]report.RenderableNode) {
	fmt.Fprintf(w, "digraph G {\n")
	fmt.Fprintf(w, "\tgraph [ overlap=false ];\n")
	fmt.Fprintf(w, "\tnode [ shape=circle, style=filled ];\n")
	fmt.Fprintf(w, "\toutputorder=edgesfirst;\n")
	fmt.Fprintf(w, "\n")

	for _, node := range m {
		fmt.Fprintf(w, "\t\"%s\" [label=\"%s\n%s\"];\n", node.ID, node.LabelMajor, node.LabelMinor)
	}
	fmt.Fprintf(w, "\n")

	// Add ranking information by default.
	// Non-dot engines don't seem to be harmed by it.
	same := map[string][]string{}
	for _, node := range m {
		k, v := node.LabelMajor, fmt.Sprintf(`"%s"`, node.ID)
		same[k] = append(same[k], v)
	}
	for _, ids := range same {
		fmt.Fprintf(w, "\t{ rank=same; %s }\n", strings.Join(ids, " "))
	}
	fmt.Fprintf(w, "\n")

	for _, src := range m {
		for _, dstID := range src.Adjacency {
			fmt.Fprintf(w, "\t\"%s\" -> \"%s\";\n", src.ID, dstID)
		}
	}
	fmt.Fprintf(w, "}\n")
}

func engine(r *http.Request) string {
	engine := r.FormValue("engine")
	if engine == "" {
		engine = "dot"
	}
	return engine
}

func mapFunc(r *http.Request) report.MapFunc {
	f, ok := report.MapFuncRegistry[strings.ToLower(r.FormValue("map_func"))]
	if !ok {
		f = report.ProcessName
	}
	return f
}

func classView(r *http.Request) bool {
	return r.FormValue("class_view") == "true"
}
