package app

import (
	"net/http"
	"time"

	"context"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

// Raw report handler
func makeRawReportHandler(rep Reporter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		timestamp := deserializeTimestamp(r.URL.Query().Get("timestamp"))
		rawReport, err := rep.Report(ctx, timestamp)
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err)
			return
		}
		censorCfg := report.GetCensorConfigFromRequest(r)
		respondWith(w, http.StatusOK, report.CensorRawReport(rawReport, censorCfg))
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
		r.ParseForm()
		if _, sparse := r.Form["sparse"]; sparse {
			// if we have reports, we must have connected probes
			hasProbes, err := rep.HasReports(ctx, time.Now())
			if err != nil {
				respondWith(w, http.StatusInternalServerError, err)
			}
			respondWith(w, http.StatusOK, hasProbes)
			return
		}
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
