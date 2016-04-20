/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package rest

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/websocket"

	utilnet "k8s.io/kubernetes/pkg/util/net"
	"k8s.io/kubernetes/pkg/util/proxy"
)

type fakeResponder struct {
	called bool
	err    error
}

func (r *fakeResponder) Error(err error) {
	if r.called {
		panic("called twice")
	}
	r.called = true
	r.err = err
}

type SimpleBackendHandler struct {
	requestURL     url.URL
	requestHeader  http.Header
	requestBody    []byte
	requestMethod  string
	responseBody   string
	responseHeader map[string]string
	t              *testing.T
}

func (s *SimpleBackendHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.requestURL = *req.URL
	s.requestHeader = req.Header
	s.requestMethod = req.Method
	var err error
	s.requestBody, err = ioutil.ReadAll(req.Body)
	if err != nil {
		s.t.Errorf("Unexpected error: %v", err)
		return
	}

	if s.responseHeader != nil {
		for k, v := range s.responseHeader {
			w.Header().Add(k, v)
		}
	}
	w.Write([]byte(s.responseBody))
}

func validateParameters(t *testing.T, name string, actual url.Values, expected map[string]string) {
	for k, v := range expected {
		actualValue, ok := actual[k]
		if !ok {
			t.Errorf("%s: Expected parameter %s not received", name, k)
			continue
		}
		if actualValue[0] != v {
			t.Errorf("%s: Parameter %s values don't match. Actual: %#v, Expected: %s",
				name, k, actualValue, v)
		}
	}
}

