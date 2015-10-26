package gcache_test

import (
	"fmt"
	"github.com/bluele/gcache"
	"testing"
)

func buildARCache(size int) gcache.Cache {
	return gcache.New(size).
		ARC().
		EvictedFunc(evictedFuncForARC).
		Build()
}

func buildLoadingARCache(size int) gcache.Cache {
	return gcache.New(size).
		ARC().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForARC).
		Build()
}

func evictedFuncForARC(key, value interface{}) {
	fmt.Printf("[ARC] Key:%v Value:%v will evicted.\n", key, value)
}

func TestARCGet(t *testing.T) {
	size := 1000
	gc := buildARCache(size)
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingARCGet(t *testing.T) {
	size := 1000
	numbers := 1000
	testGetCache(t, buildLoadingARCache(size), numbers)
}

func TestARCLength(t *testing.T) {
	gc := buildLoadingARCache(1000)
	gc.Get("test1")
	gc.Get("test2")
	length := gc.Len()
	expectedLength := 2
	if gc.Len() != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
}

func TestARCEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildLoadingARCache(cacheSize)

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}
