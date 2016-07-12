package plugins

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"path/filepath"
	"sort"
	"syscall"
	"testing"
	"time"

	"github.com/paypal/ionet"

	fs_hook "github.com/weaveworks/scope/common/fs"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fs"
	"github.com/weaveworks/scope/test/reflect"
)

func stubTransport(fn func(socket string, timeout time.Duration) (http.RoundTripper, error)) {
	transport = fn
}
func restoreTransport() { transport = makeUnixRoundTripper }

type readWriteCloseRoundTripper struct {
	io.ReadWriteCloser
}

func (rwc readWriteCloseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	conn := &closeableConn{
		Conn:   &ionet.Conn{R: rwc, W: rwc},
		Closer: rwc,
	}
	client := httputil.NewClientConn(conn, nil)
	defer client.Close()
	return client.Do(req)
}

// closeableConn gives us an overrideable Close, where ionet.Conn does not.
type closeableConn struct {
	net.Conn
	io.Closer
}

func (c *closeableConn) Close() error {
	c.Conn.Close()
	return c.Closer.Close()
}

type mockPlugin struct {
	t       *testing.T
	Name    string
	Handler http.Handler
}

func (p mockPlugin) dir() string {
	return "/plugins"
}

func (p mockPlugin) path() string {
	return filepath.Join(p.dir(), p.base())
}

func (p mockPlugin) base() string {
	return p.Name + ".sock"
}

func (p mockPlugin) file() fs.File {
	incomingR, incomingW := io.Pipe()
	outgoingR, outgoingW := io.Pipe()
	// TODO: This is a terrible hack of a little http server. Really, we should
	// implement some sort of fs.File -> net.Listener bridge and run an net/http
	// server on that.
	go func() {
		for {
			conn := httputil.NewServerConn(&ionet.Conn{R: incomingR, W: outgoingW}, nil)
			req, err := conn.Read()
			if err == io.EOF {
				outgoingW.Close()
				return
			} else if err != nil {
				p.t.Fatal(err)
			}
			resp := httptest.NewRecorder()
			p.Handler.ServeHTTP(resp, req)
			fmt.Fprintf(outgoingW, "HTTP/1.1 %d %s\nContent-Length: %d\n\n%s", resp.Code, http.StatusText(resp.Code), resp.Body.Len(), resp.Body.String())
		}
	}()
	return fs.File{
		FName:   p.base(),
		FWriter: incomingW,
		FReader: outgoingR,
		FStat:   syscall.Stat_t{Mode: syscall.S_IFSOCK},
	}
}

type chanWriter chan []byte

func (w chanWriter) Write(p []byte) (int, error) {
	w <- p
	return len(p), nil
}

func (w chanWriter) Close() error {
	close(w)
	return nil
}

func setup(t *testing.T, sockets ...fs.Entry) fs.Entry {
	mockFS := fs.Dir("", fs.Dir("plugins", sockets...))
	fs_hook.Mock(
		mockFS)

	stubTransport(func(socket string, timeout time.Duration) (http.RoundTripper, error) {
		f, err := mockFS.Open(socket)
		return readWriteCloseRoundTripper{f}, err
	})

	return mockFS
}

func restore(t *testing.T) {
	fs_hook.Restore()
	restoreTransport()
}

type iterator func(func(*Plugin))

func checkLoadedPlugins(t *testing.T, forEach iterator, expected []xfer.PluginSpec) {
	var plugins []xfer.PluginSpec
	forEach(func(p *Plugin) {
		plugins = append(plugins, p.PluginSpec)
	})
	sort.Sort(xfer.PluginSpecsByID(plugins))
	if !reflect.DeepEqual(plugins, expected) {
		t.Fatalf(test.Diff(expected, plugins))
	}
}

func checkLoadedPluginIDs(t *testing.T, forEach iterator, expectedIDs []string) {
	var pluginIDs []string
	forEach(func(p *Plugin) {
		pluginIDs = append(pluginIDs, p.ID)
	})
	sort.Strings(pluginIDs)
	if len(pluginIDs) != len(expectedIDs) {
		t.Fatalf("Expected plugins %q, got: %q", expectedIDs, pluginIDs)
	}
	for i, id := range pluginIDs {
		if id != expectedIDs[i] {
			t.Fatalf("Expected plugins %q, got: %q", expectedIDs, pluginIDs)
		}
	}
}

// stringHandler returns an http.Handler which just prints the given string
func stringHandler(status int, j string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/report" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
		fmt.Fprint(w, j)
	})
}

