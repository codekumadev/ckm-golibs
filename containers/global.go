package containers

var global = New()

func Register(resolver any) Container {
	global.bind("", resolver, false)
	return global
}

func NamedRegister(name string, resolver any) Container {
	global.bind(name, resolver, false)
	return global
}

func Singleton(resolver any) Container {
	global.bind("", resolver, true)
	return global
}

func NamedSingleton(name string, resolver any) Container {
	global.bind(name, resolver, true)
	return global
}

func Value(resolver any) Container {
	global.bind("", resolver, true)
	return global
}

func NamedValue(name string, resolver any) Container {
	global.bind(name, resolver, true)
	return global
}

func Resolve(abstraction any) error {
	return global.NamedResolve(abstraction, "")
}

func MustResolve(abstraction any) {
	err := global.NamedResolve(abstraction, "")
	if err != nil {
		panic(err)
	}
}

func MustNamedResolve(abstraction any, name string) {
	err := global.NamedResolve(abstraction, name)
	if err != nil {
		panic(err)
	}
}

func NamedResolve(abstraction any, name string) error {
	return global.NamedResolve(abstraction, name)
}

func Inject(structure interface{}) error {
	return global.Inject(structure)
}

func Reset() {
	global.Reset()
}
