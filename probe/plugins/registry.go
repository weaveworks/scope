package plugins

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"

	"github.com/weaveworks/scope/common/backoff"
	"github.com/weaveworks/scope/common/fs"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/report"
)

// Exposed for testing
var (
	transport                 = makeUnixRoundTripper
	maxResponseBytes    int64 = 50 * 1024 * 1024
	errResponseTooLarge       = fmt.Errorf("response must be shorter than 50MB")
	validPluginName           = regexp.MustCompile("^[A-Za-z0-9]+([-][A-Za-z0-9]+)*$")
)

const (
	pluginTimeout    = 500 * time.Millisecond
	scanningInterval = 5 * time.Second
)

// ReportPublisher is an interface for publishing reports immediately
type ReportPublisher interface {
	Publish(rpt report.Report)
}

// Registry maintains a list of available plugins by name.
type Registry struct {
	rootPath          string
	apiVersion        string
	handshakeMetadata map[string]string
	pluginsBySocket   map[string]*Plugin
	lock              sync.RWMutex
	context           context.Context
	cancel            context.CancelFunc
	controlsByPlugin  map[string]report.StringSet
	pluginsByID       map[string]*Plugin
	handlerRegistry   *controls.HandlerRegistry
	publisher         ReportPublisher
}

// NewRegistry creates a new registry which watches the given dir root for new
// plugins, and adds them.
func NewRegistry(rootPath, apiVersion string, handshakeMetadata map[string]string, handlerRegistry *controls.HandlerRegistry, publisher ReportPublisher) (*Registry, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Registry{
		rootPath:          rootPath,
		apiVersion:        apiVersion,
		handshakeMetadata: handshakeMetadata,
		pluginsBySocket:   map[string]*Plugin{},
		context:           ctx,
		cancel:            cancel,
		controlsByPlugin:  map[string]report.StringSet{},
		pluginsByID:       map[string]*Plugin{},
		handlerRegistry:   handlerRegistry,
		publisher:         publisher,
	}
	if err := r.scan(); err != nil {
		r.Close()
		return nil, err
	}
	go r.loop()
	return r, nil
}

// loop periodically rescans for plugins
func (r *Registry) loop() {
	ticker := time.NewTicker(scanningInterval)
	defer ticker.Stop()
	for {
		select {
		case <-r.context.Done():
			return
		case <-ticker.C:
			log.Debugf("plugins: scanning...")
			if err := r.scan(); err != nil {
				log.Warningf("plugins: error: %v", err)
			}
		}
	}
}

// Rescan the plugins directory, load new plugins, and remove missing plugins
func (r *Registry) scan() error {
	sockets, err := r.sockets(r.rootPath)
	if err != nil {
		return err
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	plugins := map[string]*Plugin{}
	pluginsByID := map[string]*Plugin{}
	// add (or keep) plugins which were found
	for _, path := range sockets {
		if plugin, ok := r.pluginsBySocket[path]; ok {
			plugins[path] = plugin
			pluginsByID[plugin.PluginSpec.ID] = plugin
			continue
		}
		tr, err := transport(path, pluginTimeout)
		if err != nil {
			log.Warningf("plugins: error loading plugin %s: %v", path, err)
			continue
		}
		client := &http.Client{Transport: tr, Timeout: pluginTimeout}
		plugin, err := NewPlugin(r.context, path, client, r.apiVersion, r.handshakeMetadata)
		if err != nil {
			log.Warningf("plugins: error loading plugin %s: %v", path, err)
			continue
		}
		plugins[path] = plugin
		pluginsByID[plugin.PluginSpec.ID] = plugin
		log.Infof("plugins: added plugin %s", path)
	}
	// remove plugins which weren't found
	pluginsToClose := map[string]*Plugin{}
	for path, plugin := range r.pluginsBySocket {
		if _, ok := plugins[path]; !ok {
			pluginsToClose[plugin.PluginSpec.ID] = plugin
			log.Infof("plugins: removed plugin %s", plugin.socket)
		}
	}
	r.closePlugins(pluginsToClose)
	r.pluginsBySocket = plugins
	r.pluginsByID = pluginsByID
	return nil
}

// sockets recursively finds all unix sockets under the path provided
func (r *Registry) sockets(path string) ([]string, error) {
	var (
		result []string
		statT  syscall.Stat_t
	)
	// TODO: use of fs.Stat (which is syscall.Stat) here makes this linux specific.
	if err := fs.Stat(path, &statT); err != nil {
		return nil, err
	}
	switch statT.Mode & syscall.S_IFMT {
	case syscall.S_IFDIR:
		files, err := fs.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			fpath := filepath.Join(path, file.Name())
			s, err := r.sockets(fpath)
			if err != nil {
				log.Warningf("plugins: error loading path %s: %v", fpath, err)
			}
			result = append(result, s...)
		}
	case syscall.S_IFSOCK:
		result = append(result, path)
	}
	return result, nil
}

