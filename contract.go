package container

import "reflect"

type Interface interface {
	BindFunc(abs interface{}, register InstanceRegister, singleton bool)
	Bind(abstract, instance interface{}, singleton bool)
	Instance(abs interface{}) (instance reflect.Value, e error)
	InstanceFor(abs interface{}, out interface{}) error
	Invoke(f interface{}) ([]reflect.Value, error)
	MethodCall(abs interface{}, method string) ([]reflect.Value, error)
	RegisterSingleton(instance interface{})
	Register(instance interface{}, singleton bool)
	Singleton(abs interface{}, instance interface{})
	WithInjector(h Injector)
	InjectionWith(i InjectionFunc)
	InjectionRequire(requireAbs interface{}, i InjectionFunc)
	InjectionCondition(f ConditionInjectionFunc, i InjectionFunc)
	IsSingleton(abs interface{}) bool
	Bound(abs interface{}) bool
	GetRegisters() map[interface{}]InstanceRegister
}