func TestRegistryLoadsExistingPlugins(t *testing.T) {
	setup(
		t,
		mockPlugin{
			t:       t,
			Name:    "testPlugin",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"testPlugin","label":"testPlugin","interfaces":["reporter"],"api_version":"1"}]}`),
		}.file(),
	)
	defer restore(t)

	root := "/plugins"
	r, err := NewRegistry(root, "1", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	checkLoadedPluginIDs(t, r.ForEach, []string{"testPlugin"})
}

func TestRegistryLoadsExistingPluginsEvenWhenOneFails(t *testing.T) {
	setup(
		t,
		// TODO: This first one needs to fail
		fs.Dir("fail",
			mockPlugin{
				t:       t,
				Name:    "aFailure",
				Handler: stringHandler(http.StatusInternalServerError, `Internal Server Error`),
			}.file(),
		),
		mockPlugin{
			t:       t,
			Name:    "testPlugin",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"testPlugin","label":"testPlugin","interfaces":["reporter"],"api_version":"1"}]}`),
		}.file(),
	)
	defer restore(t)

	root := "/plugins"
	r, err := NewRegistry(root, "1", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	checkLoadedPlugins(t, r.ForEach, []xfer.PluginSpec{
		{
			ID:     "aFailure",
			Label:  "aFailure",
			Status: "error: plugin returned non-200 status code: 500 Internal Server Error",
		},
		{
			ID:         "testPlugin",
			Label:      "testPlugin",
			Interfaces: []string{"reporter"},
			APIVersion: "1",
			Status:     "ok",
		},
	})
}

func TestRegistryDiscoversNewPlugins(t *testing.T) {
	mockFS := setup(t)
	defer restore(t)

	root := "/plugins"
	r, err := NewRegistry(root, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	checkLoadedPluginIDs(t, r.ForEach, []string{})

	// Add the new plugin
	plugin := mockPlugin{
		t:       t,
		Name:    "testPlugin",
		Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"testPlugin","label":"testPlugin","interfaces":["reporter"]}]}`),
	}
	mockFS.Add(plugin.dir(), plugin.file())
	if err := r.scan(); err != nil {
		t.Fatal(err)
	}

	r.Report()
	checkLoadedPluginIDs(t, r.ForEach, []string{"testPlugin"})
}

func TestRegistryRemovesPlugins(t *testing.T) {
	plugin := mockPlugin{
		t:       t,
		Name:    "testPlugin",
		Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"testPlugin","label":"testPlugin","interfaces":["reporter"]}]}`),
	}
	mockFS := setup(t, plugin.file())
	defer restore(t)

	root := "/plugins"
	r, err := NewRegistry(root, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	checkLoadedPluginIDs(t, r.ForEach, []string{"testPlugin"})

	// Remove the plugin
	mockFS.Remove(plugin.path())
	if err := r.scan(); err != nil {
		t.Fatal(err)
	}

	checkLoadedPluginIDs(t, r.ForEach, []string{})
}

func TestRegistryUpdatesPluginsWhenTheyChange(t *testing.T) {
	resp := `{"Plugins":[{"id":"testPlugin","label":"testPlugin","interfaces":["reporter"]}]}`
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, resp)
	})
	plugin := mockPlugin{
		t:       t,
		Name:    "testPlugin",
		Handler: handler,
	}
	setup(t, plugin.file())
	defer restore(t)

	root := "/plugins"
	r, err := NewRegistry(root, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	checkLoadedPluginIDs(t, r.ForEach, []string{"testPlugin"})

	// Update the plugin. Just change what the handler will respond with.
	resp = `{"Plugins":[{"id":"testPlugin","label":"updatedPlugin","interfaces":["reporter"]}]}`

	r.Report()
	checkLoadedPlugins(t, r.ForEach, []xfer.PluginSpec{
		{
			ID:         "testPlugin",
			Label:      "updatedPlugin",
			Interfaces: []string{"reporter"},
			Status:     "ok",
		},
	})
}

