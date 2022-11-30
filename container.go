package container

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/enorith/supports/reflection"
)

// InstanceRegister register instance for container
type InstanceRegister func(c Interface) (interface{}, error)

// Injector interface for conditional initializer
type Injector interface {
	Injection(abs interface{}, last reflect.Value) (reflect.Value, error)
	When(abs interface{}) bool
}

// InjectionFunc injection function
type InjectionFunc func(abs interface{}, last reflect.Value) (reflect.Value, error)

type injectionChain []InjectionFunc

// ConditionInjectionFunc  conditional injection function
type ConditionInjectionFunc func(abs interface{}) bool

type UnregisterdAbstractError struct {
	Abs string
}

func (u UnregisterdAbstractError) Error() string {
	return fmt.Sprintf("unregisterd abstract [%s]", u.Abs)
}

type InjectPassedError struct {
}

func (InjectPassedError) Error() string {
	return "abstract passed"
}

func (ic injectionChain) do(abs interface{}) (va reflect.Value, initilaized bool, e error) {
	t := reflection.TypeOf(abs)
	ts := reflection.StructType(abs)
	va = reflect.New(ts)
	for _, v := range ic {
		va, e = v(abs, va)
		if e != nil {
			if _, ok := e.(InjectPassedError); ok {
				e = nil
				continue
			} else {
				return
			}
		}
		initilaized = true
	}

	if t.Kind() != reflect.Ptr && va.Kind() == reflect.Ptr {
		va = va.Elem()
	}

	return
}

func conditionInjectionFunc(requireAbs interface{}, i InjectionFunc) InjectionFunc {

	return func(abs interface{}, last reflect.Value) (reflect.Value, error) {
		if f, ok := requireAbs.(ConditionInjectionFunc); ok {
			if f(abs) {
				return i(abs, last)
			}
		} else {
			as := reflection.StructType(abs)
			rs := reflection.StructType(requireAbs)

			if as == rs {
				return i(abs, last)
			}
		}

		return last, InjectPassedError{}
	}
}

//Container is a IoC-Container
type Container struct {
	parent *Container
	mu     sync.RWMutex

	registers map[interface{}]InstanceRegister

	singletons map[interface{}]bool

	resolved map[interface{}]reflect.Value

	injectionChain injectionChain
}

func (c *Container) getRegister(abs interface{}) (InstanceRegister, bool) {
	if ir, ok := c.registers[abs]; ok {
		return ir, ok
	}

	if c.parent != nil {
		ir, ok := c.parent.registers[abs]
		return ir, ok
	}

	return nil, false
}

func (c *Container) isSingleton(abs interface{}) (bool, bool) {
	is, ok := c.singletons[abs]
	if ok {
		return is, ok
	}

	if c.parent != nil {
		is, ok := c.parent.singletons[abs]
		return is, ok
	}

	return false, false
}

func (c *Container) getResolved(abs interface{}) (reflect.Value, bool) {
	rv, ok := c.resolved[abs]
	if ok {
		return rv, ok
	}

	if c.parent != nil {
		rv, ok := c.parent.resolved[abs]
		return rv, ok
	}

	return reflect.Value{}, false
}

func (c *Container) getChain() injectionChain {
	if c.parent != nil {
		c.injectionChain = append(c.parent.injectionChain, c.injectionChain...)
	}

	return c.injectionChain
}

func (c *Container) WithInjector(h Injector) {
	c.InjectionWith(conditionInjectionFunc(ConditionInjectionFunc(h.When), h.Injection))
}

func (c *Container) InjectionWith(i InjectionFunc) {
	c.injectionChain = append(c.injectionChain, i)
}

func (c *Container) InjectionRequire(requireAbs interface{}, i InjectionFunc) {
	c.InjectionWith(conditionInjectionFunc(requireAbs, i))
}

func (c *Container) InjectionCondition(f ConditionInjectionFunc, i InjectionFunc) {
	c.InjectionWith(conditionInjectionFunc(f, i))
}

// Bind: pre-bind abstract to container
// 	Abstract could be string,reflect.Type,struct or pointer
// 	Instance could be reflect.Value, struct, pointer or InstanceRegister
func (c *Container) Bind(abstract, instance interface{}, singleton bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if instance == nil {
		instance = abstract
	}
	key := absKey(abstract)

	c.registers[key] = c.getResolver(instance)
	c.singletons[key] = singleton
}

func (c *Container) BindFunc(abs interface{}, register InstanceRegister, singleton bool) {
	c.Bind(abs, register, singleton)
}

func (c *Container) Register(instance interface{}, singleton bool) {
	c.Bind(instance, nil, singleton)
}

func (c *Container) RegisterSingleton(instance interface{}) {
	c.Bind(instance, nil, true)
}

func (c *Container) Singleton(abs interface{}, instance interface{}) {
	c.Bind(abs, instance, true)
}

func (c *Container) IsSingleton(abs interface{}) bool {
	typ := absKey(abs)

	if v, ok := c.singletons[typ]; ok {
		return v
	}
	return false
}

