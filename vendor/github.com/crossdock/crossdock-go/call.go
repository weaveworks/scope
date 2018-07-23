// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package crossdock

import (
	"encoding/json"
	"log"
	"net/url"
	"testing"
	"time"

	"github.com/crossdock/crossdock-go/require"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// Call makes a GET request to the client
func Call(t *testing.T, clientURL string, behavior string, args url.Values) {
	args.Set("behavior", behavior)
	u, err := url.Parse(clientURL)
	require.NoError(t, err, "failed to parse URL")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u.RawQuery = args.Encode()
	log.Println("GET", u.String())
	res, err := ctxhttp.Get(ctx, nil, u.String())

	require.NoError(t, err, "request %v failed", args)
	defer res.Body.Close()

	var results []Entry
	require.NoError(t, json.NewDecoder(res.Body).Decode(&results),
		"failed to decode response for %v", args)

	for _, result := range results {
		if result.Status() != Passed && result.Status() != Skipped {
			output := result.Output()

			delete(result, statusKey)
			delete(result, outputKey)

			t.Errorf("request %v failed: tags: %v; output: %s", args, result, output)
		}
	}
}
