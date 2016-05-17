package main

import (
	"flag"
	"fmt"
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
	wscat wscatFlags
}

type probeFlags struct {
	token           string
	httpListen      string
	publishInterval time.Duration
	spyInterval     time.Duration
	spyProcs        bool
	procRoot        string
	pluginsRoot     string
	useConntrack    bool
	insecure        bool
	logPrefix       string
	logLevel        string
	resolver        string
	noApp           bool

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

	collectorURL     string
	controlRouterURL string
	pipeRouterURL    string
	userIDHeader     string

	awsCreateTables bool
	consulInf       string
}

type wscatFlags struct {
	blockOnEOF bool
}

func main() {
	var (
		flags         = flags{}
		mode          string
		debug         bool
		weaveHostname string
	)

	// Flags that apply to both probe and app
	flag.StringVar(&mode, "mode", "help", "For internal use.")
	flag.BoolVar(&debug, "debug", false, "Force debug logging.")
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
	flag.BoolVar(&flags.probe.spyProcs, "probe.processes", true, "report processes (needs root)")
	flag.StringVar(&flags.probe.procRoot, "probe.proc.root", "/proc", "location of the proc filesystem")
	flag.StringVar(&flags.probe.pluginsRoot, "probe.plugins.root", "/var/run/scope/plugins", "Root directory to search for plugins")
	flag.BoolVar(&flags.probe.useConntrack, "probe.conntrack", true, "also use conntrack to track connections")
	flag.BoolVar(&flags.probe.insecure, "probe.insecure", false, "(SSL) explicitly allow \"insecure\" SSL connections and transfers")
	flag.StringVar(&flags.probe.resolver, "probe.resolver", "", "IP address & port of resolver to use.  Default is to use system resolver.")
	flag.StringVar(&flags.probe.logPrefix, "probe.log.prefix", "<probe>", "prefix for each log line")
	flag.StringVar(&flags.probe.logLevel, "probe.log.level", "info", "logging threshold level: debug|info|warn|error|fatal|panic")
	flag.BoolVar(&flags.probe.dockerEnabled, "probe.docker", false, "collect Docker-related attributes for processes")
	flag.DurationVar(&flags.probe.dockerInterval, "probe.docker.interval", 10*time.Second, "how often to update Docker attributes")
	flag.StringVar(&flags.probe.dockerBridge, "probe.docker.bridge", "docker0", "the docker bridge name")
	flag.BoolVar(&flags.probe.kubernetesEnabled, "probe.kubernetes", false, "collect kubernetes-related attributes for containers, should only be enabled on the master node")
	flag.StringVar(&flags.probe.kubernetesAPI, "probe.kubernetes.api", "", "Address of kubernetes master api")
	flag.DurationVar(&flags.probe.kubernetesInterval, "probe.kubernetes.interval", 10*time.Second, "how often to do a full resync of the kubernetes data")
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
	flag.StringVar(&flags.app.controlRouterURL, "app.control.router", "local", "Control router to use (local or sqs)")
	flag.StringVar(&flags.app.pipeRouterURL, "app.pipe.router", "local", "Pipe router to use (local)")
	flag.StringVar(&flags.app.userIDHeader, "app.userid.header", "", "HTTP header to use as userid")

	flag.BoolVar(&flags.app.awsCreateTables, "app.aws.create.tables", false, "Create the tables in DynamoDB")
	flag.StringVar(&flags.app.consulInf, "app.consul.inf", "", "The interface who's address I should advertise myself under in consul")

	flag.BoolVar(&flags.wscat.blockOnEOF, "wscat.block", false, "Block on end of line (do not exit early of EOF is received)")

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

	switch mode {
	case "app":
		appMain(flags.app)
	case "probe":
		probeMain(flags.probe)
	case "version":
		fmt.Println("Weave Scope version", version)
	case "help":
		flag.PrintDefaults()
	case "wscat":
		wscat(flags.wscat)
	default:
		fmt.Printf("command '%s' not recognices", mode)
		os.Exit(1)
	}
}
