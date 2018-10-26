package render

import (
	"reflect"
	"runtime"
)

func functionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func typeName(i interface{}) string {
	if m, ok := i.(*memoise); ok {
		return "memoise:" + typeName(m.Renderer)
	}
	return reflect.TypeOf(i).String()
}
