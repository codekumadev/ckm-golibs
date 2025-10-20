package containers

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

type Container map[reflect.Type]map[string]*providerInfo

func New() Container {
	return make(Container)
}

func (c Container) Register(resolver any) Container {
	c.bind("", resolver, false)
	return c
}

func (c Container) NamedRegister(name string, resolver any) Container {
	c.bind(name, resolver, false)
	return c
}

func (c Container) Singleton(resolver any) Container {
	c.bind("", resolver, true)
	return c
}

func (c Container) NamedSingleton(name string, resolver any) Container {
	c.bind(name, resolver, true)
	return c
}

func (c Container) Value(resolver any) Container {
	c.bind("", resolver, true)
	return c
}

func (c Container) NamedValue(name string, resolver any) Container {
	c.bind(name, resolver, true)
	return c
}

func (c Container) bind(name string, resolver any, isSingleton bool) {
	to := reflect.TypeOf(resolver)
	if to.Kind() != reflect.Func {
		if isSingleton {
			if _, exist := c[to]; !exist {
				c[to] = make(map[string]*providerInfo)
			}

			c[to][name] = &providerInfo{
				resolver:    nil,
				instance:    resolver,
				isSingleton: true,
			}

			return
		}

		panic(errors.New("resolver must be a function"))
	}

	var out reflect.Type
	if to.NumOut() > 0 {
		out = to.Out(0)
		if _, exist := c[to.Out(0)]; !exist {
			c[out] = make(map[string]*providerInfo)
		}
	} else {
		panic(errors.New("resolver must return instance"))
	}

	var resolverOnce *sync.Once
	if isSingleton {
		resolverOnce = &sync.Once{}
	}

	c[out][name] = &providerInfo{
		resolver:     resolver,
		instance:     nil,
		isSingleton:  isSingleton,
		resolverOnce: resolverOnce,
	}
}

func (c Container) Resolve(abstraction any) error {
	return c.NamedResolve(abstraction, "")
}

func (c Container) MustResolve(abstraction any) {
	if err := c.NamedResolve(abstraction, ""); err != nil {
		panic(err)
	}
}

func (c Container) MustNameResolve(abstraction any, name string) {
	if err := c.NamedResolve(abstraction, name); err != nil {
		panic(err)
	}
}

func (c Container) NamedResolve(abstraction any, name string) error {
	receiverType := reflect.TypeOf(abstraction)
	if receiverType == nil {
		return errors.New("invalid abstraction in container")
	}

	if receiverType.Kind() == reflect.Ptr {
		elem := receiverType.Elem()

		if pvd, exist := c[elem][name]; exist {
			if instance, err := pvd.make(); err == nil {
				reflect.ValueOf(abstraction).Elem().Set(reflect.ValueOf(instance))
				return nil
			} else {
				return err
			}
		}

		return errors.New("no resolver found for: " + elem.String())
	}

	return errors.New("invalid abstraction in container")
}

func (c Container) Inject(structure interface{}) error {
	receiverType := reflect.TypeOf(structure)
	if receiverType == nil {
		return errors.New("invalid struct in container")
	}

	if receiverType.Kind() == reflect.Ptr {
		elem := receiverType.Elem()
		if elem.Kind() == reflect.Struct {
			s := reflect.ValueOf(structure).Elem()

			for i := 0; i < s.NumField(); i++ {
				f := s.Field(i)

				if name, exist := s.Type().Field(i).Tag.Lookup("inject"); exist {
					if concrete, exist := c[f.Type()][name]; exist {
						instance, err := concrete.make()
						if err != nil {
							return err
						}

						ptr := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
						ptr.Set(reflect.ValueOf(instance))

						continue
					}

					return fmt.Errorf("container cannot make %v field", s.Type().Field(i).Name)
				}
			}

			return nil
		}
	}

	return errors.New("invalid struct in container")
}

func (c Container) Reset() {
	for k := range c {
		delete(c, k)
	}
}
