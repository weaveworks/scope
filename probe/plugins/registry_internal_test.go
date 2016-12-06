package plugins

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/paypal/ionet"
	"github.com/ugorji/go/codec"

	fs_hook "github.com/weaveworks/common/fs"
	"github.com/weaveworks/common/test/fs"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func testRegistry(t *testing.T, apiVersion string) *Registry {
	handlerRegistry := controls.NewDefaultHandlerRegistry()
	root := "/plugins"
	r, err := NewRegistry(root, apiVersion, nil, handlerRegistry, nil)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

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

type testResponse struct {
	Status int
	Body   string
}

type testResponseMap map[string]testResponse

// mapStringHandler returns an http.Handler which just prints the given string for each path
func mapStringHandler(responses testResponseMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if response, found := responses[r.URL.Path]; found {
			w.WriteHeader(response.Status)
			fmt.Fprint(w, response.Body)
		} else {
			http.NotFound(w, r)
		}
	})
}

// stringHandler returns an http.Handler which just prints the given string
func stringHandler(status int, j string) http.Handler {
	return mapStringHandler(testResponseMap{"/report": {status, j}})
}

type testHandlerRegistryBackend struct {
	handlers map[string]xfer.ControlHandlerFunc
	t        *testing.T
	mtx      sync.Mutex
}

func newTestHandlerRegistryBackend(t *testing.T) *testHandlerRegistryBackend {
	return &testHandlerRegistryBackend{
		handlers: map[string]xfer.ControlHandlerFunc{},
		t:        t,
	}
}

// Lock locks the backend, so the batch insertions or removals can be
// performed.
func (b *testHandlerRegistryBackend) Lock() {
	b.mtx.Lock()
}

// Unlock unlocks the backend.
func (b *testHandlerRegistryBackend) Unlock() {
	b.mtx.Unlock()
}

// Register a new control handler under a given id.
func (b *testHandlerRegistryBackend) Register(control string, f xfer.ControlHandlerFunc) {
	b.handlers[control] = f
}

// Rm deletes the handler for a given name.
func (b *testHandlerRegistryBackend) Rm(control string) {
	delete(b.handlers, control)
}

// Handler gets the handler for the given id.
func (b *testHandlerRegistryBackend) Handler(control string) (xfer.ControlHandlerFunc, bool) {
	handler, ok := b.handlers[control]
	return handler, ok
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

	r := testRegistry(t, "1")
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

	r := testRegistry(t, "1")
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

	r := testRegistry(t, "")
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

	r := testRegistry(t, "")
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

	r := testRegistry(t, "")
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

	r := testRegistry(t, "")
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

	r := testRegistry(t, "")
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
			Name:    "changedID",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"differentID","label":"changedID","interfaces":["reporter"]}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "moreThanOnePlugin",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"moreThanOnePlugin","label":"moreThanOnePlugin","interfaces":["reporter"]}, {"id":"haha","label":"haha","interfaces":["reporter"]}]}`),
		}.file(),
	)
	defer restore(t)

	r := testRegistry(t, "")
	defer r.Close()

	r.Report()
	checkLoadedPlugins(t, r.ForEach, []xfer.PluginSpec{
		{
			ID:     "changedID",
			Label:  "changedID",
			Status: `error: plugin must not change its id (is "differentID", should be "changedID")`,
		},
		{
			ID:     "moreThanOnePlugin",
			Label:  "moreThanOnePlugin",
			Status: `error: report must contain exactly one plugin (found 2)`,
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

	r := testRegistry(t, "")
	defer r.Close()

	r.Report()
	checkLoadedPlugins(t, r.ForEach, []xfer.PluginSpec{
		{ID: "foo", Label: "foo", Status: `error: response must be shorter than 50MB`},
	})
}

func TestRegistryChecksForValidPluginIDs(t *testing.T) {
	setup(
		t,
		mockPlugin{
			t:       t,
			Name:    "testPlugin",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"testPlugin","label":"testPlugin","interfaces":["reporter"],"api_version":"1"}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "P-L-U-G-I-N",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"P-L-U-G-I-N","label":"testPlugin","interfaces":["reporter"],"api_version":"1"}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "another-testPlugin",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"another-testPlugin","label":"testPlugin","interfaces":["reporter"],"api_version":"1"}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "testPlugin!",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"testPlugin!","label":"testPlugin","interfaces":["reporter"],"api_version":"1"}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "test~plugin",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"test~plugin","label":"testPlugin","interfaces":["reporter"],"api_version":"1"}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "testPlugin-",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"testPlugin-","label":"testPlugin","interfaces":["reporter"],"api_version":"1"}]}`),
		}.file(),
		mockPlugin{
			t:       t,
			Name:    "-testPlugin",
			Handler: stringHandler(http.StatusOK, `{"Plugins":[{"id":"-testPlugin","label":"testPlugin","interfaces":["reporter"],"api_version":"1"}]}`),
		}.file(),
	)
	defer restore(t)

	r := testRegistry(t, "1")
	defer r.Close()

	r.Report()
	checkLoadedPluginIDs(t, r.ForEach, []string{"P-L-U-G-I-N", "another-testPlugin", "testPlugin"})
}

