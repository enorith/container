package container_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/enorith/container"
	"github.com/enorith/supports/reflection"
)

type foo struct {
	name string
}

func TestContainer_Bind(t *testing.T) {
	bt := bindTable()
	c := container.New()
	for _, v := range bt {
		t.Run(v.name, func(t *testing.T) {
			c.Bind(v.abs, v.instance, v.singleton)

			if !c.Bound(v.abs) {
				t.Fatalf("bind failed of %v, instance %v", v.abs, v.instance)
			}
			if c.IsSingleton(v.abs) != v.singleton {
				t.Fatalf("bind singleton failed of %v, instance %v", v.abs, v.instance)
			}
		})
	}
}

func TestContainer_Instance(t *testing.T) {
	bt := bindTable()
	c := container.New()
	for _, v := range bt {
		t.Run(v.name, func(t *testing.T) {
			c.Bind(v.abs, v.instance, v.singleton)
			obj, e := c.Instance(v.abs)

			if e != nil {
				t.Fatal(e)
			}

			if !obj.IsValid() {
				t.Fatalf("instance of %v is invalid", v.abs)
			}
			if i, ok := obj.Interface().(*foo); ok {
				if i.name != v.name {
					t.Fatalf("instance of %v is invalid, object name %s != %s", v.abs, i.name, v.name)
				}
			} else {
				t.Fatalf("instance of %v is invalid, got object %v", v.abs, obj)
			}
		})
	}
}

func TestContainer_Invoke(t *testing.T) {
	c := container.New()
	t.Run("invoke func", func(t *testing.T) {
		outs, err := c.Invoke(funcBar)

		if err != nil {
			t.Fatalf("invoke func fail %s", err)
		}
		if b, ok := outs[0].Interface().(bool); ok {
			if !b {
				t.Fatalf("invoke func fail got %v", b)
			}
		} else {
			t.Fatalf("invoke func fail got %v", outs)
		}
	})

	t.Run("invoke func injection", func(t *testing.T) {
		c.BindFunc(&foo{}, func(c container.Interface) (interface{}, error) {
			return reflect.ValueOf(&foo{name: "test foo"}), nil
		}, false)

		outs, err := c.Invoke(funcBarInjection)

		if err != nil {
			t.Fatalf("invoke func injection fail %s", err)
		}

		if b, ok := outs[0].Interface().(string); ok {
			if b != "test foo" {
				t.Fatalf("invoke func injection fail got %v", b)
			}
		} else {
			t.Fatalf("invoke func injection fail got %v", outs)
		}
	})
}

func TestTypeString(t *testing.T) {
	tt := []struct {
		name string
		abs  interface{}
		str  string
	}{
		{"struct type", foo{}, "container_test.foo"},
		{"ptr type", &foo{}, "*container_test.foo"},
		{"struct type type", reflect.TypeOf(foo{}), "container_test.foo"},
		{"ptr type type", reflect.TypeOf(&foo{}), "*container_test.foo"},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			str := reflection.TypeString(v.abs)

			if str != v.str {
				t.Fatalf("type if %v expect string [%s], got [%s]", v.abs, v.str, str)
			}
		})
	}
}

func TestContainer_InstanceFor(t *testing.T) {
	c := container.New()

	c.BindFunc("foo", func(c container.Interface) (interface{}, error) {

		return reflect.ValueOf(&foo{"test name"}), nil
	}, false)

	var f foo

	c.InstanceFor("foo", &f)
	if f.name != "test name" {
		t.Fatal("instance failed")
	}
	t.Log(f)
}

func TestContainer_Clone(t *testing.T) {
	c := container.New()
	c.BindFunc("bar", func(c container.Interface) (interface{}, error) {
		return TypeBar{}, nil
	}, false)
	c.WithInjector(InitializeHandler{})
	for i := 0; i < 6; i++ {

		go func(c *container.Container, i int) {
			var f foo
			cc := c.Clone()
			cc.BindFunc("foo", func(c container.Interface) (interface{}, error) {
				return &foo{fmt.Sprintf("test name %d", i)}, nil
			}, true)

			e := cc.InstanceFor("foo", &f)
			if e != nil {
				fmt.Println(e)
			} else {
				fmt.Println("foo", f, i)
			}
			if cc.Bound("bar") {
				var b TypeBar
				e = cc.InstanceFor("bar", &b)
				if e != nil {
					fmt.Println(e)
				} else {
					fmt.Println("bar", b, i)
				}
			}
		}(c, i)
	}
}

type InitializeHandler struct {
}

