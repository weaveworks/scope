package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/appclient"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/weave/common"
)

var (
	// set at build time
	version = "dev"
	// tokens to be elided when logging
	serviceTokenFlag       = "service-token"
	probeTokenFlag         = "probe.token"
	kubernetesPasswordFlag = "probe.kubernetes.password"
	kubernetesTokenFlag    = "probe.kubernetes.token"
	sensitiveFlags         = []string{
		serviceTokenFlag,
		probeTokenFlag,
		kubernetesPasswordFlag,
		kubernetesTokenFlag,
	}
	colonFinder         = regexp.MustCompile(`[^\\](:)`)
	unescapeBackslashes = regexp.MustCompile(`\\(.)`)
)

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
	noControls      bool

	useConntrack        bool // Use conntrack for endpoint topo
	conntrackBufferSize int  // Sie of kernel buffer for conntrack

	spyProcs    bool // Associate endpoints with processes (must be root)
	procEnabled bool // Produce process topology & process nodes in endpoint
	procRoot    string

	dockerEnabled  bool
	dockerInterval time.Duration
	dockerBridge   string

	kubernetesEnabled bool
	kubernetesConfig  kubernetes.ClientConfig

	ecsEnabled     bool
	ecsCacheSize   int
	ecsCacheExpiry time.Duration

	weaveEnabled  bool
	weaveAddr     string
	weaveHostname string
}

type appFlags struct {
	window         time.Duration
	listen         string
	stopTimeout    time.Duration
	logLevel       string
	logPrefix      string
	logHTTP        bool
	logHTTPHeaders bool

	weaveEnabled   bool
	weaveAddr      string
	weaveHostname  string
	containerName  string
	dockerEndpoint string

	collectorURL              string
	s3URL                     string
	controlRouterURL          string
	pipeRouterURL             string
	natsHostname              string
	memcachedHostname         string
	memcachedTimeout          time.Duration
	memcachedService          string
	memcachedExpiration       time.Duration
	memcachedCompressionLevel int
	userIDHeader              string
	externalUI                bool

	blockProfileRate int

	awsCreateTables bool
	consulInf       string
}

type containerLabelFiltersFlag struct {
	apiTopologyOptions []app.APITopologyOption
	filterNumber       int
	filterIDPrefix     string
	exclude            bool
}

func (c *containerLabelFiltersFlag) String() string {
	return fmt.Sprint(c.apiTopologyOptions)
}

func (c *containerLabelFiltersFlag) Set(flagValue string) error {
	filterID := fmt.Sprintf(c.filterIDPrefix+"%d", c.filterNumber)
	newAPITopologyOption, err := c.toAPITopologyOption(flagValue, filterID)
	if err != nil {
		return err
	}
	c.filterNumber++

	c.apiTopologyOptions = append(c.apiTopologyOptions, newAPITopologyOption)
	return nil
}

func (c *containerLabelFiltersFlag) toAPITopologyOption(flagValue string, filterID string) (app.APITopologyOption, error) {
	indexRanges := colonFinder.FindAllStringIndex(flagValue, -1)
	if len(indexRanges) != 1 {
		if len(indexRanges) == 0 {
			return app.APITopologyOption{}, fmt.Errorf("No unescaped colon found. This is needed to separate the title from the label")
		}
		return app.APITopologyOption{}, fmt.Errorf("Multiple unescaped colons. Escape colons that are part of the title and label")
	}
	splitIndices := indexRanges[0]
	titleStringEscaped := flagValue[:splitIndices[0]+1]
	labelStringEscaped := flagValue[splitIndices[1]:]
	containerFilterTitle := unescapeBackslashes.ReplaceAllString(titleStringEscaped, `$1`)
	containerFilterLabel := unescapeBackslashes.ReplaceAllString(labelStringEscaped, `$1`)
	labelKeyValuePair := strings.Split(containerFilterLabel, "=")
	if len(labelKeyValuePair) != 2 {
		return app.APITopologyOption{}, fmt.Errorf("Docker label isn't in the correct key=value format")
	}

	filterFunction := render.HasLabel
	if c.exclude {
		filterFunction = render.DoesNotHaveLabel
	}
	return app.MakeAPITopologyOption(filterID, containerFilterTitle, filterFunction(labelKeyValuePair[0], labelKeyValuePair[1]), false), nil
}

func logCensoredArgs() {
	var prettyPrintedArgs string
	// We show the flags followed by the args. This may change the original
	// ordering. However the flag parser doesn't keep positioning
	// information, not allowing for a more accurate reconstruction.
	flag.Visit(func(f *flag.Flag) {
		value := f.Value.String()
		// omit sensitive information
		for _, sensitiveFlag := range sensitiveFlags {
			if f.Name == sensitiveFlag {
				value = "<elided>"
				break
			}
		}
		prettyPrintedArgs += fmt.Sprintf(" --%s=%s", f.Name, value)
	})
	for _, arg := range flag.Args() {
		prettyPrintedArgs += " " + arg
	}
	log.Infof("command line args:%s", prettyPrintedArgs)
}