func checkControls(t *testing.T, topology report.Topology, expectedControls, expectedNodeControls []string, nodeID string) {
	controlsSet := report.MakeStringSet(expectedControls...)
	for _, id := range controlsSet {
		control, found := topology.Controls[id]
		if !found {
			t.Fatalf("Could not find an expected control %s in topology %s", id, topology.Label)
		}
		if control.ID != id {
			t.Fatalf("Control ID mismatch, expected %s, got %s", id, control.ID)
		}
	}
	if len(controlsSet) != len(topology.Controls) {
		t.Fatalf("Expected exactly %d controls in topology, got %d", len(controlsSet), len(topology.Controls))
	}

	node, found := topology.Nodes[nodeID]
	if !found {
		t.Fatalf("expected a node %s in a topology", nodeID)
	}
	actualNodeControls := []string{}
	node.LatestControls.ForEach(func(controlID string, _ time.Time, _ report.NodeControlData) {
		actualNodeControls = append(actualNodeControls, controlID)
	})
	nodeControlsSet := report.MakeStringSet(expectedNodeControls...)
	actualNodeControlsSet := report.MakeStringSet(actualNodeControls...)
	if !reflect.DeepEqual(nodeControlsSet, actualNodeControlsSet) {
		t.Fatalf("node controls in node %s in topology %s are not equal:\n%s", nodeID, topology.Label, test.Diff(nodeControlsSet, actualNodeControlsSet))
	}
}

func control(index int) (string, string) {
	return fmt.Sprintf("ctrl%d", index), fmt.Sprintf("Ctrl %d", index)
}

func controlID(index int) string {
	ID, _ := control(index)
	return ID
}

func mustMarshal(value interface{}) string {
	buf := &bytes.Buffer{}
	codec.NewEncoder(buf, &codec.JsonHandle{}).MustEncode(value)
	return buf.String()
}

func mustUnmarshal(r io.Reader, value interface{}) {
	codec.NewDecoder(r, &codec.JsonHandle{}).MustDecode(value)
}

func topologyControls(indices []int) report.Controls {
	var controls []report.Control
	for _, index := range indices {
		ID, name := control(index)
		controls = append(controls, report.Control{
			ID:    ID,
			Human: name,
			Icon:  "fa-at",
			Rank:  index,
		})
	}
	rptControls := report.Controls{}
	rptControls.AddControls(controls)
	return rptControls
}

func nodeControls(indices []int) []string {
	var IDs []string
	for _, index := range indices {
		ID, _ := control(index)
		IDs = append(IDs, ID)
	}
	return IDs
}

func topologyWithControls(label, nodeID string, controlIndices, nodeControlIndices []int) report.Topology {
	topology := report.MakeTopology().WithLabel(label, "")
	topology.Controls = topologyControls(controlIndices)
	return topology.AddNode(report.MakeNode(nodeID).WithLatestActiveControls(nodeControls(nodeControlIndices)...))
}

func pluginSpec(ID string, interfaces ...string) xfer.PluginSpec {
	return xfer.PluginSpec{
		ID:         ID,
		Label:      ID,
		Interfaces: interfaces,
		APIVersion: "1",
	}
}

func testReport(topology report.Topology, spec xfer.PluginSpec) report.Report {
	rpt := report.MakeReport()
	set := false
	f := func(t *report.Topology) {
		if t.Label != topology.Label {
			return
		}
		if set {
			panic("Two topologies with the same label")
		}
		set = true
		*t = t.Merge(topology)
	}
	rpt.WalkTopologies(f)
	if !set {
		panic(fmt.Sprintf("%s name is not a valid topology label", topology.Label))
	}
	rpt.Plugins = xfer.MakePluginSpecs(spec)
	return rpt
}