func validateHeaders(t *testing.T, name string, actual http.Header, expected map[string]string, notExpected []string) {
	for k, v := range expected {
		actualValue, ok := actual[k]
		if !ok {
			t.Errorf("%s: Expected header %s not received", name, k)
			continue
		}
		if actualValue[0] != v {
			t.Errorf("%s: Header %s values don't match. Actual: %s, Expected: %s",
				name, k, actualValue, v)
		}
	}
	if notExpected == nil {
		return
	}
	for _, h := range notExpected {
		if _, present := actual[h]; present {
			t.Errorf("%s: unexpected header: %s", name, h)
		}
	}
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		name                  string
		method                string
		requestPath           string
		expectedPath          string
		requestBody           string
		requestParams         map[string]string
		requestHeader         map[string]string
		responseHeader        map[string]string
		expectedRespHeader    map[string]string
		notExpectedRespHeader []string
		upgradeRequired       bool
		expectError           func(err error) bool
	}{
		{
			name:         "root path, simple get",
			method:       "GET",
			requestPath:  "/",
			expectedPath: "/",
		},
		{
			name:            "no upgrade header sent",
			method:          "GET",
			requestPath:     "/",
			upgradeRequired: true,
			expectError: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "Upgrade request required")
			},
		},
		{
			name:         "simple path, get",
			method:       "GET",
			requestPath:  "/path/to/test",
			expectedPath: "/path/to/test",
		},
		{
			name:          "request params",
			method:        "POST",
			requestPath:   "/some/path/",
			expectedPath:  "/some/path/",
			requestParams: map[string]string{"param1": "value/1", "param2": "value%2"},
			requestBody:   "test request body",
		},
		{
			name:          "request headers",
			method:        "PUT",
			requestPath:   "/some/path",
			expectedPath:  "/some/path",
			requestHeader: map[string]string{"Header1": "value1", "Header2": "value2"},
		},
		{
			name:         "empty path - slash should be added",
			method:       "GET",
			requestPath:  "",
			expectedPath: "/",
		},
		{
			name:         "remove CORS headers",
			method:       "GET",
			requestPath:  "/some/path",
			expectedPath: "/some/path",
			responseHeader: map[string]string{
				"Header1":                      "value1",
				"Access-Control-Allow-Origin":  "some.server",
				"Access-Control-Allow-Methods": "GET"},
			expectedRespHeader: map[string]string{
				"Header1": "value1",
			},
			notExpectedRespHeader: []string{
				"Access-Control-Allow-Origin",
				"Access-Control-Allow-Methods",
			},
		},
	}

	for i, test := range tests {
		func() {
			backendResponse := "<html><head></head><body><a href=\"/test/path\">Hello</a></body></html>"
			backendResponseHeader := test.responseHeader
			// Test a simple header if not specified in the test
			if backendResponseHeader == nil && test.expectedRespHeader == nil {
				backendResponseHeader = map[string]string{"Content-Type": "text/html"}
				test.expectedRespHeader = map[string]string{"Content-Type": "text/html"}
			}
			backendHandler := &SimpleBackendHandler{
				responseBody:   backendResponse,
				responseHeader: backendResponseHeader,
			}
			backendServer := httptest.NewServer(backendHandler)
			// TODO: Uncomment when fix #19254
			// defer backendServer.Close()

			responder := &fakeResponder{}
			backendURL, _ := url.Parse(backendServer.URL)
			backendURL.Path = test.requestPath
			proxyHandler := &UpgradeAwareProxyHandler{
				Location:        backendURL,
				Responder:       responder,
				UpgradeRequired: test.upgradeRequired,
			}
			proxyServer := httptest.NewServer(proxyHandler)
			// TODO: Uncomment when fix #19254
			// defer proxyServer.Close()
			proxyURL, _ := url.Parse(proxyServer.URL)
			proxyURL.Path = test.requestPath
			paramValues := url.Values{}
			for k, v := range test.requestParams {
				paramValues[k] = []string{v}
			}
			proxyURL.RawQuery = paramValues.Encode()
			var requestBody io.Reader
			if test.requestBody != "" {
				requestBody = bytes.NewBufferString(test.requestBody)
			}
			req, err := http.NewRequest(test.method, proxyURL.String(), requestBody)
			if test.requestHeader != nil {
				header := http.Header{}
				for k, v := range test.requestHeader {
					header.Add(k, v)
				}
				req.Header = header
			}
			if err != nil {
				t.Errorf("Error creating client request: %v", err)
			}
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				t.Errorf("Error from proxy request: %v", err)
			}

			if test.expectError != nil {
				if !responder.called {
					t.Errorf("%d: responder was not invoked", i)
					return
				}
				if !test.expectError(responder.err) {
					t.Errorf("%d: unexpected error: %v", i, responder.err)
				}
				return
			}

			// Validate backend request
			// Method
			if backendHandler.requestMethod != test.method {
				t.Errorf("Unexpected request method: %s. Expected: %s",
					backendHandler.requestMethod, test.method)
			}

			// Body
			if string(backendHandler.requestBody) != test.requestBody {
				t.Errorf("Unexpected request body: %s. Expected: %s",
					string(backendHandler.requestBody), test.requestBody)
			}

			// Path
			if backendHandler.requestURL.Path != test.expectedPath {
				t.Errorf("Unexpected request path: %s", backendHandler.requestURL.Path)
			}
			// Parameters
			validateParameters(t, test.name, backendHandler.requestURL.Query(), test.requestParams)

			// Headers
			validateHeaders(t, test.name+" backend request", backendHandler.requestHeader,
				test.requestHeader, nil)

			// Validate proxy response

			// Response Headers
			validateHeaders(t, test.name+" backend headers", res.Header, test.expectedRespHeader, test.notExpectedRespHeader)

			// Validate Body
			responseBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected error reading response body: %v", err)
			}
			if rb := string(responseBody); rb != backendResponse {
				t.Errorf("Did not get expected response body: %s. Expected: %s", rb, backendResponse)
			}

			// Error
			if responder.called {
				t.Errorf("Unexpected proxy handler error: %v", responder.err)
			}
		}()
	}
}

