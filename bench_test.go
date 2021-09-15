package container_test

import (
	"testing"

	"github.com/enorith/container"
)

type ObjectFoo struct {
	Data string
}

func Benchmark_Instance(b *testing.B) {
	c := container.New()
	b.ReportAllocs()
	c.BindFunc(ObjectFoo{}, func(c container.Interface) (interface{}, error) {
		return ObjectFoo{Data: "foo"}, nil
	}, true)
	for i := 0; i < b.N; i++ {
		_, e := c.Instance(ObjectFoo{})
		if e != nil {
			b.Error(e)
			b.Fail()
		}
		// _ = v.Interface().(ObjectFoo)
	}
}

func Benchmark_CloneInstance(b *testing.B) {
	c := container.New()
	b.ReportAllocs()
	c.BindFunc(ObjectFoo{}, func(c container.Interface) (interface{}, error) {
		return ObjectFoo{Data: "foo"}, nil
	}, false)

	for i := 0; i < b.N; i++ {
		_, e := c.Clone().Instance(ObjectFoo{})
		if e != nil {
			b.Error(e)
			b.Fail()
		}
		// _ = v.Interface().(ObjectFoo)
	}
}

func Benchmark_Invoke(b *testing.B) {
	c := container.New()
	b.ReportAllocs()
	c.BindFunc(ObjectFoo{}, func(c container.Interface) (interface{}, error) {
		return ObjectFoo{Data: "foo"}, nil
	}, true)
	fn := func(ObjectFoo) {

	}
	for i := 0; i < b.N; i++ {
		c.Invoke(fn)
		// _ = v.Interface().(ObjectFoo)
	}
}

func Benchmark_InvokeNative(b *testing.B) {
	c := container.New()
	b.ReportAllocs()
	c.BindFunc(ObjectFoo{}, func(c container.Interface) (interface{}, error) {
		return ObjectFoo{Data: "foo"}, nil
	}, false)

	fn := func(ObjectFoo) {

	}
	for i := 0; i < b.N; i++ {
		var f ObjectFoo
		c.InstanceFor(ObjectFoo{}, &f)

		fn(f)
		// _ = v.Interface().(ObjectFoo)
	}
}
