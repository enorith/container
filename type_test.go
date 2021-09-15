package container_test

import (
	"reflect"
	"testing"
)

func TestHashingType(t *testing.T) {
	var s string

	v := reflect.New(reflect.TypeOf(s)).Elem()

	t.Log(v.IsValid())
}
