package process

import (
	"strconv"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/coocood/freecache"

	"github.com/weaveworks/scope/common/fs"
)

const (
	generalTimeout = 30 // seconds
	statsTimeout   = 10 //seconds
)

var (
	hitMetricsKey  = []string{"process", "cache", "hit"}
	missMetricsKey = []string{"process", "cache", "miss"}
)

var fileCache = freecache.NewCache(1024 * 16)

type entry struct {
	buf []byte
	err error
	ts  time.Time
}

func cachedReadFile(path string) ([]byte, error) {
	key := []byte(path)
	if v, err := fileCache.Get(key); err == nil {
		metrics.IncrCounter(hitMetricsKey, 1.0)
		return v, nil
	}

	buf, err := fs.ReadFile(path)
	fileCache.Set(key, buf, generalTimeout)
	metrics.IncrCounter(missMetricsKey, 1.0)
	return buf, err
}

// we cache the stats, but for a shorter period
func readStats(path string) (int, int, error) {
	var (
		key = []byte(path)
		buf []byte
		err error
	)
	if buf, err = fileCache.Get(key); err == nil {
		metrics.IncrCounter(hitMetricsKey, 1.0)
	} else {
		buf, err = fs.ReadFile(path)
		if err != nil {
			return -1, -1, err
		}
		fileCache.Set(key, buf, statsTimeout)
		metrics.IncrCounter(missMetricsKey, 1.0)
	}
	splits := strings.Fields(string(buf))
	ppid, err := strconv.Atoi(splits[3])
	if err != nil {
		return -1, -1, err
	}
	threads, err := strconv.Atoi(splits[19])
	if err != nil {
		return -1, -1, err
	}
	return ppid, threads, nil
}
