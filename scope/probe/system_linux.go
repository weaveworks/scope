package main

import (
	"io/ioutil"
	"strconv"
	"strings"
)

func getLoads() (float64, float64, float64) {
	buf, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return -1, -1, -1
	}
	toks := strings.Fields(string(buf))
	if len(toks) < 3 {
		return -1, -1, -1
	}
	one, err := strconv.ParseFloat(toks[0], 64)
	if err != nil {
		return -1, -1, -1
	}
	five, err := strconv.ParseFloat(toks[1], 64)
	if err != nil {
		return -1, -1, -1
	}
	fifteen, err := strconv.ParseFloat(toks[2], 64)
	if err != nil {
		return -1, -1, -1
	}
	return one, five, fifteen
}
