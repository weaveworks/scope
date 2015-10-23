package gcache_test

import (
	"fmt"
	"github.com/bluele/gcache"
	"testing"
	"time"
)

func evictedFuncForLFU(key, value interface{}) {
	fmt.Printf("[LFU] Key:%v Value:%v will evicted.\n", key, value)
}

func buildLFUCache(size int) gcache.Cache {
	return gcache.New(size).
		LFU().
		EvictedFunc(evictedFuncForLFU).
		Expiration(time.Second).
		Build()
}

func buildLoadingLFUCache(size int, loader gcache.LoaderFunc) gcache.Cache {
	return gcache.New(size).
		LFU().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForLFU).
		Expiration(time.Second).
		Build()
}

func TestLFUGet(t *testing.T) {
	size := 1000
	numbers := 1000

	gc := buildLoadingLFUCache(size, loader)
	testSetCache(t, gc, numbers)
	testGetCache(t, gc, numbers)
}

func TestLoadingLFUGet(t *testing.T) {
	size := 1000
	numbers := 1000

	gc := buildLoadingLFUCache(size, loader)
	testGetCache(t, gc, numbers)
}

func TestLFULength(t *testing.T) {
	gc := buildLoadingLFUCache(1000, loader)
	gc.Get("test1")
	gc.Get("test2")
	length := gc.Len()
	expectedLength := 2
	if gc.Len() != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
}

func TestLFUEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildLoadingLFUCache(cacheSize, loader)

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}