func TestProxyUpgrade(t *testing.T) {

	localhostPool := x509.NewCertPool()
	if !localhostPool.AppendCertsFromPEM(localhostCert) {
		t.Errorf("error setting up localhostCert pool")
	}

	testcases := map[string]struct {
		ServerFunc     func(http.Handler) *httptest.Server
		ProxyTransport http.RoundTripper
	}{
		"http": {
			ServerFunc:     httptest.NewServer,
			ProxyTransport: nil,
		},
		"https (invalid hostname + InsecureSkipVerify)": {
			ServerFunc: func(h http.Handler) *httptest.Server {
				cert, err := tls.X509KeyPair(exampleCert, exampleKey)
				if err != nil {
					t.Errorf("https (invalid hostname): proxy_test: %v", err)
				}
				ts := httptest.NewUnstartedServer(h)
				ts.TLS = &tls.Config{
					Certificates: []tls.Certificate{cert},
				}
				ts.StartTLS()
				return ts
			},
			ProxyTransport: utilnet.SetTransportDefaults(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}),
		},
		"https (valid hostname + RootCAs)": {
			ServerFunc: func(h http.Handler) *httptest.Server {
				cert, err := tls.X509KeyPair(localhostCert, localhostKey)
				if err != nil {
					t.Errorf("https (valid hostname): proxy_test: %v", err)
				}
				ts := httptest.NewUnstartedServer(h)
				ts.TLS = &tls.Config{
					Certificates: []tls.Certificate{cert},
				}
				ts.StartTLS()
				return ts
			},
			ProxyTransport: utilnet.SetTransportDefaults(&http.Transport{TLSClientConfig: &tls.Config{RootCAs: localhostPool}}),
		},
		"https (valid hostname + RootCAs + custom dialer)": {
			ServerFunc: func(h http.Handler) *httptest.Server {
				cert, err := tls.X509KeyPair(localhostCert, localhostKey)
				if err != nil {
					t.Errorf("https (valid hostname): proxy_test: %v", err)
				}
				ts := httptest.NewUnstartedServer(h)
				ts.TLS = &tls.Config{
					Certificates: []tls.Certificate{cert},
				}
				ts.StartTLS()
				return ts
			},
			ProxyTransport: utilnet.SetTransportDefaults(&http.Transport{Dial: net.Dial, TLSClientConfig: &tls.Config{RootCAs: localhostPool}}),
		},
	}

	for k, tc := range testcases {

		backendServer := tc.ServerFunc(websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			body := make([]byte, 5)
			ws.Read(body)
			ws.Write([]byte("hello " + string(body)))
		}))
		// TODO: Uncomment when fix #19254
		// defer backendServer.Close()

		serverURL, _ := url.Parse(backendServer.URL)
		proxyHandler := &UpgradeAwareProxyHandler{
			Location:  serverURL,
			Transport: tc.ProxyTransport,
		}
		proxy := httptest.NewServer(proxyHandler)
		// TODO: Uncomment when fix #19254
		// defer proxy.Close()

		ws, err := websocket.Dial("ws://"+proxy.Listener.Addr().String()+"/some/path", "", "http://127.0.0.1/")
		if err != nil {
			t.Fatalf("%s: websocket dial err: %s", k, err)
		}
		defer ws.Close()

		if _, err := ws.Write([]byte("world")); err != nil {
			t.Fatalf("%s: write err: %s", k, err)
		}

		response := make([]byte, 20)
		n, err := ws.Read(response)
		if err != nil {
			t.Fatalf("%s: read err: %s", k, err)
		}
		if e, a := "hello world", string(response[0:n]); e != a {
			t.Fatalf("%s: expected '%#v', got '%#v'", k, e, a)
		}
	}
}

func TestDefaultProxyTransport(t *testing.T) {
	tests := []struct {
		name,
		url,
		location,
		expectedScheme,
		expectedHost,
		expectedPathPrepend string
	}{

		{
			name:                "simple path",
			url:                 "http://test.server:8080/a/test/location",
			location:            "http://localhost/location",
			expectedScheme:      "http",
			expectedHost:        "test.server:8080",
			expectedPathPrepend: "/a/test",
		},
		{
			name:                "empty path",
			url:                 "http://test.server:8080/a/test/",
			location:            "http://localhost",
			expectedScheme:      "http",
			expectedHost:        "test.server:8080",
			expectedPathPrepend: "/a/test",
		},
		{
			name:                "location ending in slash",
			url:                 "http://test.server:8080/a/test/",
			location:            "http://localhost/",
			expectedScheme:      "http",
			expectedHost:        "test.server:8080",
			expectedPathPrepend: "/a/test",
		},
	}

	for _, test := range tests {
		locURL, _ := url.Parse(test.location)
		URL, _ := url.Parse(test.url)
		h := UpgradeAwareProxyHandler{
			Location: locURL,
		}
		result := h.defaultProxyTransport(URL, nil)
		transport := result.(*corsRemovingTransport).RoundTripper.(*proxy.Transport)
		if transport.Scheme != test.expectedScheme {
			t.Errorf("%s: unexpected scheme. Actual: %s, Expected: %s", test.name, transport.Scheme, test.expectedScheme)
		}
		if transport.Host != test.expectedHost {
			t.Errorf("%s: unexpected host. Actual: %s, Expected: %s", test.name, transport.Host, test.expectedHost)
		}
		if transport.PathPrepend != test.expectedPathPrepend {
			t.Errorf("%s: unexpected path prepend. Actual: %s, Expected: %s", test.name, transport.PathPrepend, test.expectedPathPrepend)
		}
	}
}

