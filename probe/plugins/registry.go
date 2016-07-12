package plugins

import (
	"fmt"
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

// Registry maintains a list of available plugins by name.
type Registry struct {
	rootPath          string
	apiVersion        string
	handshakeMetadata map[string]string
	pluginsBySocket   map[string]*Plugin
	lock              sync.RWMutex
	context           context.Context
	cancel            context.CancelFunc
}

// NewRegistry creates a new registry which watches the given dir root for new
// plugins, and adds them.
func NewRegistry(rootPath, apiVersion string, handshakeMetadata map[string]string) (*Registry, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Registry{
		rootPath:          rootPath,
		apiVersion:        apiVersion,
		handshakeMetadata: handshakeMetadata,
		pluginsBySocket:   map[string]*Plugin{},
		context:           ctx,
		cancel:            cancel,
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
	plugins := map[string]*Plugin{}
	// add (or keep) plugins which were found
	for _, path := range sockets {
		if plugin, ok := r.pluginsBySocket[path]; ok {
			plugins[path] = plugin
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
		log.Infof("plugins: added plugin %s", path)
	}
	// remove plugins which weren't found
	for path, plugin := range r.pluginsBySocket {
		if _, ok := plugins[path]; !ok {
			plugin.Close()
			log.Infof("plugins: removed plugin %s", plugin.socket)
		}
	}
	r.pluginsBySocket = plugins
	r.lock.Unlock()
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

// ForEach walks through all the plugins running f for each one.
func (r *Registry) ForEach(f func(p *Plugin)) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	paths := []string{}
	for path := range r.pluginsBySocket {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		f(r.pluginsBySocket[path])
	}
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
	r.ForEach(func(plugin *Plugin) {
		pluginReport, err := plugin.Report()
		if err != nil {
			log.Errorf("plugins: %s: /report error: %v", plugin.socket, err)
		}
		rpt = rpt.Merge(pluginReport)
	})
	return rpt, nil
}

// Close shuts down the registry. It can still be used after this, but will be
// out of date.
func (r *Registry) Close() {
	r.cancel()
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, plugin := range r.pluginsBySocket {
		plugin.Close()
	}
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

	foundReporter := false
	for _, i := range spec.Interfaces {
		if i == "reporter" {
			foundReporter = true
			break
		}
	}
	switch {
	case spec.APIVersion != p.expectedAPIVersion:
		err = fmt.Errorf("incorrect API version: expected %q, got %q", p.expectedAPIVersion, spec.APIVersion)
	case spec.ID == "":
		err = fmt.Errorf("spec must contain an id")
	case spec.Label == "":
		err = fmt.Errorf("spec must contain a label")
	case !foundReporter:
		err = fmt.Errorf("spec must implement the \"reporter\" interface")
	}

	return result, err
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("plugin returned non-200 status code: %s", resp.Status)
	}
	defer resp.Body.Close()
	err = codec.NewDecoder(MaxBytesReader(resp.Body, maxResponseBytes, errResponseTooLarge), &codec.JsonHandle{}).Decode(&result)
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
