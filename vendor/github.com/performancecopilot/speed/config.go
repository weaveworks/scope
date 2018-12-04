package speed

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
)

// rootPath stores path to the pcp root installation
var rootPath string

// confPath stores path to pcp.conf
var confPath string

// config stores the configuration as defined in current PCP environment
var config map[string]string

// pat stores a valid key-value pattern line
var pat = "([A-Z0-9_]+)=(.*)"

// initConfig initializes the config constants
func initConfig() error {
	re, _ := regexp.Compile(pat)

	r, ok := os.LookupEnv("PCP_DIR")
	if !ok {
		r = "/"
	}
	rootPath = r

	c, ok := os.LookupEnv("PCP_CONF")
	if !ok {
		c = filepath.Join(rootPath, "etc", "pcp.conf")
	}
	confPath = c

	f, err := os.Open(confPath)
	if err != nil {
		return err
	}

	// if we reach at this point, it means we have a valid config
	// that can be read, so we can make the map non-nil
	config = make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		if re.MatchString(t) {
			matches := re.FindStringSubmatch(t)
			config[matches[1]] = matches[2]
		}
	}

	return nil
}
