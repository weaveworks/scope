package plugins

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/weaveworks/scope/common/exec"
	"github.com/weaveworks/scope/common/fs"
	"github.com/weaveworks/scope/report"
)

const pluginPollDuration = 1 * time.Minute

// PluginRegistry keeps a list of plugins and periodically refreshes it.
type PluginRegistry struct {
	sync.Mutex
	pluginPath string
	plugins    []plugin
	quit       chan struct{}
}

// NewPluginRegistry makes a new plugin registry.
func NewPluginRegistry(pluginPath string) *PluginRegistry {
	result := &PluginRegistry{
		pluginPath: pluginPath,
		quit:       make(chan struct{}),
	}
	go result.loop()
	return result
}

// Stop the plugin registry.
func (r *PluginRegistry) Stop() {
	close(r.quit)
}

func (r *PluginRegistry) loop() {
	ticker := time.Tick(pluginPollDuration)
	for {
		r.scanPlugins()
		select {
		case <-ticker:
		case <-r.quit:
			return
		}
	}
}

func (r *PluginRegistry) scanPlugins() {
	files, err := fs.ReadDir(r.pluginPath)
	if err != nil {
		log.Println("Error listing plugin dir:", err)
		return
	}

	plugins := []plugin{}
	for _, file := range files {
		// check it a regular, executable file
		mode := file.Mode()
		if mode&os.ModeType != 0 {
			log.Println("Not regular file, skipping:", file.Name())
		}
		if mode.Perm()&0111 == 0 {
			log.Println("Not executable, skipping:", file.Name())
			continue
		}
		plugins = append(plugins, plugin{
			path: path.Join(r.pluginPath, file.Name()),
		})
	}

	r.Lock()
	defer r.Unlock()
	r.plugins = plugins
}

// Name implements Tagger
func (*PluginRegistry) Name() string {
	return "Plugins"
}

func (r *PluginRegistry) withAll(f func(plugin) report.Report) report.Report {
	r.Lock()
	defer r.Unlock()
	reports := make(chan report.Report, len(r.plugins))
	wg := sync.WaitGroup{}
	wg.Add(len(r.plugins))
	for _, p := range r.plugins {
		go func(p plugin) {
			reports <- f(p)
			wg.Done()
		}(p)
	}
	wg.Wait()
	close(reports)
	result := report.MakeReport()
	for out := range reports {
		result = result.Merge(out)
	}
	return result
}

// Tag implements Tagger
func (r *PluginRegistry) Tag(rep report.Report) (report.Report, error) {
	result := r.withAll(func(p plugin) report.Report {
		out, err := p.tag(rep)
		if err != nil {
			log.Printf("Plugin %s error tagging report: %v", p.path, err)
			return rep
		}
		return out
	})
	return result, nil
}

// Report implements Reporter
func (r *PluginRegistry) Report() (report.Report, error) {
	result := r.withAll(func(p plugin) report.Report {
		out, err := p.report()
		if err != nil {
			log.Printf("Plugin %s error tagging report: %v", p.path, err)
			return report.MakeReport()
		}
		return out
	})
	return result, nil
}

type plugin struct {
	path string
}

func (p plugin) doRequest(req *http.Request) (*http.Response, error) {
	cmd := exec.Command(p.path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer cmd.Wait()

	// Write a HTTP request to the exec'd process' stdin.
	req.Write(stdin)
	stdin.Close()

	// Read a HTTRP response from the exec'd process' stdout.
	return http.ReadResponse(bufio.NewReader(stdout), req)
}

func (p plugin) tag(r report.Report) (report.Report, error) {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(r); err != nil {
		return report.MakeReport(), err
	}
	req, err := http.NewRequest("POST", "/tag", &buf)
	if err != nil {
		return report.MakeReport(), err
	}
	resp, err := p.doRequest(req)
	if err != nil {
		return report.MakeReport(), err
	}
	result := report.MakeReport()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return report.MakeReport(), err
	}
	return result, nil
}

func (p plugin) report() (report.Report, error) {
	req, err := http.NewRequest("GET", "/report", nil)
	if err != nil {
		return report.MakeReport(), err
	}
	resp, err := p.doRequest(req)
	if err != nil {
		return report.MakeReport(), err
	}
	result := report.MakeReport()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return report.MakeReport(), err
	}
	return result, nil
}
