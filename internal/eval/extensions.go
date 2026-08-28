package eval

import (
	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
	"github.com/out-lang/out/internal/stdlib"
)

func init() {
	boot()
}

// registry holds every Go module exposed to OUT (stdlib + extensions).
var registry *module.Registry

func boot() {
	registry = module.NewRegistry()
	stdlib.RegisterAll(registry)
}

// lookupModuleMember resolves "moduleName::member" through the Go extension
// registry. Returns a builtin function wrapper, or nil if not found.
func lookupModuleMember(objectName, member string) object.Object {
	m, ok := registry.Lookup(objectName)
	if !ok {
		return nil
	}
	fn, ok := m.Get(member)
	if !ok {
		return nil
	}
	return &object.BuiltinFunction{Fn: fn}
}