func TestRegistryRewritesControlReports(t *testing.T) {
	setup(
		t,
		mockPlugin{
			t:    t,
			Name: "testPlugin",
			Handler: mapStringHandler(testResponseMap{
				"/report":  {http.StatusOK, mustMarshal(testReport(topologyWithControls("pod", "node1", []int{1}, []int{1, 2}), pluginSpec("testPlugin", "reporter", "controller")))},
				"/control": {http.StatusOK, mustMarshal(PluginResponse{})},
			}),
		}.file(),
		mockPlugin{
			t:    t,
			Name: "testPluginReporterOnly",
			Handler: mapStringHandler(testResponseMap{
				"/report": {http.StatusOK, mustMarshal(testReport(topologyWithControls("host", "node1", []int{1}, []int{1, 2}), pluginSpec("testPluginReporterOnly", "reporter")))},
			}),
		}.file(),
	)
	defer restore(t)

	r := testRegistry(t, "1")
	defer r.Close()

	rpt, err := r.Report()
	if err != nil {
		t.Fatal(err)
	}
	// in a Pod topology, ctrl1 should be faked, ctrl2 should be left intact
	expectedPodControls := []string{fakeControlID("testPlugin", controlID(1))}
	expectedPodNodeControls := []string{fakeControlID("testPlugin", controlID(1)), controlID(2)}
	checkControls(t, rpt.Pod, expectedPodControls, expectedPodNodeControls, "node1")
	// in a Host topology, controls should be kept untouched
	expectedHostControls := []string{controlID(1)}
	expectedHostNodeControls := []string{controlID(1), controlID(2)}
	checkControls(t, rpt.Host, expectedHostControls, expectedHostNodeControls, "node1")
}

func TestRegistryRegistersHandlers(t *testing.T) {
	setup(
		t,
		mockPlugin{
			t:    t,
			Name: "testPlugin",
			Handler: mapStringHandler(testResponseMap{
				"/report":  {http.StatusOK, mustMarshal(testReport(topologyWithControls("pod", "node1", []int{1}, []int{1, 2}), pluginSpec("testPlugin", "reporter", "controller")))},
				"/control": {http.StatusOK, mustMarshal(PluginResponse{})},
			}),
		}.file(),
		mockPlugin{
			t:    t,
			Name: "testPlugin2",
			Handler: mapStringHandler(testResponseMap{
				"/report":  {http.StatusOK, mustMarshal(testReport(topologyWithControls("pod", "node2", []int{1, 2}, []int{1}), pluginSpec("testPlugin2", "reporter", "controller")))},
				"/control": {http.StatusOK, mustMarshal(PluginResponse{})},
			}),
		}.file(),
	)
	defer restore(t)

	testBackend := newTestHandlerRegistryBackend(t)
	handlerRegistry := controls.NewHandlerRegistry(testBackend)
	root := "/plugins"
	r, err := NewRegistry(root, "1", nil, handlerRegistry, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	expectedLen := 3
	if len(testBackend.handlers) != expectedLen {
		t.Fatalf("Expected %d registered handler, got %d", expectedLen, len(testBackend.handlers))
	}
	fakeIDs := []string{
		fakeControlID("testPlugin", controlID(1)),
		fakeControlID("testPlugin2", controlID(1)),
		fakeControlID("testPlugin2", controlID(2)),
	}
	for _, fakeID := range fakeIDs {
		if _, found := testBackend.Handler(fakeID); !found {
			t.Fatalf("Expected to have a handler for %s", fakeID)
		}
	}
}

func TestRegistryHandlersCallPlugins(t *testing.T) {
	setup(
		t,
		mockPlugin{
			t:    t,
			Name: "testPlugin",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/report":
					w.WriteHeader(http.StatusOK)
					rpt := mustMarshal(testReport(topologyWithControls("pod", "node1", []int{1}, []int{1}), pluginSpec("testPlugin", "reporter", "controller")))
					fmt.Fprint(w, rpt)
				case "/control":
					xreq := xfer.Request{}
					mustUnmarshal(r.Body, &xreq)
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, mustMarshal(PluginResponse{Response: xfer.Response{Value: fmt.Sprintf("%s,%s", xreq.NodeID, xreq.Control)}}))
				default:
					http.NotFound(w, r)
				}
			}),
		}.file(),
	)
	defer restore(t)

	handlerRegistry := controls.NewDefaultHandlerRegistry()
	root := "/plugins"
	r, err := NewRegistry(root, "1", nil, handlerRegistry, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.Report()
	fakeID := fakeControlID("testPlugin", controlID(1))
	req := xfer.Request{NodeID: "node1", Control: fakeID}
	res := handlerRegistry.HandleControlRequest(req)
	if res.Value != fmt.Sprintf("node1,%s", controlID(1)) {
		t.Fatalf("Got unexpected response: %#v", res)
	}
}
