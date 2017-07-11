package detailed

import (
	"bytes"
	"net/url"
	"strings"
	"text/template"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"

	"github.com/ugorji/go/codec"
)

// MetricLink describes a URL referencing a metric.
type MetricLink struct {
	// References the metric id
	ID       string `json:"id,omitempty"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	Priority int    `json:"priority"`
}

// Variable name for the query within the metrics graph url
const urlQueryVarName = ":query"

var (
	// As configured by the user
	metricsGraphURL = ""

	// Available metric links
	linkTemplates = []MetricLink{
		{ID: docker.CPUTotalUsage, Label: "CPU", Priority: 1},
		{ID: docker.MemoryUsage, Label: "Memory", Priority: 2},
	}

	// Prometheus queries for topologies
	topologyQueries = map[string]map[string]*template.Template{
		report.Pod: {
			// `container_memory_usage_bytes` is provided by cAdvisor in Kubelets.
			docker.MemoryUsage:   parsedTemplate(`sum(container_memory_usage_bytes{pod_name="{{.Label}}"})`),
			// `container_cpu_usage_seconds_total` is provided by cAdvisor in Kubelets.
			docker.CPUTotalUsage: parsedTemplate(`sum(rate(container_cpu_usage_seconds_total{pod_name="{{.Label}}"}[1m]))`),
		},
		report.Deployment: {
			// `container_memory_usage_bytes` is provided by cAdvisor in Kubelets.
			docker.MemoryUsage:   parsedTemplate(`sum(container_memory_usage_bytes{pod_name=~"{{.Label}}-[^-]+-[^-]+"})`),
			// `container_cpu_usage_seconds_total` is provided by cAdvisor in Kubelets.
			docker.CPUTotalUsage: parsedTemplate(`sum(rate(container_cpu_usage_seconds_total{pod_name=~"{{.Label}}-[^-]+-[^-]+"}[1m]))`),
		},
		report.DaemonSet: {
			/*
			A recording rule defines `namespace_name:container_memory_usage_bytes:sum`:

			    namespace_name:container_memory_usage_bytes:sum =
			       sum by (namespace, name) (
				  sum(container_memory_usage_bytes{image!=""}) by (pod_name, namespace)
				* on (pod_name) group_left(name)
				  k8s_pod_labels{job="monitoring/kube-api-exporter"}
			       )

			Additionally, we filter by `monitor=""` for cortex-only data.
			 */

			docker.MemoryUsage:   parsedTemplate(`namespace_name:container_memory_usage_bytes:sum{name="{{.Label}}",monitor=""}`),
			/*
			A recording rule defines `namespace_name:container_cpu_usage_seconds_total:sum_rate`:

			    namespace_name:container_cpu_usage_seconds_total:sum_rate =
			       sum by (namespace, name) (
				  sum(rate(container_cpu_usage_seconds_total{image!=""}[5m])) by (pod_name, namespace)
				* on (pod_name) group_left(name)
				  k8s_pod_labels{job="monitoring/kube-api-exporter"}
			       )
			 */
			docker.CPUTotalUsage: parsedTemplate(`namespace_name:container_cpu_usage_seconds_total:sum_rate{name="{{.Label}}"}`),
		},
		report.Service: {
			/*
			A recording rule defines `namespace_name:container_memory_usage_bytes:sum`:

			    namespace_name:container_memory_usage_bytes:sum =
			       sum by (namespace, name) (
				  sum(container_memory_usage_bytes{image!=""}) by (pod_name, namespace)
				* on (pod_name) group_left(name)
				  k8s_pod_labels{job="monitoring/kube-api-exporter"}
			       )

			Additionally, we filter by `monitor=""` for cortex-only data.
			 */
			docker.MemoryUsage:   parsedTemplate(`namespace_name:container_memory_usage_bytes:sum{name="{{.Label}}",monitor=""}`),

			/*
			A recording rule defines `namespace_name:container_cpu_usage_seconds_total:sum_rate`:

			    namespace_name:container_cpu_usage_seconds_total:sum_rate =
			       sum by (namespace, name) (
				  sum(rate(container_cpu_usage_seconds_total{image!=""}[5m])) by (pod_name, namespace)
				* on (pod_name) group_left(name)
				  k8s_pod_labels{job="monitoring/kube-api-exporter"}
			       )
			 */
			docker.CPUTotalUsage: parsedTemplate(`namespace_name:container_cpu_usage_seconds_total:sum_rate{name="{{.Label}}"}`),
		},
	}
)

// SetMetricsGraphURL sets the URL we deduce our eventual metric link from.
// Supports placeholders such as `:orgID` and `:query`. An empty url disables
// this feature. If the `:query` part is missing, a JSON version will be
// appended, see `queryParamsAsJSON()` for more info.
func SetMetricsGraphURL(url string) {
	metricsGraphURL = url
}

// NodeMetricLinks returns the links of a node. The links are collected
// by a predefined set but filtered depending on whether a query
// is configured or not for the particular topology.
func NodeMetricLinks(_ report.Report, n report.Node) []MetricLink {
	if metricsGraphURL == "" {
		return nil
	}

	queries := topologyQueries[n.Topology]
	if len(queries) == 0 {
		return nil
	}

	links := []MetricLink{}
	for _, link := range linkTemplates {
		if _, ok := queries[link.ID]; ok {
			links = append(links, link)
		}
	}

	return links
}

// RenderMetricLinks executes the templated links by supplying the node summary as data.
// It returns the modified summary.
func RenderMetricLinks(summary NodeSummary, n report.Node) NodeSummary {
	queries := topologyQueries[n.Topology]
	if len(queries) == 0 || len(summary.MetricLinks) == 0 {
		return summary
	}

	links := []MetricLink{}
	var bs bytes.Buffer
	for _, link := range summary.MetricLinks {
		tpl := queries[link.ID]
		if tpl == nil {
			continue
		}

		bs.Reset()
		if err := tpl.Execute(&bs, summary); err != nil {
			continue
		}

		link.URL = buildURL(bs.String())
		links = append(links, link)
	}
	summary.MetricLinks = links

	return summary
}

// buildURL puts together the URL by looking at the configured
// `metricsGraphURL`.
func buildURL(query string) string {
	if strings.Contains(metricsGraphURL, urlQueryVarName) {
		return strings.Replace(metricsGraphURL, urlQueryVarName, url.PathEscape(query), -1)
	}

	params, err := queryParamsAsJSON(query)
	if err != nil {
		return ""
	}

	if metricsGraphURL[len(metricsGraphURL)-1] != '/' {
		metricsGraphURL += "/"
	}

	return metricsGraphURL + url.PathEscape(params)
}

// queryParamsAsJSON packs the query into a JSON of the
// format `{"cells":[{"queries":[$query]}]}`.
func queryParamsAsJSON(query string) (string, error) {
	type cell struct {
		Queries []string `json:"queries"`
	}
	type queryParams struct {
		Cells []cell `json:"cells"`
	}
	params := &queryParams{[]cell{{[]string{query}}}}

	buf := &bytes.Buffer{}
	encoder := codec.NewEncoder(buf, &codec.JsonHandle{})
	if err := encoder.Encode(params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// parsedTemplate initializes unnamed text templates.
func parsedTemplate(query string) *template.Template {
	tpl, err := template.New("").Parse(query)
	if err != nil {
		panic(err)
	}

	return tpl
}
