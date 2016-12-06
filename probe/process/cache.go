package process

import (
	"time"

	"github.com/armon/go-metrics"
	"github.com/coocood/freecache"

	"github.com/weaveworks/common/fs"
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
