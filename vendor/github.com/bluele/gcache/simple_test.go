package gcache_test

import (
	"fmt"
	gcache "github.com/bluele/gcache"
	"testing"
)

func buildSimpleCache(size int) gcache.Cache {
	return gcache.New(size).
		Simple().
		EvictedFunc(evictedFuncForSimple).
		Build()
}

func buildLoadingSimpleCache(size int, loader gcache.LoaderFunc) gcache.Cache {
	return gcache.New(size).
		LoaderFunc(loader).
		Simple().
		EvictedFunc(evictedFuncForSimple).
		Build()
}

func evictedFuncForSimple(key, value interface{}) {
	fmt.Printf("[Simple] Key:%v Value:%v will evicted.\n", key, value)
}

func TestSimpleGet(t *testing.T) {
	size := 1000
	gc := buildSimpleCache(size)
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingSimpleGet(t *testing.T) {
	size := 1000
	numbers := 1000
	testGetCache(t, buildLoadingSimpleCache(size, loader), numbers)
}

func TestSimpleLength(t *testing.T) {
	gc := buildLoadingSimpleCache(1000, loader)
	gc.Get("test1")
	gc.Get("test2")
	length := gc.Len()
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
}

func TestSimpleEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildLoadingSimpleCache(cacheSize, loader)

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}
