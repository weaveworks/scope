// Copyright 2012-2016 Apcera Inc. All rights reserved.

package server

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"time"

	// Allow dynamic profiling.
	_ "net/http/pprof"
)

// Info is the information sent to clients to help them understand information
// about this server.
type Info struct {
	ID           string `json:"server_id"`
	Version      string `json:"version"`
	GoVersion    string `json:"go"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	AuthRequired bool   `json:"auth_required"`
	SSLRequired  bool   `json:"ssl_required"` // DEPRECATED: ssl json used for older clients
	TLSRequired  bool   `json:"tls_required"`
	TLSVerify    bool   `json:"tls_verify"`
	MaxPayload   int    `json:"max_payload"`
	IP           string `json:"ip,omitempty"`
}

// Server is our main struct.
type Server struct {
	gcid uint64
	grid uint64
	stats
	mu            sync.Mutex
	info          Info
	infoJSON      []byte
	sl            *Sublist
	opts          *Options
	cAuth         Auth
	rAuth         Auth
	trace         bool
	debug         bool
	running       bool
	listener      net.Listener
	clients       map[uint64]*client
	routes        map[uint64]*client
	remotes       map[string]*client
	totalClients  uint64
	done          chan bool
	start         time.Time
	http          net.Listener
	httpReqStats  map[string]uint64
	routeListener net.Listener
	routeInfo     Info
	routeInfoJSON []byte
	rcQuit        chan bool
	grMu          sync.Mutex
	grTmpClients  map[uint64]*client
	grRunning     bool
	grWG          sync.WaitGroup // to wait on various go routines
}

// Make sure all are 64bits for atomic use
type stats struct {
	inMsgs        int64
	outMsgs       int64
	inBytes       int64
	outBytes      int64
	slowConsumers int64
}

// New will setup a new server struct after parsing the options.
func New(opts *Options) *Server {
	processOptions(opts)

	// Process TLS options, including whether we require client certificates.
	tlsReq := opts.TLSConfig != nil
	verify := (tlsReq && opts.TLSConfig.ClientAuth == tls.RequireAnyClientCert)

	info := Info{
		ID:           genID(),
		Version:      VERSION,
		GoVersion:    runtime.Version(),
		Host:         opts.Host,
		Port:         opts.Port,
		AuthRequired: false,
		TLSRequired:  tlsReq,
		SSLRequired:  tlsReq,
		TLSVerify:    verify,
		MaxPayload:   opts.MaxPayload,
	}

	s := &Server{
		info:  info,
		sl:    NewSublist(),
		opts:  opts,
		debug: opts.Debug,
		trace: opts.Trace,
		done:  make(chan bool, 1),
		start: time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// For tracking clients
	s.clients = make(map[uint64]*client)

	// For tracking connections that are not yet registered
	// in s.routes, but for which readLoop has started.
	s.grTmpClients = make(map[uint64]*client)

	// For tracking routes and their remote ids
	s.routes = make(map[uint64]*client)
	s.remotes = make(map[string]*client)

	// Used to kick out all of the route
	// connect Go routines.
	s.rcQuit = make(chan bool)
	s.generateServerInfoJSON()
	s.handleSignals()

	return s
}

// SetClientAuthMethod sets the authentication method for clients.
func (s *Server) SetClientAuthMethod(authMethod Auth) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.info.AuthRequired = true
	s.cAuth = authMethod

	s.generateServerInfoJSON()
}

// SetRouteAuthMethod sets the authentication method for routes.
func (s *Server) SetRouteAuthMethod(authMethod Auth) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rAuth = authMethod
}

func (s *Server) generateServerInfoJSON() {
	// Generate the info json
	b, err := json.Marshal(s.info)
	if err != nil {
		Fatalf("Error marshalling INFO JSON: %+v\n", err)
		return
	}
	s.infoJSON = []byte(fmt.Sprintf("INFO %s %s", b, CR_LF))
}

// PrintAndDie is exported for access in other packages.
func PrintAndDie(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(1)
}

// PrintServerAndExit will print our version and exit.
func PrintServerAndExit() {
	fmt.Printf("nats-server version %s\n", VERSION)
	os.Exit(0)
}

// Signal Handling
func (s *Server) handleSignals() {
	if s.opts.NoSigs {
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			Debugf("Trapped Signal; %v", sig)
			// FIXME, trip running?
			Noticef("Server Exiting..")
			os.Exit(0)
		}
	}()
}

// Protected check on running state
func (s *Server) isRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Server) logPid() {
	pidStr := strconv.Itoa(os.Getpid())
	err := ioutil.WriteFile(s.opts.PidFile, []byte(pidStr), 0660)
	if err != nil {
		PrintAndDie(fmt.Sprintf("Could not write pidfile: %v\n", err))
	}
}

// Start up the server, this will block.
// Start via a Go routine if needed.
func (s *Server) Start() {
	Noticef("Starting nats-server version %s", VERSION)
	Debugf("Go build version %s", s.info.GoVersion)

	s.running = true
	s.grMu.Lock()
	s.grRunning = true
	s.grMu.Unlock()

	// Log the pid to a file
	if s.opts.PidFile != _EMPTY_ {
		s.logPid()
	}

	// Start up the http server if needed.
	if s.opts.HTTPPort != 0 {
		s.StartHTTPMonitoring()
	}

	// Start up the https server if needed.
	if s.opts.HTTPSPort != 0 {
		if s.opts.TLSConfig == nil {
			Fatalf("TLS cert and key required for HTTPS")
			return
		}
		s.StartHTTPSMonitoring()
	}

	// Start up routing as well if needed.
	if s.opts.ClusterPort != 0 {
		s.StartRouting()
	}

	// Pprof http endpoint for the profiler.
	if s.opts.ProfPort != 0 {
		s.StartProfiler()
	}

	// Wait for clients.
	s.AcceptLoop()
}

// Shutdown will shutdown the server instance by kicking out the AcceptLoop
// and closing all associated clients.
func (s *Server) Shutdown() {
	s.mu.Lock()

	// Prevent issues with multiple calls.
	if !s.running {
		s.mu.Unlock()
		return
	}

	s.running = false
	s.grMu.Lock()
	s.grRunning = false
	s.grMu.Unlock()

	conns := make(map[uint64]*client)

	// Copy off the clients
	for i, c := range s.clients {
		conns[i] = c
	}
	// Copy off the connections that are not yet registered
	// in s.routes, but for which the readLoop has started
	s.grMu.Lock()
	for i, c := range s.grTmpClients {
		conns[i] = c
	}
	s.grMu.Unlock()
	// Copy off the routes
	for i, r := range s.routes {
		conns[i] = r
	}

	// Number of done channel responses we expect.
	doneExpected := 0

	// Kick client AcceptLoop()
	if s.listener != nil {
		doneExpected++
		s.listener.Close()
		s.listener = nil
	}

	// Kick route AcceptLoop()
	if s.routeListener != nil {
		doneExpected++
		s.routeListener.Close()
		s.routeListener = nil
	}

	// Kick HTTP monitoring if its running
	if s.http != nil {
		doneExpected++
		s.http.Close()
		s.http = nil
	}

	// Release the solicited routes connect go routines.
	close(s.rcQuit)

	s.mu.Unlock()

	// Close client and route connections
	for _, c := range conns {
		c.closeConnection()
	}

	// Block until the accept loops exit
	for doneExpected > 0 {
		<-s.done
		doneExpected--
	}

	// Wait for go routines to be done.
	s.grWG.Wait()
}

// AcceptLoop is exported for easier testing.
func (s *Server) AcceptLoop() {
	hp := net.JoinHostPort(s.opts.Host, strconv.Itoa(s.opts.Port))
	Noticef("Listening for client connections on %s", hp)
	l, e := net.Listen("tcp", hp)
	if e != nil {
		Fatalf("Error listening on port: %s, %q", hp, e)
		return
	}

	// Alert of TLS enabled.
	if s.opts.TLSConfig != nil {
		Noticef("TLS required for client connections")
	}

	Debugf("Server id is %s", s.info.ID)
	Noticef("Server is ready")

	// Setup state that can enable shutdown
	s.mu.Lock()
	s.listener = l

	// If server was started with RANDOM_PORT (-1), opts.Port would be equal
	// to 0 at the beginning this function. So we need to get the actual port
	if s.opts.Port == 0 {
		// Write resolved port back to options.
		_, port, err := net.SplitHostPort(l.Addr().String())
		if err != nil {
			Fatalf("Error parsing server address (%s): %s", l.Addr().String(), e)
			s.mu.Unlock()
			return
		}
		portNum, err := strconv.Atoi(port)
		if err != nil {
			Fatalf("Error parsing server address (%s): %s", l.Addr().String(), e)
			s.mu.Unlock()
			return
		}
		s.opts.Port = portNum
	}
	s.mu.Unlock()

	tmpDelay := ACCEPT_MIN_SLEEP

	for s.isRunning() {
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				Debugf("Temporary Client Accept Error(%v), sleeping %dms",
					ne, tmpDelay/time.Millisecond)
				time.Sleep(tmpDelay)
				tmpDelay *= 2
				if tmpDelay > ACCEPT_MAX_SLEEP {
					tmpDelay = ACCEPT_MAX_SLEEP
				}
			} else if s.isRunning() {
				Noticef("Accept error: %v", err)
			}
			continue
		}
		tmpDelay = ACCEPT_MIN_SLEEP
		s.startGoRoutine(func() {
			s.createClient(conn)
			s.grWG.Done()
		})
	}
	Noticef("Server Exiting..")
	s.done <- true
}

// StartProfiler is called to enable dynamic profiling.
func (s *Server) StartProfiler() {
	Noticef("Starting profiling on http port %d", s.opts.ProfPort)
	hp := net.JoinHostPort(s.opts.Host, strconv.Itoa(s.opts.ProfPort))
	go func() {
		err := http.ListenAndServe(hp, nil)
		if err != nil {
			Fatalf("error starting monitor server: %s", err)
		}
	}()
}

// StartHTTPMonitoring will enable the HTTP monitoring port.
func (s *Server) StartHTTPMonitoring() {
	s.startMonitoring(false)
}

// StartHTTPSMonitoring will enable the HTTPS monitoring port.
func (s *Server) StartHTTPSMonitoring() {
	s.startMonitoring(true)
}

// HTTP endpoints
const (
	RootPath    = "/"
	VarzPath    = "/varz"
	ConnzPath   = "/connz"
	RoutezPath  = "/routez"
	SubszPath   = "/subsz"
	StackszPath = "/stacksz"
)

// Start the monitoring server
func (s *Server) startMonitoring(secure bool) {

	// Used to track HTTP requests
	s.httpReqStats = map[string]uint64{
		RootPath:   0,
		VarzPath:   0,
		ConnzPath:  0,
		RoutezPath: 0,
		SubszPath:  0,
	}

	var hp string
	var err error

	if secure {
		hp = net.JoinHostPort(s.opts.HTTPHost, strconv.Itoa(s.opts.HTTPSPort))
		Noticef("Starting https monitor on %s", hp)
		config := *s.opts.TLSConfig
		config.ClientAuth = tls.NoClientCert
		s.http, err = tls.Listen("tcp", hp, &config)

	} else {
		hp = net.JoinHostPort(s.opts.HTTPHost, strconv.Itoa(s.opts.HTTPPort))
		Noticef("Starting http monitor on %s", hp)
		s.http, err = net.Listen("tcp", hp)
	}

	if err != nil {
		Fatalf("Can't listen to the monitor port: %v", err)
		return
	}

	mux := http.NewServeMux()

	// Root
	mux.HandleFunc(RootPath, s.HandleRoot)
	// Varz
	mux.HandleFunc(VarzPath, s.HandleVarz)
	// Connz
	mux.HandleFunc(ConnzPath, s.HandleConnz)
	// Routez
	mux.HandleFunc(RoutezPath, s.HandleRoutez)
	// Subz
	mux.HandleFunc(SubszPath, s.HandleSubsz)
	// Subz alias for backwards compatibility
	mux.HandleFunc("/subscriptionsz", s.HandleSubsz)
	// Stacksz
	mux.HandleFunc(StackszPath, s.HandleStacksz)

	srv := &http.Server{
		Addr:           hp,
		Handler:        mux,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		srv.Serve(s.http)
		srv.Handler = nil
		s.done <- true
	}()
}

func (s *Server) createClient(conn net.Conn) *client {
	c := &client{srv: s, nc: conn, opts: defaultOpts, mpay: s.info.MaxPayload, start: time.Now()}

	// Grab JSON info string
	s.mu.Lock()
	info := s.infoJSON
	authRequired := s.info.AuthRequired
	tlsRequired := s.info.TLSRequired
	s.totalClients++
	s.mu.Unlock()

	// Grab lock
	c.mu.Lock()

	// Initialize
	c.initClient()

	c.Debugf("Client connection created")

	// Check for Auth
	if authRequired {
		ttl := secondsToDuration(s.opts.AuthTimeout)
		c.setAuthTimer(ttl)
	}

	// Send our information.
	c.sendInfo(info)

	// Unlock to register
	c.mu.Unlock()

	// Register with the server.
	s.mu.Lock()
	// If server is not running, Shutdown() may have already gathered the
	// list of connections to close. It won't contain this one, so we need
	// to bail out now otherwise the readLoop started down there would not
	// be interrupted.
	if !s.running {
		s.mu.Unlock()
		return c
	}
	s.clients[c.cid] = c
	s.mu.Unlock()

	// Re-Grab lock
	c.mu.Lock()

	// Check for TLS
	if tlsRequired {
		c.Debugf("Starting TLS client connection handshake")
		c.nc = tls.Server(c.nc, s.opts.TLSConfig)
		conn := c.nc.(*tls.Conn)

		// Setup the timeout
		ttl := secondsToDuration(s.opts.TLSTimeout)
		time.AfterFunc(ttl, func() { tlsTimeout(c, conn) })
		conn.SetReadDeadline(time.Now().Add(ttl))

		// Force handshake
		c.mu.Unlock()
		if err := conn.Handshake(); err != nil {
			c.Debugf("TLS handshake error: %v", err)
			c.sendErr("Secure Connection - TLS Required")
			c.closeConnection()
			return nil
		}
		// Reset the read deadline
		conn.SetReadDeadline(time.Time{})

		// Re-Grab lock
		c.mu.Lock()
	}

	// The connection may have been closed
	if c.nc == nil {
		c.mu.Unlock()
		return c
	}

	if tlsRequired {
		// Rewrap bw
		c.bw = bufio.NewWriterSize(c.nc, startBufSize)
	}

	// Do final client initialization

	// Set the Ping timer
	c.setPingTimer()

	// Spin up the read loop.
	s.startGoRoutine(func() { c.readLoop() })

	if tlsRequired {
		c.Debugf("TLS handshake complete")
		cs := c.nc.(*tls.Conn).ConnectionState()
		c.Debugf("TLS version %s, cipher suite %s", tlsVersion(cs.Version), tlsCipher(cs.CipherSuite))
	}

	c.mu.Unlock()

	return c
}

// Handle closing down a connection when the handshake has timedout.
func tlsTimeout(c *client, conn *tls.Conn) {
	c.mu.Lock()
	nc := c.nc
	c.mu.Unlock()
	// Check if already closed
	if nc == nil {
		return
	}
	cs := conn.ConnectionState()
	if !cs.HandshakeComplete {
		c.Debugf("TLS handshake timeout")
		c.sendErr("Secure Connection - TLS Required")
		c.closeConnection()
	}
}

// Seems silly we have to write these
func tlsVersion(ver uint16) string {
	switch ver {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	}
	return fmt.Sprintf("Unknown [%x]", ver)
}

// We use hex here so we don't need multiple versions
func tlsCipher(cs uint16) string {
	switch cs {
	case 0x0005:
		return "TLS_RSA_WITH_RC4_128_SHA"
	case 0x000a:
		return "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
	case 0x002f:
		return "TLS_RSA_WITH_AES_128_CBC_SHA"
	case 0x0035:
		return "TLS_RSA_WITH_AES_256_CBC_SHA"
	case 0xc007:
		return "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA"
	case 0xc009:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
	case 0xc00a:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
	case 0xc011:
		return "TLS_ECDHE_RSA_WITH_RC4_128_SHA"
	case 0xc012:
		return "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA"
	case 0xc013:
		return "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
	case 0xc014:
		return "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
	case 0xc02f:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	case 0xc02b:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	case 0xc030:
		return "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case 0xc02c:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	}
	return fmt.Sprintf("Unknown [%x]", cs)
}

func (s *Server) checkClientAuth(c *client) bool {
	if s.cAuth == nil {
		return true
	}
	return s.cAuth.Check(c)
}

func (s *Server) checkRouterAuth(c *client) bool {
	if s.rAuth == nil {
		return true
	}
	return s.rAuth.Check(c)
}

// Check auth and return boolean indicating if client is ok
func (s *Server) checkAuth(c *client) bool {
	switch c.typ {
	case CLIENT:
		return s.checkClientAuth(c)
	case ROUTER:
		return s.checkRouterAuth(c)
	default:
		return false
	}
}

// Remove a client or route from our internal accounting.
func (s *Server) removeClient(c *client) {
	var rID string
	c.mu.Lock()
	cid := c.cid
	typ := c.typ
	r := c.route
	if r != nil {
		rID = r.remoteID
	}
	c.mu.Unlock()

	s.mu.Lock()
	switch typ {
	case CLIENT:
		delete(s.clients, cid)
	case ROUTER:
		delete(s.routes, cid)
		if r != nil {
			rc, ok := s.remotes[rID]
			// Only delete it if it is us..
			if ok && c == rc {
				delete(s.remotes, rID)
			}
		}
	}
	s.mu.Unlock()
}

/////////////////////////////////////////////////////////////////
// These are some helpers for accounting in functional tests.
/////////////////////////////////////////////////////////////////

// NumRoutes will report the number of registered routes.
func (s *Server) NumRoutes() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.routes)
}

// NumRemotes will report number of registered remotes.
func (s *Server) NumRemotes() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.remotes)
}

// NumClients will report the number of registered clients.
func (s *Server) NumClients() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}

// NumSubscriptions will report how many subscriptions are active.
func (s *Server) NumSubscriptions() uint32 {
	s.mu.Lock()
	subs := s.sl.Count()
	s.mu.Unlock()
	return subs
}

// Addr will return the net.Addr object for the current listener.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// GetListenEndpoint will return a string of the form host:port suitable for
// a connect. Will return empty string if the server is not ready to accept
// client connections.
func (s *Server) GetListenEndpoint() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Wait for the listener to be set, see note about RANDOM_PORT below
	if s.listener == nil {
		return ""
	}

	host := s.opts.Host

	// On windows, a connect with host "0.0.0.0" (or "::") will fail.
	// We replace it with "localhost" when that's the case.
	if host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "localhost"
	}

	// Return the opts's Host and Port. Note that the Port may be set
	// when the listener is started, due to the use of RANDOM_PORT
	return net.JoinHostPort(host, strconv.Itoa(s.opts.Port))
}

// GetRouteListenEndpoint will return a string of the form host:port suitable
// for a connect. Will return empty string if the server is not configured for
// routing or not ready to accept route connections.
func (s *Server) GetRouteListenEndpoint() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.routeListener == nil {
		return ""
	}

	host := s.opts.ClusterHost

	// On windows, a connect with host "0.0.0.0" (or "::") will fail.
	// We replace it with "localhost" when that's the case.
	if host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "localhost"
	}

	// Return the cluster's Host and Port.
	return net.JoinHostPort(host, strconv.Itoa(s.opts.ClusterPort))
}

// ID returns the server's ID
func (s *Server) ID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.info.ID
}

func (s *Server) startGoRoutine(f func()) {
	s.grMu.Lock()
	if s.grRunning {
		s.grWG.Add(1)
		go f()
	}
	s.grMu.Unlock()
}
