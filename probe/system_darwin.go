package main

import (
	"os/exec"
	"strconv"
	"strings"
)

func getLoads() (float64, float64, float64) {
	out, err := exec.Command("w").CombinedOutput()
	if err != nil {
		return -1, -1, -1
	}
	noCommas := strings.NewReplacer(",", "")
	firstLine := strings.Split(string(out), "\n")[0]
	toks := strings.Fields(firstLine)
	if len(toks) < 5 {
		return -1, -1, -1
	}
	one, err := strconv.ParseFloat(noCommas.Replace(toks[len(toks)-3]), 64)
	if err != nil {
		return -1, -1, -1
	}
	five, err := strconv.ParseFloat(noCommas.Replace(toks[len(toks)-2]), 64)
	if err != nil {
		return -1, -1, -1
	}
	fifteen, err := strconv.ParseFloat(noCommas.Replace(toks[len(toks)-1]), 64)
	if err != nil {
		return -1, -1, -1
	}
	return one, five, fifteen
}
