package container_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/enorith/container"
)

type FooBar struct {
	Name string
}

type TestG[T interface{}] struct {
	Obj T
}

func (t *TestG[T]) Get() T {
	return t.Obj
}

func TestHashingType(t *testing.T) {
	var s string

	v := reflect.New(reflect.TypeOf(s)).Elem()

	t.Log(v.IsValid())
}

func TestTypeGenerics(t *testing.T) {
	ioc := container.New()
	ioc.BindFunc(FooBar{}, func(c container.Interface) (interface{}, error) {
		return FooBar{Name: "hahahahha"}, nil
	}, true)

	_, e := ioc.Invoke(Handler)
	if e != nil {
		t.Fatal(e)
	}
}

func Handler(tg TestG[FooBar]) {
	fmt.Println(tg.Get().Name)
}
