// Package module provides the core extension system for OUT.
// It lets Go code register modules that OUT scripts can call,
// the same way Python extensions written in C are callable from Python.
package module

import "github.com/out-lang/out/internal/object"

// Func is the signature of a Go function exposed to OUT.
// It mirrors object.BuiltinFunction so modules can be built
// independently without importing the evaluator.
type Func func(args ...object.Object) object.Object

// Module is a named collection of Go functions callable from OUT
// as module::function(args).
type Module struct {
	Name    string
	Version string
	Desc    string
	// Funcs maps a member name to its Go implementation.
	Funcs map[string]Func
}

// New creates a blank module with the given name.
func New(name string) *Module {
	return &Module{Name: name, Funcs: make(map[string]Func)}
}

// Set registers a Go function under the given member name.
func (m *Module) Set(name string, fn Func) *Module {
	m.Funcs[name] = fn
	return m
}

// Get returns the Go function for a member, if present.
func (m *Module) Get(name string) (Func, bool) {
	fn, ok := m.Funcs[name]
	return fn, ok
}

// Registry holds every registered module, keyed by module name.
type Registry struct {
	modules map[string]*Module
}

// NewRegistry creates an empty module registry.
func NewRegistry() *Registry {
	return &Registry{modules: make(map[string]*Module)}
}

// Register adds a module to the registry.
func (r *Registry) Register(m *Module) {
	r.modules[m.Name] = m
}

// Lookup returns a module by name.
func (r *Registry) Lookup(name string) (*Module, bool) {
	m, ok := r.modules[name]
	return m, ok
}

// Names returns all registered module names (sorted for stable output).
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}