// forEach walks through all the plugins running f for each one.
func (r *Registry) forEach(lock sync.Locker, f func(p *Plugin)) {
	lock.Lock()
	defer lock.Unlock()
	paths := []string{}
	for path := range r.pluginsBySocket {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		f(r.pluginsBySocket[path])
	}
}

// ForEach walks through all the plugins running f for each one.
func (r *Registry) ForEach(f func(p *Plugin)) {
	r.forEach(r.lock.RLocker(), f)
}

// Implementers walks the available plugins fulfilling the given interface
func (r *Registry) Implementers(iface string, f func(p *Plugin)) {
	r.ForEach(func(p *Plugin) {
		for _, piface := range p.Interfaces {
			if piface == iface {
				f(p)
			}
		}
	})
}

// Name implements the Reporter interface
func (r *Registry) Name() string { return "plugins" }

// Report implements the Reporter interface
func (r *Registry) Report() (report.Report, error) {
	rpt := report.MakeReport()
	// All plugins are assumed to (and must) implement reporter
	r.forEach(&r.lock, func(plugin *Plugin) {
		pluginReport, err := plugin.Report()
		if err != nil {
			log.Errorf("plugins: %s: /report error: %v", plugin.socket, err)
		}
		if plugin.Implements("controller") {
			r.updateAndRegisterControlsInReport(&pluginReport)
		}
		rpt = rpt.Merge(pluginReport)
	})
	return rpt, nil
}

func (r *Registry) updateAndRegisterControlsInReport(rpt *report.Report) {
	key := rpt.Plugins.Keys()[0]
	spec, _ := rpt.Plugins.Lookup(key)
	pluginID := spec.ID
	topologies := topologyPointers(rpt)
	var newPluginControls []string
	for _, topology := range topologies {
		newPluginControls = append(newPluginControls, r.updateAndGetControlsInTopology(pluginID, topology)...)
	}
	r.updatePluginControls(pluginID, report.MakeStringSet(newPluginControls...))
}

func topologyPointers(rpt *report.Report) []*report.Topology {
	// We cannot use rpt.Topologies(), because it makes a slice of
	// topology copies and we need original locations to modify
	// them.
	return []*report.Topology{
		&rpt.Endpoint,
		&rpt.Process,
		&rpt.Container,
		&rpt.ContainerImage,
		&rpt.Pod,
		&rpt.Service,
		&rpt.Deployment,
		&rpt.ReplicaSet,
		&rpt.Host,
		&rpt.Overlay,
	}
}

func (r *Registry) updateAndGetControlsInTopology(pluginID string, topology *report.Topology) []string {
	var pluginControls []string
	newControls := report.Controls{}
	for controlID, control := range topology.Controls {
		fakeID := fakeControlID(pluginID, controlID)
		log.Debugf("plugins: replacing control %s with %s", controlID, fakeID)
		control.ID = fakeID
		newControls.AddControl(control)
		pluginControls = append(pluginControls, controlID)
	}
	newNodes := report.Nodes{}
	for name, node := range topology.Nodes {
		log.Debugf("plugins: checking node controls in node %s of %s", name, topology.Label)
		newNode := node.WithID(name)
		var nodeControls []string
		for _, controlID := range node.Controls.Controls {
			log.Debugf("plugins: got node control %s", controlID)
			newControlID := ""
			if _, found := topology.Controls[controlID]; !found {
				log.Debugf("plugins: node control %s does not exist in topology controls", controlID)
				newControlID = controlID
			} else {
				newControlID = fakeControlID(pluginID, controlID)
				log.Debugf("plugins: will replace node control %s with %s", controlID, newControlID)
			}
			nodeControls = append(nodeControls, newControlID)
		}
		newNode.Controls.Controls = report.MakeStringSet(nodeControls...)
		newNodes[newNode.ID] = newNode
	}
	topology.Controls = newControls
	topology.Nodes = newNodes
	return pluginControls
}