func (i InitializeHandler) Injection(abs interface{}, last reflect.Value) (reflect.Value, error) {

	return reflect.ValueOf(foo{"test foo"}), nil
}

func (i InitializeHandler) When(abs interface{}) bool {

	str := reflection.TypeString(abs)

	return str == "container_test.foo"
}

func TestContainer_HandleInitialize(t *testing.T) {
	c := container.New()

	c.WithInjector(InitializeHandler{})

	i, e := c.Instance("container_test.foo")

	if e != nil {
		t.Fatal(e)
	}
	t.Log(i)
	if !i.IsValid() {
		t.Fatal("instance failed")
	}
}

func funcBar() bool {

	return true
}

func funcBarInjection(f *foo) string {
	return f.name
}

func bindTable() []struct {
	name      string
	abs       interface{}
	instance  interface{}
	singleton bool
} {
	typ := reflect.TypeOf(&foo{})

	return []struct {
		name      string
		abs       interface{}
		instance  interface{}
		singleton bool
	}{
		{"string abs", "foo_s", &foo{name: "string abs"}, false},
		{"string abs singleton", "foo_ss", &foo{name: "string abs singleton"}, true},
		{"string abs func", "foo_s_f", container.InstanceRegister(func(c container.Interface) (interface{}, error) {
			return reflect.ValueOf(&foo{name: "string abs func"}), nil
		}), false},
		{"type abs", typ, &foo{name: "type abs"}, false},
		{"object abs", &foo{}, &foo{name: "object abs"}, false},
	}
}

type BaseStruct struct {
	prop1 string
}

type BaseStruct2 struct {
	prop2 string
}

type testStruct struct {
	BaseStruct2
	BaseStruct
}

type testStructInjector1 struct {
	index int
}

func (t *testStructInjector1) Injection(abs interface{}, last reflect.Value) (reflect.Value, error) {
	sv := reflection.StructValue(last)
	sv.Field(t.index).Set(reflect.ValueOf(BaseStruct{prop1: "base prop1"}))

	return last, nil
}

func (t *testStructInjector1) When(abs interface{}) bool {
	t.index = reflection.SubStructOf(abs, BaseStruct{})
	return t.index > -1
}

type testStructInjector2 struct {
	index int
}

func (t *testStructInjector2) Injection(abs interface{}, last reflect.Value) (reflect.Value, error) {
	sv := reflection.StructValue(last)
	sv.Field(t.index).Set(reflect.ValueOf(BaseStruct2{prop2: "base prop2"}))

	return last, nil
}

func (t *testStructInjector2) When(abs interface{}) bool {
	t.index = reflection.SubStructOf(abs, BaseStruct2{})
	return t.index > -1
}

func getProp(ts testStruct) string {
	return fmt.Sprintf("prop1: %s, prop2: %s", ts.prop1, ts.prop2)
}

func TestContainer_InjectionChain(t *testing.T) {
	c := container.New()
	c.WithInjector(&testStructInjector1{})
	c.WithInjector(&testStructInjector2{})
	v, e := c.Invoke(getProp)
	if e != nil {
		t.Fatal(e)
	}
	if s, ok := v[0].Interface().(string); ok {
		t.Log("invoke result", s)
	} else {
		t.Fatalf("invoke failed")
	}
}

type TypeBar struct {
	a int
}

func (tf TypeBar) GetName() string {
	return fmt.Sprintf("bar %d", tf.a)
}

func CallTypeBar(tb TypeBar) string {
	return tb.GetName()
}

func TestContainer_InstanceConstructInject(t *testing.T) {
	c := container.New()
	// c.BindFunc(TypeBar{}, func(c container.Interface) (interface{}, error) {
	// 	return TypeBar{a: 42}, nil
	// }, false)

	res, e := c.Invoke(CallTypeBar)
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("invoke result %v", res)
}

func TestContainer_Singleton(t *testing.T) {
	c := container.New()
	c.BindFunc(&Single{}, func(c container.Interface) (interface{}, error) {
		return &Single{}, nil
	}, true)
	v1, e := c.Invoke(Sfunc1)
	if e != nil {
		t.Fatal(e)
	}

	v2, e := c.Invoke(Sfunc2)
	if e != nil {
		t.Fatal(e)
	}

	if v1[0].Interface().(string) != v2[0].Interface().(string) {
		t.Log("singleton value unchanged")
		t.Fail()
	}
}

type Single struct {
	Name string
}

func Sfunc1(s *Single) string {
	s.Name = "123"
	return s.Name
}

func Sfunc2(s *Single) string {
	return s.Name
}
