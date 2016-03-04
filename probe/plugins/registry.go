package plugins

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
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
	transport = makeUnixRoundTripper
)

const (
	pluginTimeout   = 500 * time.Millisecond
	pluginRetry     = 5 * time.Second
	pollingInterval = 5 * time.Second
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
	ticker := time.NewTicker(pollingInterval)
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
		plugins[path] = NewPlugin(r.context, path, client, r.apiVersion, r.handshakeMetadata)
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
	context context.Context
	socket  string
	client  *http.Client
	cancel  context.CancelFunc
	backoff backoff.Interface
}

// NewPlugin loads and initializes a new plugin. If client is nil,
// http.DefaultClient will be used.
func NewPlugin(ctx context.Context, socket string, client *http.Client, expectedAPIVersion string, handshakeMetadata map[string]string) *Plugin {
	params := url.Values{}
	for k, v := range handshakeMetadata {
		params.Add(k, v)
	}

	ctx, cancel := context.WithCancel(ctx)
	p := &Plugin{context: ctx, socket: socket, client: client, cancel: cancel}
	f := p.handshake(ctx, expectedAPIVersion, params)
	f() // try the first time synchronously
	p.backoff = backoff.New(f, "plugin handshake")
	go p.backoff.Start()
	return p
}

type handshakeResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  string   `json:"api_version,omitempty"`
}

// handshake tries the handshake with this plugin.
func (p *Plugin) handshake(ctx context.Context, expectedAPIVersion string, params url.Values) func() (bool, error) {
	return func() (bool, error) {
		var resp handshakeResponse
		if err := p.get("/", params, &resp); err != nil {
			return err == context.Canceled, fmt.Errorf("plugins: error loading plugin %s: %v", p.socket, err)
		}

		if resp.Name == "" {
			return false, fmt.Errorf("plugins: error loading plugin %s: plugin did not provide a name", p.socket)
		}
		if resp.APIVersion != expectedAPIVersion {
			return false, fmt.Errorf("plugins: error loading plugin %s: plugin did not provide correct API version: expected %q, got %q", p.socket, expectedAPIVersion, resp.APIVersion)
		}
		p.ID, p.Label = resp.Name, resp.Name
		p.Description = resp.Description
		p.Interfaces = resp.Interfaces
		log.Infof("plugins: loaded plugin %s: %s", p.ID, strings.Join(p.Interfaces, ", "))
		return true, nil
	}
}

// Report gets the latest report from the plugin
func (p *Plugin) Report() (report.Report, error) {
	result := report.MakeReport()
	err := p.get("/report", nil, &result)
	return result, err
}

// TODO(paulbellamy): better error handling on wrong status codes
func (p *Plugin) get(path string, params url.Values, result interface{}) error {
	ctx, cancel := context.WithTimeout(p.context, pluginTimeout)
	defer cancel()
	resp, err := ctxhttp.Get(ctx, p.client, fmt.Sprintf("unix://%s?%s", path, params.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return codec.NewDecoder(resp.Body, &codec.JsonHandle{}).Decode(&result)
}

// Close closes the client
func (p *Plugin) Close() {
	p.backoff.Stop()
	p.cancel()
}
