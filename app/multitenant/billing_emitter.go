package multitenant

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"math"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	billing "github.com/weaveworks/billing-client"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

// BillingEmitterConfig has everything we need to make a billing emitter
type BillingEmitterConfig struct {
	Enabled         bool
	DefaultInterval time.Duration
	UserIDer        UserIDer
}

// RegisterFlags registers the billing emitter flags with the main flag set.
func (cfg *BillingEmitterConfig) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&cfg.Enabled, "app.billing.enabled", false, "enable emitting billing info")
	f.DurationVar(&cfg.DefaultInterval, "app.billing.default-publish-interval", 3*time.Second, "default publish interval to assume for reports")
}

// BillingEmitter is the billing emitter
type BillingEmitter struct {
	app.Collector
	BillingEmitterConfig
	billing *billing.Client

	sync.Mutex
	intervalCache map[string]time.Duration
	rounding      map[string]float64
}

// NewBillingEmitter changes a new billing emitter which emits billing events
func NewBillingEmitter(upstream app.Collector, billingClient *billing.Client, cfg BillingEmitterConfig) (*BillingEmitter, error) {
	return &BillingEmitter{
		Collector:            upstream,
		billing:              billingClient,
		BillingEmitterConfig: cfg,
		intervalCache:        make(map[string]time.Duration),
		rounding:             make(map[string]float64),
	}, nil
}

// Add implements app.Collector
func (e *BillingEmitter) Add(ctx context.Context, rep report.Report, buf []byte) error {
	now := time.Now().UTC()
	userID, err := e.UserIDer(ctx)
	if err != nil {
		// Underlying collector needs to get userID too, so it's OK to abort
		// here. If this fails, so will underlying collector so no point
		// proceeding.
		return err
	}
	rowKey, colKey := calculateDynamoKeys(userID, now)

	interval, nodes := e.scanReport(rep)
	// Cache the last-known value of interval for this user, and use
	// it if we didn't find one in this report.
	e.Lock()
	if interval != 0 {
		e.intervalCache[userID] = interval
	} else {
		if lastKnown, found := e.intervalCache[userID]; found {
			interval = lastKnown
		} else {
			interval = e.DefaultInterval
		}
	}
	// Billing takes an integer number of seconds, so keep track of the amount lost to rounding
	nodeSeconds := interval.Seconds()*float64(nodes) + e.rounding[userID]
	rounding := nodeSeconds - math.Floor(nodeSeconds)
	e.rounding[userID] = rounding
	e.Unlock()

	hasher := sha256.New()
	hasher.Write(buf)
	hash := "sha256:" + base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	weaveNetCount := 0
	if hasWeaveNet(rep) {
		weaveNetCount = 1
	}

	amounts := billing.Amounts{
		billing.ContainerSeconds: int64(interval/time.Second) * int64(len(rep.Container.Nodes)),
		billing.NodeSeconds:      int64(nodeSeconds),
		billing.WeaveNetSeconds:  int64(interval/time.Second) * int64(weaveNetCount),
	}
	metadata := map[string]string{
		"row_key": rowKey,
		"col_key": colKey,
	}

	err = e.billing.AddAmounts(
		hash,
		userID,
		now,
		amounts,
		metadata,
	)
	if err != nil {
		// No return, because we want to proceed even if we fail to emit
		// billing data, so that defects in the billing system don't break
		// report collection. Just log the fact & carry on.
		log.Errorf("Failed emitting billing data: %v", err)
	}

	return e.Collector.Add(ctx, rep, buf)
}

func commandParameter(cmd, flag string) (string, bool) {
	i := strings.Index(cmd, flag)
	if i != -1 {
		// here we expect the command looks like `-foo=bar` or `-foo bar`
		aft := strings.Fields(cmd[i+len(flag):])
		if len(aft) > 0 && len(aft[0]) > 0 {
			if aft[0][0] == '=' {
				return aft[0][1:], true
			}
			return aft[0], true
		}
	}
	return "", false
}

func intervalFromCommand(cmd string) string {
	if strings.Contains(cmd, "scope") {
		if publishInterval, ok := commandParameter(cmd, "probe.publish.interval"); ok {
			// If spy interval is higher than publish interval, some reports will have no process data
			if spyInterval, ok := commandParameter(cmd, "spy.interval"); ok {
				pubDuration, err1 := time.ParseDuration(publishInterval)
				spyDuration, err2 := time.ParseDuration(spyInterval)
				if err1 == nil && err2 == nil && spyDuration > pubDuration {
					return spyInterval
				}
			}
			return publishInterval
		}
	}
	return ""
}

// scanReport counts the nodes tries to find any custom report interval
// of this report. If it is malformed, or not set, it returns zero.
func (e *BillingEmitter) scanReport(r report.Report) (time.Duration, int) {
	nHosts := 0
	// We scan the host nodes looking for ones reported by a per-node probe;
	// the Kubernetes cluster probe also makes host nodes but they only have a few fields set
	for _, h := range r.Host.Nodes {
		// Relying here on Uptime being something that changes in each report, hence will be in a delta report
		if _, ok := h.Latest.Lookup(report.Uptime); ok {
			nHosts++
		}
	}
	if r.Window != 0 {
		return r.Window, nHosts
	}
	var inter string
	for _, c := range r.Container.Nodes {
		if cmd, ok := c.Latest.Lookup(report.DockerContainerCommand); ok {
			if inter = intervalFromCommand(cmd); inter != "" {
				break
			}
		}
	}
	if inter == "" { // not found in containers: look in processes
		for _, c := range r.Process.Nodes {
			if cmd, ok := c.Latest.Lookup(report.Cmdline); ok {
				if inter = intervalFromCommand(cmd); inter != "" {
					break
				}
			}
		}
	}
	if inter == "" {
		return 0, nHosts
	}
	d, err := time.ParseDuration(inter)
	if err != nil {
		return 0, nHosts
	}
	return d, nHosts
}

// Tries to determine if this report came from a host running Weave Net
func hasWeaveNet(r report.Report) bool {
	for _, n := range r.Overlay.Nodes {
		overlayType, _ := report.ParseOverlayNodeID(n.ID)
		if overlayType == report.WeaveOverlayPeerPrefix {
			return true
		}
	}

	return false
}

// Close shuts down the billing emitter and billing client flushing events.
func (e *BillingEmitter) Close() {
	e.Collector.Close()
	_ = e.billing.Close()
}
