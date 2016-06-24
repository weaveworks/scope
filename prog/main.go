package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/weave/common"
)

var version = "dev" // set at build time

type prefixFormatter struct {
	prefix []byte
	next   log.Formatter
}

func (f *prefixFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.next.Format(entry)
	if err != nil {
		return formatted, err
	}
	return append(f.prefix, formatted...), nil
}

func setLogFormatter(prefix string) {
	if !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}
	f := prefixFormatter{
		prefix: []byte(prefix),
		// reuse weave's log format
		next: common.Log.Formatter,
	}
	log.SetFormatter(&f)
}

func setLogLevel(levelname string) {
	level, err := log.ParseLevel(levelname)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(level)
}

type flags struct {
	probe probeFlags
	app   appFlags
}

type probeFlags struct {
	token           string
	httpListen      string
	publishInterval time.Duration
	spyInterval     time.Duration
	pluginsRoot     string
	insecure        bool
	logPrefix       string
	logLevel        string
	resolver        string
	noApp           bool

	useConntrack bool // Use conntrack for endpoint topo
	spyProcs     bool // Associate endpoints with processes (must be root)
	procEnabled  bool // Produce process topology & process nodes in endpoint
	procRoot     string

	dockerEnabled  bool
	dockerInterval time.Duration
	dockerBridge   string

	kubernetesEnabled  bool
	kubernetesAPI      string
	kubernetesInterval time.Duration

	weaveAddr     string
	weaveHostname string
}

type appFlags struct {
	window    time.Duration
	listen    string
	logLevel  string
	logPrefix string
	logHTTP   bool

	weaveAddr      string
	weaveHostname  string
	containerName  string
	dockerEndpoint string

	collectorURL      string
	s3URL             string
	controlRouterURL  string
	pipeRouterURL     string
	natsHostname      string
	memcachedHostname string
	memcachedTimeout  time.Duration
	memcachedService  string
	userIDHeader      string

	awsCreateTables bool
	consulInf       string
}

