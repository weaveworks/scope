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

import "sort"

// Combinations takes a map from axis name to list of values for that axis and
// returns a collection of entries which contain all combinations of each axis
// value with every other axis' values.
func Combinations(axes map[string][]string) []map[string]string {
	// sort the names to get deterministic ordering
	names := make([]string, 0, len(axes))
	for name := range axes {
		names = append(names, name)
	}
	sort.Strings(names)

	var xs []axis
	for _, name := range names {
		xs = append(xs, axis{name, axes[name]})
	}
	return axisCombinations(xs)
}

func axisCombinations(axes []axis) []map[string]string {
	if len(axes) == 0 {
		return nil
	}

	if len(axes) == 1 {
		return axes[0].entries()
	}

	var entries []map[string]string
	for _, remaining := range axisCombinations(axes[1:]) {
		for _, entry := range axes[0].entries() {
			for k, v := range remaining {
				entry[k] = v
			}
			entries = append(entries, entry)
		}
	}

	return entries
}

type axis struct {
	name   string
	values []string
}

func (x axis) entries() []map[string]string {
	items := make([]map[string]string, len(x.values))
	for i, value := range x.values {
		items[i] = map[string]string{x.name: value}
	}
	return items
}