func TestProxyRequestContentLengthAndTransferEncoding(t *testing.T) {
	chunk := func(data []byte) []byte {
		out := &bytes.Buffer{}
		chunker := httputil.NewChunkedWriter(out)
		for _, b := range data {
			if _, err := chunker.Write([]byte{b}); err != nil {
				panic(err)
			}
		}
		chunker.Close()
		out.Write([]byte("\r\n"))
		return out.Bytes()
	}

	zip := func(data []byte) []byte {
		out := &bytes.Buffer{}
		zipper := gzip.NewWriter(out)
		if _, err := zipper.Write(data); err != nil {
			panic(err)
		}
		zipper.Close()
		return out.Bytes()
	}

	sampleData := []byte("abcde")

	table := map[string]struct {
		reqHeaders http.Header
		reqBody    []byte

		expectedHeaders http.Header
		expectedBody    []byte
	}{
		"content-length": {
			reqHeaders: http.Header{
				"Content-Length": []string{"5"},
			},
			reqBody: sampleData,

			expectedHeaders: http.Header{
				"Content-Length":    []string{"5"},
				"Content-Encoding":  nil, // none set
				"Transfer-Encoding": nil, // none set
			},
			expectedBody: sampleData,
		},

		"content-length + identity transfer-encoding": {
			reqHeaders: http.Header{
				"Content-Length":    []string{"5"},
				"Transfer-Encoding": []string{"identity"},
			},
			reqBody: sampleData,

			expectedHeaders: http.Header{
				"Content-Length":    []string{"5"},
				"Content-Encoding":  nil, // none set
				"Transfer-Encoding": nil, // gets removed
			},
			expectedBody: sampleData,
		},

		"content-length + gzip content-encoding": {
			reqHeaders: http.Header{
				"Content-Length":   []string{strconv.Itoa(len(zip(sampleData)))},
				"Content-Encoding": []string{"gzip"},
			},
			reqBody: zip(sampleData),

			expectedHeaders: http.Header{
				"Content-Length":    []string{strconv.Itoa(len(zip(sampleData)))},
				"Content-Encoding":  []string{"gzip"},
				"Transfer-Encoding": nil, // none set
			},
			expectedBody: zip(sampleData),
		},

		"chunked transfer-encoding": {
			reqHeaders: http.Header{
				"Transfer-Encoding": []string{"chunked"},
			},
			reqBody: chunk(sampleData),

			expectedHeaders: http.Header{
				"Content-Length":    nil, // none set
				"Content-Encoding":  nil, // none set
				"Transfer-Encoding": nil, // Transfer-Encoding gets removed
			},
			expectedBody: sampleData, // sample data is unchunked
		},

		"chunked transfer-encoding + gzip content-encoding": {
			reqHeaders: http.Header{
				"Content-Encoding":  []string{"gzip"},
				"Transfer-Encoding": []string{"chunked"},
			},
			reqBody: chunk(zip(sampleData)),

			expectedHeaders: http.Header{
				"Content-Length":    nil, // none set
				"Content-Encoding":  []string{"gzip"},
				"Transfer-Encoding": nil, // gets removed
			},
			expectedBody: zip(sampleData), // sample data is unchunked, but content-encoding is preserved
		},

		// "Transfer-Encoding: gzip" is not supported by go
		// See http/transfer.go#fixTransferEncoding (https://golang.org/src/net/http/transfer.go#L427)
		// Once it is supported, this test case should succeed
		//
		// "gzip+chunked transfer-encoding": {
		// 	reqHeaders: http.Header{
		// 		"Transfer-Encoding": []string{"chunked,gzip"},
		// 	},
		// 	reqBody: chunk(zip(sampleData)),
		//
		// 	expectedHeaders: http.Header{
		// 		"Content-Length":    nil, // no content-length headers
		// 		"Transfer-Encoding": nil, // Transfer-Encoding gets removed
		// 	},
		// 	expectedBody: sampleData,
		// },
	}

	successfulResponse := "backend passed tests"
	for k, item := range table {
		// Start the downstream server
		downstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Verify headers
			for header, v := range item.expectedHeaders {
				if !reflect.DeepEqual(v, req.Header[header]) {
					t.Errorf("%s: Expected headers for %s to be %v, got %v", k, header, v, req.Header[header])
				}
			}

			// Read body
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Errorf("%s: unexpected error %v", k, err)
			}
			req.Body.Close()

			// Verify length
			if req.ContentLength > 0 && req.ContentLength != int64(len(body)) {
				t.Errorf("%s: ContentLength was %d, len(data) was %d", k, req.ContentLength, len(body))
			}

			// Verify content
			if !bytes.Equal(item.expectedBody, body) {
				t.Errorf("%s: Expected %q, got %q", k, string(item.expectedBody), string(body))
			}

			// Write successful response
			w.Write([]byte(successfulResponse))
		}))
		// TODO: Uncomment when fix #19254
		// defer downstreamServer.Close()

		responder := &fakeResponder{}
		backendURL, _ := url.Parse(downstreamServer.URL)
		proxyHandler := &UpgradeAwareProxyHandler{
			Location:        backendURL,
			Responder:       responder,
			UpgradeRequired: false,
		}
		proxyServer := httptest.NewServer(proxyHandler)
		// TODO: Uncomment when fix #19254
		// defer proxyServer.Close()

		// Dial the proxy server
		conn, err := net.Dial(proxyServer.Listener.Addr().Network(), proxyServer.Listener.Addr().String())
		if err != nil {
			t.Errorf("unexpected error %v", err)
			continue
		}
		defer conn.Close()

		// Add standard http 1.1 headers
		if item.reqHeaders == nil {
			item.reqHeaders = http.Header{}
		}
		item.reqHeaders.Add("Connection", "close")
		item.reqHeaders.Add("Host", proxyServer.Listener.Addr().String())

		// Write the request headers
		if _, err := fmt.Fprint(conn, "POST / HTTP/1.1\r\n"); err != nil {
			t.Fatalf("%s unexpected error %v", k, err)
		}
		for header, values := range item.reqHeaders {
			for _, value := range values {
				if _, err := fmt.Fprintf(conn, "%s: %s\r\n", header, value); err != nil {
					t.Fatalf("%s: unexpected error %v", k, err)
				}
			}
		}
		// Header separator
		if _, err := fmt.Fprint(conn, "\r\n"); err != nil {
			t.Fatalf("%s: unexpected error %v", k, err)
		}
		// Body
		if _, err := conn.Write(item.reqBody); err != nil {
			t.Fatalf("%s: unexpected error %v", k, err)
		}

		// Read response
		response, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Errorf("%s: unexpected error %v", k, err)
			continue
		}
		if !strings.HasSuffix(string(response), successfulResponse) {
			t.Errorf("%s: Did not get successful response: %s", k, string(response))
			continue
		}
	}
}