func (c *Container) MethodCall(abs interface{}, method string) ([]reflect.Value, error) {
	instance, e := c.Instance(abs)
	if e != nil {
		return nil, e
	}

	if !instance.IsValid() {
		return nil, fmt.Errorf("invalid method for type %v method [%s]", reflect.TypeOf(abs), method)
	}

	m := instance.MethodByName(method)

	return c.Invoke(m)
}

func (c *Container) getResolver(instance interface{}) InstanceRegister {
	var r InstanceRegister

	if t, ok := instance.(reflect.Type); ok {
		r = func(c Interface) (interface{}, error) {
			return reflect.New(t).Elem(), nil
		}
	} else if t, ok := instance.(InstanceRegister); ok {
		r = t
	} else {
		r = func(c Interface) (interface{}, error) {
			return instance, nil
		}
	}

	return r
}

// Instance return reflect.Value of gaving abstract
func (c *Container) Instance(abs interface{}) (instance reflect.Value, e error) {
	defer func() {
		if x := recover(); x != nil {
			instance = reflect.Value{}
			if err, ok := x.(error); ok {
				e = err
			}
			if s, ok := x.(string); ok {
				e = errors.New(s)
			}
		}
	}()

	// whether this abstract is constructed
	// constructed := false

	fallback := func() {
		var va reflect.Value
		va, _, e = c.getChain().do(abs)

		if va.IsValid() {
			instance = va
			// constructed = true
		} else {
			instance = reflect.Value{}
		}
	}

	resolve := func(abs interface{}) {
		resolved, err := c.getResolve(abs)
		if _, ok := err.(UnregisterdAbstractError); !ok {
			e = err
		}

		if resolved.IsValid() {
			instance = resolved
			// constructed = true
		} else if e == nil {
			fallback()
		}
	}
	// c.mu.RLock()
	// defer c.mu.RUnlock()
	resolve(abs)

	if e == nil {
		// construct injection
		switch instance.Kind() {
		case reflect.Ptr, reflect.Struct:
			ind := reflect.Indirect(instance)
			for i := 0; i < ind.NumField(); i++ {
				fv := ind.Field(i)
				if fv.IsZero() && fv.CanSet() &&
					(fv.Kind() == reflect.Ptr ||
						fv.Kind() == reflect.Struct ||
						fv.Kind() == reflect.Interface) {
					v, e := c.Instance(fv.Type())
					if e == nil {
						fv.Set(v)
					}
				}
			}
		}
	}

	return instance, e
}

func (c *Container) InstanceFor(abs interface{}, out interface{}) error {
	v, e := c.Instance(abs)
	if e != nil {
		return e
	}

	o := reflect.ValueOf(out)

	if !o.IsValid() {
		return fmt.Errorf("instance for abstact [%s] failed", reflection.TypeString(abs))
	}

	if o.Kind() == reflect.Ptr {
		o = o.Elem()
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	o.Set(v)

	return nil
}

func (c *Container) Invoke(f interface{}) ([]reflect.Value, error) {
	var t reflect.Type
	var fun reflect.Value

	if typ, ok := f.(reflect.Value); ok {
		t = typ.Type()
		fun = typ
	} else {
		t = reflect.TypeOf(f)
		fun = reflect.ValueOf(f)
	}
	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("invoke failed, type if %v is invalid, expect func", t)
	}

	var in = make([]reflect.Value, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		param, e := c.Instance(argType)
		if e != nil {
			return nil, e
		}

		if !param.IsValid() {
			return nil, fmt.Errorf("inject %v failed, parameter [%d] of type %v is invalid", t.String(), i, argType)
		}
		in[i] = param
	}

	return fun.Call(in), nil
}

func (c *Container) Clone() Interface {
	return &Container{
		parent:     c,
		registers:  make(map[interface{}]InstanceRegister),
		singletons: map[interface{}]bool{},
		resolved:   make(map[interface{}]reflect.Value),
	}
}

func (c *Container) GetRegisters() map[interface{}]InstanceRegister {
	return c.registers
}

func (c *Container) getResolve(abs interface{}) (reflect.Value, error) {
	key := absKey(abs)

	if resolved, ok := c.getResolved(key); ok {

		return resolved, nil
	}
	if resolver, o := c.getRegister(key); o {
		instance, e := resolver(c)
		if e != nil {
			return reflect.Value{}, e
		}
		instanceVal := reflection.ValueOf(instance)

		if _, r := c.resolved[key]; !r {
			if is, _ := c.isSingleton(key); is {
				c.resolved[key] = instanceVal
			}
		}

		return instanceVal, nil
	}

	return reflect.Value{}, UnregisterdAbstractError{Abs: reflection.TypeString(abs)}
}

func (c *Container) Bound(abs interface{}) bool {
	_, o := c.getRegister(absKey(abs))

	return o
}

func New() *Container {
	c := &Container{
		registers:  make(map[interface{}]InstanceRegister),
		singletons: make(map[interface{}]bool),
		resolved:   make(map[interface{}]reflect.Value),
		mu:         sync.RWMutex{},
	}

	return c
}

func absKey(abs interface{}) interface{} {
	if s, ok := abs.(string); ok {
		return s
	}
	return reflection.TypeOf(abs)
}
