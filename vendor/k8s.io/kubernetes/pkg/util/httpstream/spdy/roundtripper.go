/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spdy

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	apierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/util/httpstream"
	"k8s.io/kubernetes/third_party/golang/netutil"
)

// SpdyRoundTripper knows how to upgrade an HTTP request to one that supports
// multiplexed streams. After RoundTrip() is invoked, Conn will be set
// and usable. SpdyRoundTripper implements the UpgradeRoundTripper interface.
type SpdyRoundTripper struct {
	//tlsConfig holds the TLS configuration settings to use when connecting
	//to the remote server.
	tlsConfig *tls.Config

	/* TODO according to http://golang.org/pkg/net/http/#RoundTripper, a RoundTripper
	   must be safe for use by multiple concurrent goroutines. If this is absolutely
	   necessary, we could keep a map from http.Request to net.Conn. In practice,
	   a client will create an http.Client, set the transport to a new insteace of
	   SpdyRoundTripper, and use it a single time, so this hopefully won't be an issue.
	*/
	// conn is the underlying network connection to the remote server.
	conn net.Conn

	// Dialer is the dialer used to connect.  Used if non-nil.
	Dialer *net.Dialer

	// proxier knows which proxy to use given a request, defaults to http.ProxyFromEnvironment
	// Used primarily for mocking the proxy discovery in tests.
	proxier func(req *http.Request) (*url.URL, error)
}

// NewRoundTripper creates a new SpdyRoundTripper that will use
// the specified tlsConfig.
func NewRoundTripper(tlsConfig *tls.Config) httpstream.UpgradeRoundTripper {
	return NewSpdyRoundTripper(tlsConfig)
}

// NewSpdyRoundTripper creates a new SpdyRoundTripper that will use
// the specified tlsConfig. This function is mostly meant for unit tests.
func NewSpdyRoundTripper(tlsConfig *tls.Config) *SpdyRoundTripper {
	return &SpdyRoundTripper{tlsConfig: tlsConfig}
}

// dial dials the host specified by req, using TLS if appropriate, optionally
// using a proxy server if one is configured via environment variables.
func (s *SpdyRoundTripper) dial(req *http.Request) (net.Conn, error) {
	proxier := s.proxier
	if proxier == nil {
		proxier = http.ProxyFromEnvironment
	}
	proxyURL, err := proxier(req)
	if err != nil {
		return nil, err
	}

	if proxyURL == nil {
		return s.dialWithoutProxy(req.URL)
	}

	// ensure we use a canonical host with proxyReq
	targetHost := netutil.CanonicalAddr(req.URL)

	// proxying logic adapted from http://blog.h6t.eu/post/74098062923/golang-websocket-with-http-proxy-support
	proxyReq := http.Request{
		Method: "CONNECT",
		URL:    &url.URL{},
		Host:   targetHost,
	}

	if pa := s.proxyAuth(proxyURL); pa != "" {
		proxyReq.Header = http.Header{}
		proxyReq.Header.Set("Proxy-Authorization", pa)
	}

	proxyDialConn, err := s.dialWithoutProxy(proxyURL)
	if err != nil {
		return nil, err
	}

	proxyClientConn := httputil.NewProxyClientConn(proxyDialConn, nil)
	_, err = proxyClientConn.Do(&proxyReq)
	if err != nil && err != httputil.ErrPersistEOF {
		return nil, err
	}

	rwc, _ := proxyClientConn.Hijack()

	if req.URL.Scheme != "https" {
		return rwc, nil
	}

	host, _, err := net.SplitHostPort(targetHost)
	if err != nil {
		return nil, err
	}

	if len(s.tlsConfig.ServerName) == 0 {
		s.tlsConfig.ServerName = host
	}

	tlsConn := tls.Client(rwc, s.tlsConfig)

	// need to manually call Handshake() so we can call VerifyHostname() below
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}

	// Return if we were configured to skip validation
	if s.tlsConfig != nil && s.tlsConfig.InsecureSkipVerify {
		return tlsConn, nil
	}

	if err := tlsConn.VerifyHostname(host); err != nil {
		return nil, err
	}

	return tlsConn, nil
}

// dialWithoutProxy dials the host specified by url, using TLS if appropriate.
func (s *SpdyRoundTripper) dialWithoutProxy(url *url.URL) (net.Conn, error) {
	dialAddr := netutil.CanonicalAddr(url)

	if url.Scheme == "http" {
		if s.Dialer == nil {
			return net.Dial("tcp", dialAddr)
		} else {
			return s.Dialer.Dial("tcp", dialAddr)
		}
	}

	// TODO validate the TLSClientConfig is set up?
	var conn *tls.Conn
	var err error
	if s.Dialer == nil {
		conn, err = tls.Dial("tcp", dialAddr, s.tlsConfig)
	} else {
		conn, err = tls.DialWithDialer(s.Dialer, "tcp", dialAddr, s.tlsConfig)
	}
	if err != nil {
		return nil, err
	}

	// Return if we were configured to skip validation
	if s.tlsConfig != nil && s.tlsConfig.InsecureSkipVerify {
		return conn, nil
	}

	host, _, err := net.SplitHostPort(dialAddr)
	if err != nil {
		return nil, err
	}
	err = conn.VerifyHostname(host)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// proxyAuth returns, for a given proxy URL, the value to be used for the Proxy-Authorization header
func (s *SpdyRoundTripper) proxyAuth(proxyURL *url.URL) string {
	if proxyURL == nil || proxyURL.User == nil {
		return ""
	}
	credentials := proxyURL.User.String()
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encodedAuth)
}

// RoundTrip executes the Request and upgrades it. After a successful upgrade,
// clients may call SpdyRoundTripper.Connection() to retrieve the upgraded
// connection.
func (s *SpdyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// TODO what's the best way to clone the request?
	r := *req
	req = &r
	req.Header.Add(httpstream.HeaderConnection, httpstream.HeaderUpgrade)
	req.Header.Add(httpstream.HeaderUpgrade, HeaderSpdy31)

	conn, err := s.dial(req)
	if err != nil {
		return nil, err
	}

	err = req.Write(conn)
	if err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}

	s.conn = conn

	return resp, nil
}

// NewConnection validates the upgrade response, creating and returning a new
// httpstream.Connection if there were no errors.
func (s *SpdyRoundTripper) NewConnection(resp *http.Response) (httpstream.Connection, error) {
	connectionHeader := strings.ToLower(resp.Header.Get(httpstream.HeaderConnection))
	upgradeHeader := strings.ToLower(resp.Header.Get(httpstream.HeaderUpgrade))
	if (resp.StatusCode != http.StatusSwitchingProtocols) || !strings.Contains(connectionHeader, strings.ToLower(httpstream.HeaderUpgrade)) || !strings.Contains(upgradeHeader, strings.ToLower(HeaderSpdy31)) {
		defer resp.Body.Close()
		responseError := ""
		responseErrorBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			responseError = "unable to read error from server response"
		} else {
			// TODO: I don't belong here, I should be abstracted from this class
			if obj, _, err := api.Codecs.UniversalDecoder().Decode(responseErrorBytes, nil, &unversioned.Status{}); err == nil {
				if status, ok := obj.(*unversioned.Status); ok {
					return nil, &apierrors.StatusError{ErrStatus: *status}
				}
			}
			responseError = string(responseErrorBytes)
			responseError = strings.TrimSpace(responseError)
		}

		return nil, fmt.Errorf("unable to upgrade connection: %s", responseError)
	}

	return NewClientConnection(s.conn)
}