func (r *Registry) updatePluginControls(pluginID string, newPluginControls report.StringSet) {
	oldFakePluginControls := r.fakePluginControls(pluginID)
	newFakePluginControls := map[string]xfer.ControlHandlerFunc{}
	for _, controlID := range newPluginControls {
		newFakePluginControls[fakeControlID(pluginID, controlID)] = r.pluginControlHandler
	}
	r.handlerRegistry.Batch(oldFakePluginControls, newFakePluginControls)
	r.controlsByPlugin[pluginID] = newPluginControls
}

// PluginResponse is an extension of xfer.Response that allows plugins
// to send the shortcut reports
type PluginResponse struct {
	xfer.Response
	ShortcutReport *report.Report `json:"shortcutReport,omitempty"`
}

func (r *Registry) pluginControlHandler(req xfer.Request) xfer.Response {
	pluginID, controlID := realPluginAndControlID(req.Control)
	req.Control = controlID
	r.lock.RLock()
	defer r.lock.RUnlock()
	if plugin, found := r.pluginsByID[pluginID]; found {
		response := plugin.Control(req)
		if response.ShortcutReport != nil {
			r.updateAndRegisterControlsInReport(response.ShortcutReport)
			response.ShortcutReport.Shortcut = true
			r.publisher.Publish(*response.ShortcutReport)
		}
		return response.Response
	}
	return xfer.ResponseErrorf("plugin %s not found", pluginID)
}

func realPluginAndControlID(fakeID string) (string, string) {
	parts := strings.SplitN(fakeID, "~", 2)
	if len(parts) != 2 {
		return "", fakeID
	}
	return parts[0], parts[1]
}

// Close shuts down the registry. It can still be used after this, but will be
// out of date.
func (r *Registry) Close() {
	r.cancel()
	r.lock.Lock()
	defer r.lock.Unlock()
	r.closePlugins(r.pluginsByID)
}

func (r *Registry) closePlugins(plugins map[string]*Plugin) {
	var toRemove []string
	for pluginID, plugin := range plugins {
		toRemove = append(toRemove, r.fakePluginControls(pluginID)...)
		delete(r.controlsByPlugin, pluginID)
		plugin.Close()
	}
	r.handlerRegistry.Batch(toRemove, nil)
}

func (r *Registry) fakePluginControls(pluginID string) []string {
	oldPluginControls := r.controlsByPlugin[pluginID]
	var oldFakePluginControls []string
	for _, controlID := range oldPluginControls {
		oldFakePluginControls = append(oldFakePluginControls, fakeControlID(pluginID, controlID))
	}
	return oldFakePluginControls
}

func fakeControlID(pluginID, controlID string) string {
	return fmt.Sprintf("%s~%s", pluginID, controlID)
}

// Plugin is the implementation of a plugin. It is responsible for doing the
// plugin handshake, gathering reports, etc.
type Plugin struct {
	xfer.PluginSpec
	context            context.Context
	socket             string
	expectedAPIVersion string
	handshakeMetadata  url.Values
	client             *http.Client
	cancel             context.CancelFunc
	backoff            backoff.Interface
}

// NewPlugin loads and initializes a new plugin. If client is nil,
// http.DefaultClient will be used.
func NewPlugin(ctx context.Context, socket string, client *http.Client, expectedAPIVersion string, handshakeMetadata map[string]string) (*Plugin, error) {
	id := strings.TrimSuffix(filepath.Base(socket), filepath.Ext(socket))
	if !validPluginName.MatchString(id) {
		return nil, fmt.Errorf("invalid plugin id %q", id)
	}

	params := url.Values{}
	for k, v := range handshakeMetadata {
		params.Add(k, v)
	}

	ctx, cancel := context.WithCancel(ctx)
	plugin := &Plugin{
		PluginSpec:         xfer.PluginSpec{ID: id, Label: id},
		context:            ctx,
		socket:             socket,
		expectedAPIVersion: expectedAPIVersion,
		handshakeMetadata:  params,
		client:             client,
		cancel:             cancel,
	}
	return plugin, nil
}

