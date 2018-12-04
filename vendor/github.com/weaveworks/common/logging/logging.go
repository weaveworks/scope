package logging

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/promrus"
)

// Setup configures logging output to stderr, sets the log level and sets the formatter.
func Setup(logLevel string) error {
	log.SetOutput(os.Stderr)
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("Error parsing log level: %v", err)
	}
	log.SetLevel(level)
	log.SetFormatter(&textFormatter{})
	hook, err := promrus.NewPrometheusHook() // Expose number of log messages as Prometheus metrics.
	if err != nil {
		return err
	}
	log.AddHook(hook)
	return nil
}
