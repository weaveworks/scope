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
	"log"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// Wait sends attempts HEAD requests to url
func Wait(t *testing.T, url string, attempts int) {
	ctx := context.Background()

	for a := 0; a < attempts; a++ {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		log.Println("HEAD", url)
		_, err := ctxhttp.Head(ctx, nil, url)
		if err == nil {
			log.Println("Client is ready, beginning test...")
			return
		}

		sleepFor := 100 * time.Millisecond
		log.Println(err, "- sleeping for", sleepFor)
		time.Sleep(sleepFor)
	}

	t.Fatalf("could not talk to client in %d attempts", attempts)
}
