package container_test

import (
	"fmt"
	"reflect"
	"testing"
)

type FooInterface interface {
}
type FooType struct {
}

func TestHashingType(t *testing.T) {
	m := make(map[interface{}]int)
	m["foo"] = 1
	var (
		ft  FooType
		ftp *FooType
	)
	m[reflect.TypeOf(ft)] = 2
	m[reflect.TypeOf(ftp)] = 3

	fmt.Println(m["foo"], m[reflect.TypeOf(ft)], m[reflect.TypeOf(ftp)])
}