func TestRegistryReturnsPluginsByInterface(t *testing.T) {
	setup(
		t,
		mockPlugin{
			t:       t,
			Name:    "plugin1",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"plugin1","label":"plugin1","interfaces":["reporter"]}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "plugin2",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"plugin2","label":"plugin2","interfaces":["other"]}]}`),
		}.file(),
	)
	defer restore(t)

	root := "/plugins"
	r, err := NewRegistry(root, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	checkLoadedPluginIDs(t, r.ForEach, []string{"plugin1", "plugin2"})
	checkLoadedPluginIDs(t, func(fn func(*Plugin)) { r.Implementers("reporter", fn) }, []string{"plugin1"})
	checkLoadedPluginIDs(t, func(fn func(*Plugin)) { r.Implementers("other", fn) }, []string{"plugin2"})
}

func TestRegistryHandlesConflictingPlugins(t *testing.T) {
	setup(
		t,
		mockPlugin{
			t:       t,
			Name:    "plugin1",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"plugin1","label":"plugin1","interfaces":["reporter"]}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "plugin1",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"plugin1","label":"plugin2","interfaces":["reporter","other"]}]}`),
		}.file(),
	)
	defer restore(t)

	root := "/plugins"
	r, err := NewRegistry(root, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	// Should just have the second one (we just log conflicts)
	checkLoadedPluginIDs(t, r.ForEach, []string{"plugin1"})
	checkLoadedPluginIDs(t, func(fn func(*Plugin)) { r.Implementers("other", fn) }, []string{"plugin1"})
}

func TestRegistryRejectsErroneousPluginResponses(t *testing.T) {
	setup(
		t,
		mockPlugin{
			t:       t,
			Name:    "okPlugin",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"okPlugin","label":"okPlugin","interfaces":["reporter"]}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "non200ResponseCode",
			Handler: stringHandler(http.StatusInternalServerError, `{"Plugins":[{"id":"non200ResponseCode","label":"non200ResponseCode","interfaces":["reporter"]}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "noLabel",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"noLabel","interfaces":["reporter"]}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "noInterface",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"noInterface","label":"noInterface","interfaces":[]}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "wrongApiVersion",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"wrongApiVersion","label":"wrongApiVersion","interfaces":["reporter"],"api_version":"foo"}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "nonJSONResponseBody",
			Handler: stringHandler(http.StatusOK, `notJSON`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "changedId",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"differentId","label":"changedId","interfaces":["reporter"]}]}`),
		}.file(),
	)
	defer restore(t)

	root := "/plugins"
	r, err := NewRegistry(root, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	checkLoadedPlugins(t, r.ForEach, []xfer.PluginSpec{
		{
			ID:     "changedId",
			Label:  "changedId",
			Status: `error: plugin must not change its id (is "differentId", should be "changedId")`,
		},
		{
			ID:         "noInterface",
			Label:      "noInterface",
			Interfaces: []string{},
			Status:     `error: spec must implement the "reporter" interface`,
		},
		{
			ID:         "noLabel",
			Interfaces: []string{"reporter"},
			Status:     `error: spec must contain a label`,
		},
		{
			ID:     "non200ResponseCode",
			Label:  "non200ResponseCode",
			Status: "error: plugin returned non-200 status code: 500 Internal Server Error",
		},
		{
			ID:     "nonJSONResponseBody",
			Label:  "nonJSONResponseBody",
			Status: "error: decoding error: [pos 4]: json: expecting ull: got otJ",
		},
		{
			ID:         "okPlugin",
			Label:      "okPlugin",
			Interfaces: []string{"reporter"},
			Status:     `ok`,
		},
		{
			ID:         "wrongApiVersion",
			Label:      "wrongApiVersion",
			Interfaces: []string{"reporter"},
			APIVersion: "foo",
			Status:     `error: incorrect API version: expected "", got "foo"`,
		},
	})
}

func TestRegistryRejectsPluginResponsesWhichAreTooLarge(t *testing.T) {
	description := ""
	for i := 0; i < 129; i++ {
		description += "a"
	}
	response := fmt.Sprintf(
		`{
			"Plugins": [
				{
					"id": "foo",
					"label": "foo",
					"description": %q,
					"interfaces": ["reporter"]
				}
			]
		}`,
		description,
	)
	setup(t, mockPlugin{t: t, Name: "foo", Handler: stringHandler(http.StatusOK, response)}.file())
	oldMaxResponseBytes := maxResponseBytes
	maxResponseBytes = 128

	defer func() {
		maxResponseBytes = oldMaxResponseBytes
		restore(t)
	}()

	root := "/plugins"
	r, err := NewRegistry(root, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	checkLoadedPlugins(t, r.ForEach, []xfer.PluginSpec{
		{ID: "foo", Label: "foo", Status: `error: response must be shorter than 50MB`},
	})
}
