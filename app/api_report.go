package app

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

// Raw report handler
func makeRawReportHandler(rep Reporter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		report, err := rep.Report(ctx, time.Now())
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err)
			return
		}
		respondWith(w, http.StatusOK, report)
	}
}

type probeDesc struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	Version  string    `json:"version"`
	LastSeen time.Time `json:"lastSeen"`
}

// Probe handler
func makeProbeHandler(rep Reporter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		rpt, err := rep.Report(ctx, time.Now())
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err)
			return
		}
		result := []probeDesc{}
		for _, n := range rpt.Host.Nodes {
			id, _ := n.Latest.Lookup(report.ControlProbeID)
			hostname, _ := n.Latest.Lookup(host.HostName)
			version, dt, _ := n.Latest.LookupEntry(host.ScopeVersion)
			result = append(result, probeDesc{
				ID:       id,
				Hostname: hostname,
				Version:  version,
				LastSeen: dt,
			})
		}
		respondWith(w, http.StatusOK, result)
	}
}
