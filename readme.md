# golang implement of IoC-Container

## basic usage

```go
package main
import (
    "fmt"
    "github.com/enorith/container"
    "reflect"
)

type Foo struct{
    name string
}

func main(){
    c := container.New()
 
    // bind
    c.BindFunc(Foo{}, func(c *container.Container) reflect.Value {
        return reflect.ValueOf(Foo{"foo"})
    }, false)
    // get instance
    v, _ := c.Instance(Foo{})
    fmt.Println(v.Interface().(Foo).name)
    var f Foo
    // get instance
    c.InstanceFor(Foo{}, &f)
    fmt.Println(f.name)
    // bind with name
    c.BindFunc("foo", func(c *container.Container) reflect.Value {
        return reflect.ValueOf(Foo{"foo"})
    }, false)
    v2, _ := c.Instance("foo")
    fmt.Println(v2.Interface().(Foo).name)
}
```