// Report gets the latest report from the plugin
func (p *Plugin) Report() (result report.Report, err error) {
	result = report.MakeReport()
	defer func() {
		p.setStatus(err)
		result.Plugins = result.Plugins.Add(p.PluginSpec)
		if err != nil {
			result = report.MakeReport()
			result.Plugins = xfer.MakePluginSpecs(p.PluginSpec)
		}
	}()

	if err := p.get("/report", p.handshakeMetadata, &result); err != nil {
		return result, err
	}
	if result.Plugins.Size() != 1 {
		return result, fmt.Errorf("report must contain exactly one plugin (found %d)", result.Plugins.Size())
	}

	key := result.Plugins.Keys()[0]
	spec, _ := result.Plugins.Lookup(key)
	if spec.ID != p.PluginSpec.ID {
		return result, fmt.Errorf("plugin must not change its id (is %q, should be %q)", spec.ID, p.PluginSpec.ID)
	}
	p.PluginSpec = spec

	switch {
	case spec.APIVersion != p.expectedAPIVersion:
		err = fmt.Errorf("incorrect API version: expected %q, got %q", p.expectedAPIVersion, spec.APIVersion)
	case spec.Label == "":
		err = fmt.Errorf("spec must contain a label")
	case !p.Implements("reporter"):
		err = fmt.Errorf("spec must implement the \"reporter\" interface")
	}

	return result, err
}

// Control sends a control message to a plugin
func (p *Plugin) Control(request xfer.Request) (res PluginResponse) {
	var err error
	defer func() {
		p.setStatus(err)
		if err != nil {
			res = PluginResponse{Response: xfer.ResponseError(err)}
		}
	}()

	if p.Implements("controller") {
		err = p.post("/control", p.handshakeMetadata, request, &res)
	} else {
		err = fmt.Errorf("the %s plugin does not implement the controller interface", p.PluginSpec.Label)
	}
	return res
}

// Implements checks if the plugin implements the given interface
func (p *Plugin) Implements(iface string) bool {
	for _, i := range p.PluginSpec.Interfaces {
		if i == iface {
			return true
		}
	}
	return false
}

func (p *Plugin) setStatus(err error) {
	if err == nil {
		p.Status = "ok"
	} else {
		p.Status = fmt.Sprintf("error: %v", err)
	}
}

func (p *Plugin) get(path string, params url.Values, result interface{}) error {
	// Context here lets us either timeout req. or cancel it in Plugin.Close
	ctx, cancel := context.WithTimeout(p.context, pluginTimeout)
	defer cancel()
	resp, err := ctxhttp.Get(ctx, p.client, fmt.Sprintf("http://plugin%s?%s", path, params.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("plugin returned non-200 status code: %s", resp.Status)
	}
	return getResult(resp.Body, result)
}

func (p *Plugin) post(path string, params url.Values, data interface{}, result interface{}) error {
	// Context here lets us either timeout req. or cancel it in Plugin.Close
	ctx, cancel := context.WithTimeout(p.context, pluginTimeout)
	defer cancel()
	buf := &bytes.Buffer{}
	if err := codec.NewEncoder(buf, &codec.JsonHandle{}).Encode(data); err != nil {
		return fmt.Errorf("encoding error: %s", err)
	}
	resp, err := ctxhttp.Post(ctx, p.client, fmt.Sprintf("http://plugin%s?%s", path, params.Encode()), "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("plugin returned non-200 status code: %s", resp.Status)
	}
	return getResult(resp.Body, result)
}

func getResult(body io.ReadCloser, result interface{}) error {
	err := codec.NewDecoder(MaxBytesReader(body, maxResponseBytes, errResponseTooLarge), &codec.JsonHandle{}).Decode(&result)
	if err == errResponseTooLarge {
		return err
	}
	if err != nil {
		return fmt.Errorf("decoding error: %s", err)
	}
	return nil
}

// Close closes the client
func (p *Plugin) Close() {
	if p.backoff != nil {
		p.backoff.Stop()
	}
	p.cancel()
}
