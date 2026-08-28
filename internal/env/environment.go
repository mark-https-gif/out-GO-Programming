package env

import "github.com/out-lang/out/internal/object"

type Environment struct {
	store map[string]object.Object
	outer *Environment
}

func New() *Environment {
	return &Environment{store: make(map[string]object.Object)}
}

func NewEnclosed(outer *Environment) *Environment {
	env := New()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (object.Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val object.Object) object.Object {
	e.store[name] = val
	return val
}