// exampleCert was generated from crypto/tls/generate_cert.go with the following command:
//    go run generate_cert.go  --rsa-bits 512 --host example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var exampleCert = []byte(`-----BEGIN CERTIFICATE-----
MIIBcjCCAR6gAwIBAgIQBOUTYowZaENkZi0faI9DgTALBgkqhkiG9w0BAQswEjEQ
MA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2MDAw
MFowEjEQMA4GA1UEChMHQWNtZSBDbzBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQCZ
xfR3sgeHBraGFfF/24tTn4PRVAHOf2UOOxSQRs+aYjNqimFqf/SRIblQgeXdBJDR
gVK5F1Js2zwlehw0bHxRAgMBAAGjUDBOMA4GA1UdDwEB/wQEAwIApDATBgNVHSUE
DDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MBYGA1UdEQQPMA2CC2V4YW1w
bGUuY29tMAsGCSqGSIb3DQEBCwNBAI/mfBB8dm33IpUl+acSyWfL6gX5Wc0FFyVj
dKeesE1XBuPX1My/rzU6Oy/YwX7LOL4FaeNUS6bbL4axSLPKYSs=
-----END CERTIFICATE-----`)

var exampleKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBAJnF9HeyB4cGtoYV8X/bi1Ofg9FUAc5/ZQ47FJBGz5piM2qKYWp/
9JEhuVCB5d0EkNGBUrkXUmzbPCV6HDRsfFECAwEAAQJBAJLH9yPuButniACTn5L5
IJQw1mWQt6zBw9eCo41YWkA0866EgjC53aPZaRjXMp0uNJGdIsys2V5rCOOLWN2C
ODECIQDICHsi8QQQ9wpuJy8X5l8MAfxHL+DIqI84wQTeVM91FQIhAMTME8A18/7h
1Ad6drdnxAkuC0tX6Sx0LDozrmen+HFNAiAlcEDrt0RVkIcpOrg7tuhPLQf0oudl
Zvb3Xlj069awSQIgcT15E/43w2+RASifzVNhQ2MCTr1sSA8lL+xzK+REmnUCIBhQ
j4139pf8Re1J50zBxS/JlQfgDQi9sO9pYeiHIxNs
-----END RSA PRIVATE KEY-----`)

