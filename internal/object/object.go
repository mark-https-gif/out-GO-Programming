package object

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/out-lang/out/internal/ast"
)

type ObjectType string

const (
	INTEGER_OBJ  = "INTEGER"
	FLOAT_OBJ    = "FLOAT"
	STRING_OBJ   = "STRING"
	BOOL_OBJ     = "BOOL"
	NULL_OBJ     = "NULL"
	RETURN_OBJ   = "RETURN_VALUE"
	ERROR_OBJ    = "ERROR"
	FUNCTION_OBJ = "FUNCTION"
	BUILTIN_OBJ  = "BUILTIN"
	ARRAY_OBJ    = "ARRAY"
	HASH_OBJ     = "HASH"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string {
	return fmt.Sprintf("%g", f.Value)
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOL_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
	Line    int
}

func (e *Error) Type() ObjectType    { return ERROR_OBJ }
func (e *Error) Inspect() string     { return e.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        EnvStore
}

type EnvStore interface {
	Get(name string) (Object, bool)
	Set(name string, val Object) Object
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out string
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out += "def"
	out += "("
	out += strings.Join(params, ", ")
	out += ") {\n"
	out += f.Body.String()
	out += "\n}"
	return out
}

type BuiltinFunction struct {
	Fn func(args ...Object) Object
}

func (b *BuiltinFunction) Type() ObjectType { return BUILTIN_OBJ }
func (b *BuiltinFunction) Inspect() string  { return "builtin function" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out string
	elems := []string{}
	for _, e := range a.Elements {
		elems = append(elems, e.Inspect())
	}
	out += "["
	out += strings.Join(elems, ", ")
	out += "]"
	return out
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[uint64]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out string
	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, pair.Key.Inspect()+": "+pair.Value.Inspect())
	}
	out += "{"
	out += strings.Join(pairs, ", ")
	out += "}"
	return out
}

func HashKey(obj Object) uint64 {
	switch o := obj.(type) {
	case *Integer:
		return uint64(o.Value)
	case *Boolean:
		if o.Value {
			return 1
		}
		return 0
	case *String:
		h := fnv.New64a()
		h.Write([]byte(o.Value))
		return h.Sum64()
	default:
		return 0
	}
}

func IsHashable(obj Object) bool {
	switch obj.(type) {
	case *Integer, *Boolean, *String:
		return true
	default:
		return false
	}
}
