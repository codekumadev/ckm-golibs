package containers

import (
	"errors"
	"reflect"
	"sync"
)

type providerInfo struct {
	resolver     any
	instance     any
	isSingleton  bool
	resolverOnce *sync.Once
}

func (p *providerInfo) make() (any, error) {
	if p.isSingleton {
		if p.instance != nil {
			return p.instance, nil
		}

		var err error
		p.resolverOnce.Do(func() {
			p.instance, err = p.create()
		})

		return p.instance, err
	}

	instance, err := p.create()

	return instance, err
}

func (p *providerInfo) create() (any, error) {
	rets := reflect.ValueOf(p.resolver).Call(nil)
	if len(rets) == 1 || len(rets) == 2 {
		if len(rets) == 2 && rets[1].CanInterface() {
			if //goland:noinspection GoTypeAssertionOnErrors
			err, ok := rets[1].Interface().(error); ok {
				return rets[0].Interface(), err
			}
		}
		return rets[0].Interface(), nil
	}

	return nil, errors.New("resolver function signature is invalid")
}