// localhostCert was generated from crypto/tls/generate_cert.go with the following command:
//     go run generate_cert.go  --rsa-bits 512 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIIBdzCCASOgAwIBAgIBADALBgkqhkiG9w0BAQUwEjEQMA4GA1UEChMHQWNtZSBD
bzAeFw03MDAxMDEwMDAwMDBaFw00OTEyMzEyMzU5NTlaMBIxEDAOBgNVBAoTB0Fj
bWUgQ28wWjALBgkqhkiG9w0BAQEDSwAwSAJBAN55NcYKZeInyTuhcCwFMhDHCmwa
IUSdtXdcbItRB/yfXGBhiex00IaLXQnSU+QZPRZWYqeTEbFSgihqi1PUDy8CAwEA
AaNoMGYwDgYDVR0PAQH/BAQDAgCkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1Ud
EwEB/wQFMAMBAf8wLgYDVR0RBCcwJYILZXhhbXBsZS5jb22HBH8AAAGHEAAAAAAA
AAAAAAAAAAAAAAEwCwYJKoZIhvcNAQEFA0EAAoQn/ytgqpiLcZu9XKbCJsJcvkgk
Se6AbGXgSlq+ZCEVo0qIwSgeBqmsJxUu7NCSOwVJLYNEBO2DtIxoYVk+MA==
-----END CERTIFICATE-----`)

// localhostKey is the private key for localhostCert.
var localhostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBPAIBAAJBAN55NcYKZeInyTuhcCwFMhDHCmwaIUSdtXdcbItRB/yfXGBhiex0
0IaLXQnSU+QZPRZWYqeTEbFSgihqi1PUDy8CAwEAAQJBAQdUx66rfh8sYsgfdcvV
NoafYpnEcB5s4m/vSVe6SU7dCK6eYec9f9wpT353ljhDUHq3EbmE4foNzJngh35d
AekCIQDhRQG5Li0Wj8TM4obOnnXUXf1jRv0UkzE9AHWLG5q3AwIhAPzSjpYUDjVW
MCUXgckTpKCuGwbJk7424Nb8bLzf3kllAiA5mUBgjfr/WtFSJdWcPQ4Zt9KTMNKD
EUO0ukpTwEIl6wIhAMbGqZK3zAAFdq8DD2jPx+UJXnh0rnOkZBzDtJ6/iN69AiEA
1Aq8MJgTaYsDQWyU/hDq5YkDJc9e9DSCvUIzqxQWMQE=
-----END RSA PRIVATE KEY-----`)
