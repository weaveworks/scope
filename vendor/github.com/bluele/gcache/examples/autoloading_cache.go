package main

import (
	"fmt"
	"github.com/bluele/gcache"
)

func main() {
	gc := gcache.New(10).
		LFU().
		LoaderFunc(func(key interface{}) (interface{}, error) {
		return fmt.Sprintf("%v-value", key), nil
	}).
		Build()

	v, err := gc.Get("key")
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}
