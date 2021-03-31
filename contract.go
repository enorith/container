package container

import "reflect"

type Register func(c Interface) reflect.Value

type Interface interface {
	//BindFunc(abs interface{}, register Register, singleton bool)
	Bind(abstract, instance interface{}, singleton bool)
	Instance(abs interface{}, params ...interface{}) (instance reflect.Value, e error)
	InstanceFor(abs interface{}, out interface{}, params ...interface{}) error
	Invoke(f interface{}, params ...interface{}) ([]reflect.Value, error)
	MethodCall(abs interface{}, method string, params ...interface{}) ([]reflect.Value, error)
	RegisterSingleton(instance interface{})
	Register(instance interface{}, singleton bool)
}