func main() {
	var (
		flags         = flags{}
		mode          string
		debug         bool
		weaveHostname string
		dryRun        bool
	)

	// Flags that apply to both probe and app
	flag.StringVar(&mode, "mode", "help", "For internal use.")
	flag.BoolVar(&debug, "debug", false, "Force debug logging.")
	flag.BoolVar(&dryRun, "dry-run", false, "Don't start scope, just parse the arguments.  For internal use only.")
	flag.StringVar(&weaveHostname, "weave.hostname", "", "Hostname to advertise/lookup in WeaveDNS")

	// We need to know how to parse them, but they are mainly interpreted by the entrypoint script.
	// They are also here so they are included in usage, and the probe uses them to decide if to
	// publish to localhost.
	noApp := flag.Bool("no-app", false, "Don't run the app.")
	probeOnly := flag.Bool("app-only", false, "Only run the app")
	flag.Bool("probe-only", false, "Only run the probe.")
	flag.Bool("no-probe", false, "Don't run the probe.")

	// Probe flags
	flag.StringVar(&flags.probe.token, "service-token", "", "Token to use to authenticate with scope.weave.works")
	flag.StringVar(&flags.probe.token, "probe.token", "", "Token to use to authenticate with scope.weave.works")
	flag.StringVar(&flags.probe.httpListen, "probe.http.listen", "", "listen address for HTTP profiling and instrumentation server")
	flag.DurationVar(&flags.probe.publishInterval, "probe.publish.interval", 3*time.Second, "publish (output) interval")
	flag.DurationVar(&flags.probe.spyInterval, "probe.spy.interval", time.Second, "spy (scan) interval")
	flag.StringVar(&flags.probe.pluginsRoot, "probe.plugins.root", "/var/run/scope/plugins", "Root directory to search for plugins")

	flag.BoolVar(&flags.probe.insecure, "probe.insecure", false, "(SSL) explicitly allow \"insecure\" SSL connections and transfers")
	flag.StringVar(&flags.probe.resolver, "probe.resolver", "", "IP address & port of resolver to use.  Default is to use system resolver.")
	flag.StringVar(&flags.probe.logPrefix, "probe.log.prefix", "<probe>", "prefix for each log line")
	flag.StringVar(&flags.probe.logLevel, "probe.log.level", "info", "logging threshold level: debug|info|warn|error|fatal|panic")

	// Proc & endpoint
	flag.BoolVar(&flags.probe.useConntrack, "probe.conntrack", true, "also use conntrack to track connections")
	flag.BoolVar(&flags.probe.spyProcs, "probe.proc.spy", true, "associate endpoints with processes (needs root)")
	flag.StringVar(&flags.probe.procRoot, "probe.proc.root", "/proc", "location of the proc filesystem")
	flag.BoolVar(&flags.probe.procEnabled, "probe.processes", true, "produce process topology & include procspied connections")

	// Docker
	flag.BoolVar(&flags.probe.dockerEnabled, "probe.docker", false, "collect Docker-related attributes for processes")
	flag.DurationVar(&flags.probe.dockerInterval, "probe.docker.interval", 10*time.Second, "how often to update Docker attributes")
	flag.StringVar(&flags.probe.dockerBridge, "probe.docker.bridge", "docker0", "the docker bridge name")

	// K8s
	flag.BoolVar(&flags.probe.kubernetesEnabled, "probe.kubernetes", false, "collect kubernetes-related attributes for containers, should only be enabled on the master node")
	flag.StringVar(&flags.probe.kubernetesAPI, "probe.kubernetes.api", "", "Address of kubernetes master api")
	flag.DurationVar(&flags.probe.kubernetesInterval, "probe.kubernetes.interval", 10*time.Second, "how often to do a full resync of the kubernetes data")

	// Weave
	flag.StringVar(&flags.probe.weaveAddr, "probe.weave.addr", "127.0.0.1:6784", "IP address & port of the Weave router")
	flag.StringVar(&flags.probe.weaveHostname, "probe.weave.hostname", app.DefaultHostname, "Hostname to lookup in WeaveDNS")

	// App flags
	flag.DurationVar(&flags.app.window, "app.window", 15*time.Second, "window")
	flag.StringVar(&flags.app.listen, "app.http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
	flag.StringVar(&flags.app.logLevel, "app.log.level", "info", "logging threshold level: debug|info|warn|error|fatal|panic")
	flag.StringVar(&flags.app.logPrefix, "app.log.prefix", "<app>", "prefix for each log line")
	flag.BoolVar(&flags.app.logHTTP, "app.log.http", false, "Log individual HTTP requests")

	flag.StringVar(&flags.app.weaveAddr, "app.weave.addr", app.DefaultWeaveURL, "Address on which to contact WeaveDNS")
	flag.StringVar(&flags.app.weaveHostname, "app.weave.hostname", app.DefaultHostname, "Hostname to advertise in WeaveDNS")
	flag.StringVar(&flags.app.containerName, "app.container.name", app.DefaultContainerName, "Name of this container (to lookup container ID)")
	flag.StringVar(&flags.app.dockerEndpoint, "app.docker", app.DefaultDockerEndpoint, "Location of docker endpoint (to lookup container ID)")

	flag.StringVar(&flags.app.collectorURL, "app.collector", "local", "Collector to use (local of dynamodb)")
	flag.StringVar(&flags.app.s3URL, "app.collector.s3", "local", "S3 URL to use (when collector is dynamodb)")
	flag.StringVar(&flags.app.controlRouterURL, "app.control.router", "local", "Control router to use (local or sqs)")
	flag.StringVar(&flags.app.pipeRouterURL, "app.pipe.router", "local", "Pipe router to use (local)")
	flag.StringVar(&flags.app.natsHostname, "app.nats", "", "Hostname for NATS service to use for shortcut reports.  If empty, shortcut reporting will be disabled.")
	flag.StringVar(&flags.app.memcachedHostname, "app.memcached.hostname", "", "Hostname for memcached service to use when caching reports.  If empty, no memcached will be used.")
	flag.DurationVar(&flags.app.memcachedTimeout, "app.memcached.timeout", 100*time.Millisecond, "Maximum time to wait before giving up on memcached requests.")
	flag.StringVar(&flags.app.memcachedService, "app.memcached.service", "memcached", "SRV service used to discover memcache servers.")
	flag.StringVar(&flags.app.userIDHeader, "app.userid.header", "", "HTTP header to use as userid")

	flag.BoolVar(&flags.app.awsCreateTables, "app.aws.create.tables", false, "Create the tables in DynamoDB")
	flag.StringVar(&flags.app.consulInf, "app.consul.inf", "", "The interface who's address I should advertise myself under in consul")

	flag.Parse()

	// Deal with common args
	if debug {
		flags.probe.logLevel = "debug"
		flags.app.logLevel = "debug"
	}
	if weaveHostname != "" {
		flags.probe.weaveHostname = weaveHostname
		flags.app.weaveHostname = weaveHostname
	}
	flags.probe.noApp = *noApp || *probeOnly

	// Special case for #1191, check listen address is well formed
	_, _, err := net.SplitHostPort(flags.app.listen)
	if err != nil {
		log.Errorf("Invalid value for -app.http.address: %v", err)
	}
	if flags.probe.httpListen != "" {
		_, _, err := net.SplitHostPort(flags.probe.httpListen)
		if err != nil {
			log.Errorf("Invalid value for -app.http.address: %v", err)
		}
	}

	if dryRun {
		return
	}

	switch mode {
	case "app":
		appMain(flags.app)
	case "probe":
		probeMain(flags.probe)
	case "version":
		fmt.Println("Weave Scope version", version)
	case "help":
		flag.PrintDefaults()
	default:
		fmt.Printf("command '%s' not recognices", mode)
		os.Exit(1)
	}
}