func makeBaseCheckpointFlags() map[string]string {
	release, _, err := host.GetKernelReleaseAndVersion()
	if err != nil {
		release = "unknown"
	}
	return map[string]string{
		// Inconsistent key (using a dash) to match Weave Net
		"kernel-version": release,
		"os":             runtime.GOOS,
	}
}

func main() {
	var (
		flags                            = flags{}
		mode                             string
		debug                            bool
		weaveEnabled                     bool
		weaveHostname                    string
		dryRun                           bool
		containerLabelFilterFlags        = containerLabelFiltersFlag{exclude: false, filterIDPrefix: "containerLabelFilterExclude"}
		containerLabelFilterFlagsExclude = containerLabelFiltersFlag{exclude: true, filterIDPrefix: "containerLabelFilter"}
	)

	// Flags that apply to both probe and app
	flag.StringVar(&mode, "mode", "help", "For internal use.")
	flag.BoolVar(&debug, "debug", false, "Force debug logging.")
	flag.BoolVar(&dryRun, "dry-run", false, "Don't start scope, just parse the arguments.  For internal use only.")
	flag.BoolVar(&weaveEnabled, "weave", true, "Enable Weave Net integrations.")
	flag.StringVar(&weaveHostname, "weave.hostname", "", "Hostname to advertise/lookup in WeaveDNS")

	// We need to know how to parse them, but they are mainly interpreted by the entrypoint script.
	// They are also here so they are included in usage, and the probe uses them to decide if to
	// publish to localhost.
	noApp := flag.Bool("no-app", false, "Don't run the app.")
	probeOnly := flag.Bool("app-only", false, "Only run the app")
	flag.Bool("probe-only", false, "Only run the probe.")
	flag.Bool("no-probe", false, "Don't run the probe.")

	// Probe flags
	flag.StringVar(&flags.probe.token, serviceTokenFlag, "", "Token to use to authenticate with cloud.weave.works")
	flag.StringVar(&flags.probe.token, probeTokenFlag, "", "Token to use to authenticate with cloud.weave.works")
	flag.StringVar(&flags.probe.httpListen, "probe.http.listen", "", "listen address for HTTP profiling and instrumentation server")
	flag.DurationVar(&flags.probe.publishInterval, "probe.publish.interval", 3*time.Second, "publish (output) interval")
	flag.DurationVar(&flags.probe.spyInterval, "probe.spy.interval", time.Second, "spy (scan) interval")
	flag.StringVar(&flags.probe.pluginsRoot, "probe.plugins.root", "/var/run/scope/plugins", "Root directory to search for plugins")
	flag.BoolVar(&flags.probe.noControls, "probe.no-controls", false, "Disable controls (e.g. start/stop containers, terminals, logs ...)")

	flag.BoolVar(&flags.probe.insecure, "probe.insecure", false, "(SSL) explicitly allow \"insecure\" SSL connections and transfers")
	flag.StringVar(&flags.probe.resolver, "probe.resolver", "", "IP address & port of resolver to use.  Default is to use system resolver.")
	flag.StringVar(&flags.probe.logPrefix, "probe.log.prefix", "<probe>", "prefix for each log line")
	flag.StringVar(&flags.probe.logLevel, "probe.log.level", "info", "logging threshold level: debug|info|warn|error|fatal|panic")

	// Proc & endpoint
	flag.BoolVar(&flags.probe.useConntrack, "probe.conntrack", true, "also use conntrack to track connections")
	flag.IntVar(&flags.probe.conntrackBufferSize, "probe.conntrack.buffersize", 208*1024, "conntrack buffer size")
	flag.BoolVar(&flags.probe.spyProcs, "probe.proc.spy", true, "associate endpoints with processes (needs root)")
	flag.StringVar(&flags.probe.procRoot, "probe.proc.root", "/proc", "location of the proc filesystem")
	flag.BoolVar(&flags.probe.procEnabled, "probe.processes", true, "produce process topology & include procspied connections")

	// Docker
	flag.BoolVar(&flags.probe.dockerEnabled, "probe.docker", false, "collect Docker-related attributes for processes")
	flag.DurationVar(&flags.probe.dockerInterval, "probe.docker.interval", 10*time.Second, "how often to update Docker attributes")
	flag.StringVar(&flags.probe.dockerBridge, "probe.docker.bridge", "docker0", "the docker bridge name")

	// K8s
	flag.BoolVar(&flags.probe.kubernetesEnabled, "probe.kubernetes", false, "collect kubernetes-related attributes for containers, should only be enabled on the master node")
	flag.DurationVar(&flags.probe.kubernetesConfig.Interval, "probe.kubernetes.interval", 10*time.Second, "how often to do a full resync of the kubernetes data")
	flag.StringVar(&flags.probe.kubernetesConfig.Server, "probe.kubernetes.api", "", "The address and port of the Kubernetes API server (deprecated in favor of equivalent probe.kubernetes.server)")
	flag.StringVar(&flags.probe.kubernetesConfig.CertificateAuthority, "probe.kubernetes.certificate-authority", "", "Path to a cert. file for the certificate authority")
	flag.StringVar(&flags.probe.kubernetesConfig.ClientCertificate, "probe.kubernetes.client-certificate", "", "Path to a client certificate file for TLS")
	flag.StringVar(&flags.probe.kubernetesConfig.ClientKey, "probe.kubernetes.client-key", "", "Path to a client key file for TLS")
	flag.StringVar(&flags.probe.kubernetesConfig.Cluster, "probe.kubernetes.cluster", "", "The name of the kubeconfig cluster to use")
	flag.StringVar(&flags.probe.kubernetesConfig.Context, "probe.kubernetes.context", "", "The name of the kubeconfig context to use")
	flag.BoolVar(&flags.probe.kubernetesConfig.Insecure, "probe.kubernetes.insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	flag.StringVar(&flags.probe.kubernetesConfig.Kubeconfig, "probe.kubernetes.kubeconfig", "", "Path to the kubeconfig file to use")
	flag.StringVar(&flags.probe.kubernetesConfig.Password, kubernetesPasswordFlag, "", "Password for basic authentication to the API server")
	flag.StringVar(&flags.probe.kubernetesConfig.Server, "probe.kubernetes.server", "", "The address and port of the Kubernetes API server")
	flag.StringVar(&flags.probe.kubernetesConfig.Token, kubernetesTokenFlag, "", "Bearer token for authentication to the API server")
	flag.StringVar(&flags.probe.kubernetesConfig.User, "probe.kubernetes.user", "", "The name of the kubeconfig user to use")
	flag.StringVar(&flags.probe.kubernetesConfig.Username, "probe.kubernetes.username", "", "Username for basic authentication to the API server")

	// AWS ECS
	flag.BoolVar(&flags.probe.ecsEnabled, "probe.ecs", false, "Collect ecs-related attributes for containers on this node")
	flag.IntVar(&flags.probe.ecsCacheSize, "probe.ecs.cache.size", 1024*1024, "Max size of cached info for each ECS cluster")
	flag.DurationVar(&flags.probe.ecsCacheExpiry, "probe.ecs.cache.expiry", time.Hour, "How long to keep cached ECS info")

	// Weave
	flag.StringVar(&flags.probe.weaveAddr, "probe.weave.addr", "127.0.0.1:6784", "IP address & port of the Weave router")
	flag.StringVar(&flags.probe.weaveHostname, "probe.weave.hostname", app.DefaultHostname, "Hostname to lookup in WeaveDNS")

	// App flags
	flag.DurationVar(&flags.app.window, "app.window", 15*time.Second, "window")
	flag.StringVar(&flags.app.listen, "app.http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
	flag.DurationVar(&flags.app.stopTimeout, "app.stopTimeout", 5*time.Second, "How long to wait for http requests to finish when shutting down")
	flag.StringVar(&flags.app.logLevel, "app.log.level", "info", "logging threshold level: debug|info|warn|error|fatal|panic")
	flag.StringVar(&flags.app.logPrefix, "app.log.prefix", "<app>", "prefix for each log line")
	flag.BoolVar(&flags.app.logHTTP, "app.log.http", false, "Log individual HTTP requests")
	flag.BoolVar(&flags.app.logHTTPHeaders, "app.log.httpHeaders", false, "Log HTTP headers. Needs app.log.http to be enabled.")

	flag.StringVar(&flags.app.weaveAddr, "app.weave.addr", app.DefaultWeaveURL, "Address on which to contact WeaveDNS")
	flag.StringVar(&flags.app.weaveHostname, "app.weave.hostname", app.DefaultHostname, "Hostname to advertise in WeaveDNS")
	flag.StringVar(&flags.app.containerName, "app.container.name", app.DefaultContainerName, "Name of this container (to lookup container ID)")
	flag.StringVar(&flags.app.dockerEndpoint, "app.docker", app.DefaultDockerEndpoint, "Location of docker endpoint (to lookup container ID)")
	flag.Var(&containerLabelFilterFlags, "app.container-label-filter", "Add container label-based view filter, specified as title:label. Multiple flags are accepted. Example: --app.container-label-filter='Database Containers:role=db'")
	flag.Var(&containerLabelFilterFlagsExclude, "app.container-label-filter-exclude", "Add container label-based view filter that excludes containers with the given label, specified as title:label. Multiple flags are accepted. Example: --app.container-label-filter-exclude='Database Containers:role=db'")

	flag.StringVar(&flags.app.collectorURL, "app.collector", "local", "Collector to use (local, dynamodb, or file)")
	flag.StringVar(&flags.app.s3URL, "app.collector.s3", "local", "S3 URL to use (when collector is dynamodb)")
	flag.StringVar(&flags.app.controlRouterURL, "app.control.router", "local", "Control router to use (local or sqs)")
	flag.StringVar(&flags.app.pipeRouterURL, "app.pipe.router", "local", "Pipe router to use (local)")
	flag.StringVar(&flags.app.natsHostname, "app.nats", "", "Hostname for NATS service to use for shortcut reports.  If empty, shortcut reporting will be disabled.")
	flag.StringVar(&flags.app.memcachedHostname, "app.memcached.hostname", "", "Hostname for memcached service to use when caching reports.  If empty, no memcached will be used.")
	flag.DurationVar(&flags.app.memcachedTimeout, "app.memcached.timeout", 100*time.Millisecond, "Maximum time to wait before giving up on memcached requests.")
	flag.DurationVar(&flags.app.memcachedExpiration, "app.memcached.expiration", 2*15*time.Second, "How long reports stay in the memcache.")
	flag.StringVar(&flags.app.memcachedService, "app.memcached.service", "memcached", "SRV service used to discover memcache servers.")
	flag.IntVar(&flags.app.memcachedCompressionLevel, "app.memcached.compression", gzip.DefaultCompression, "How much to compress reports stored in memcached.")
	flag.StringVar(&flags.app.userIDHeader, "app.userid.header", "", "HTTP header to use as userid")
	flag.BoolVar(&flags.app.externalUI, "app.externalUI", false, "Point to externally hosted static UI assets")

	flag.IntVar(&flags.app.blockProfileRate, "app.block.profile.rate", 0, "If more than 0, enable block profiling. The profiler aims to sample an average of one blocking event per rate nanoseconds spent blocked.")

	flag.BoolVar(&flags.app.awsCreateTables, "app.aws.create.tables", false, "Create the tables in DynamoDB")
	flag.StringVar(&flags.app.consulInf, "app.consul.inf", "", "The interface who's address I should advertise myself under in consul")

	flag.Parse()

	app.AddContainerFilters(append(containerLabelFilterFlags.apiTopologyOptions, containerLabelFilterFlagsExclude.apiTopologyOptions...)...)

	// Deal with common args
	if debug {
		flags.probe.logLevel = "debug"
		flags.app.logLevel = "debug"
	}
	if weaveHostname != "" {
		flags.probe.weaveHostname = weaveHostname
		flags.app.weaveHostname = weaveHostname
	}
	flags.probe.weaveEnabled = weaveEnabled
	flags.app.weaveEnabled = weaveEnabled
	flags.probe.noApp = *noApp || *probeOnly

	// Special case for #1191, check listen address is well formed
	_, port, err := net.SplitHostPort(flags.app.listen)
	if err != nil {
		log.Fatalf("Invalid value for -app.http.address: %v", err)
	}
	if flags.probe.httpListen != "" {
		_, _, err := net.SplitHostPort(flags.probe.httpListen)
		if err != nil {
			log.Fatalf("Invalid value for -app.http.address: %v", err)
		}
	}

	// Special case probe push address parsing
	targets := []appclient.Target{}
	if mode == "probe" || dryRun {
		args := []string{}
		if flags.probe.token != "" {
			// service mode
			if len(flag.Args()) == 0 {
				args = append(args, defaultServiceHost)
			}
		} else if !flags.probe.noApp {
			// We hardcode 127.0.0.1 instead of using localhost
			// since it leads to problems in exotic DNS setups
			args = append(args, fmt.Sprintf("127.0.0.1:%s", port))
		}
		args = append(args, flag.Args()...)
		if !dryRun {
			log.Infof("publishing to: %s", strings.Join(args, ", "))
		}
		targets, err = appclient.ParseTargets(args)
		if err != nil {
			log.Fatalf("Invalid targets: %v", err)
		}
	}

	if dryRun {
		return
	}

	switch mode {
	case "app":
		appMain(flags.app)
	case "probe":
		probeMain(flags.probe, targets)
	case "version":
		fmt.Println("Weave Scope version", version)
	case "help":
		flag.PrintDefaults()
	default:
		fmt.Printf("command '%s' not recognized", mode)
		os.Exit(1)
	}
}
