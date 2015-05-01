package main

import "os"

func hostname() string {
	if hostname := os.Getenv("CELLO_HOSTNAME"); hostname != "" {
		return hostname
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "(unknown)"
	}
	return hostname
}
