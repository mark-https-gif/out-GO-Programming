package eval

import (
	"fmt"
	"sort"
	"strings"

	"github.com/out-lang/out/internal/env"
	"github.com/out-lang/out/internal/object"
)

func dispatchMethod(receiver object.Object, method string, args []object.Object) object.Object {
	switch r := receiver.(type) {
	case *object.Array:
		return dispatchArrayMethod(r, method, args)
	case *object.String:
		return dispatchStringMethod(r, method, args)
	case *object.Hash:
		return dispatchHashMethod(r, method, args)
	default:
		return newError("no methods on %s", receiver.Type())
	}
}

func callFunc(fn object.Object, receiver object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return NULL
	}
	outer, _ := function.Env.(*env.Environment)
	childEnv := env.NewEnclosed(outer)
	if len(function.Parameters) > 0 {
		childEnv.Set(function.Parameters[0].Value, receiver)
	}
	for i, arg := range args {
		if i+1 < len(function.Parameters) {
			childEnv.Set(function.Parameters[i+1].Value, arg)
		}
	}
	result := Eval(function.Body, childEnv)
	if rv, ok := result.(*object.ReturnValue); ok {
		return rv.Value
	}
	return result
}

func dispatchArrayMethod(arr *object.Array, method string, args []object.Object) object.Object {
	switch method {
	case "filter":
		if len(args) < 1 {
			return newError("filter requires 1 argument (function)")
		}
		var result []object.Object
		for _, elem := range arr.Elements {
			cond := callFunc(args[0], elem, nil)
			if isTruthy(cond) {
				result = append(result, elem)
			}
		}
		return &object.Array{Elements: result}

	case "map":
		if len(args) < 1 {
			return newError("map requires 1 argument (function)")
		}
		var result []object.Object
		for _, elem := range arr.Elements {
			mapped := callFunc(args[0], elem, nil)
			result = append(result, mapped)
		}
		return &object.Array{Elements: result}

	case "reduce":
		if len(args) < 2 {
			return newError("reduce requires 2 arguments (function, initial)")
		}
		acc := args[1]
		for _, elem := range arr.Elements {
			acc = callFunc(args[0], acc, []object.Object{elem})
		}
		return acc

	case "sort":
		elems := make([]object.Object, len(arr.Elements))
		copy(elems, arr.Elements)
		sort.Slice(elems, func(i, j int) bool {
			return elems[i].Inspect() < elems[j].Inspect()
		})
		return &object.Array{Elements: elems}

	case "sort_by":
		if len(args) < 1 {
			return newError("sort_by requires 1 argument (function)")
		}
		elems := make([]object.Object, len(arr.Elements))
		copy(elems, arr.Elements)
		sort.Slice(elems, func(i, j int) bool {
			vi := callFunc(args[0], elems[i], nil)
			vj := callFunc(args[0], elems[j], nil)
			return vi.Inspect() < vj.Inspect()
		})
		return &object.Array{Elements: elems}

	case "unique":
		seen := map[string]bool{}
		var result []object.Object
		for _, elem := range arr.Elements {
			key := elem.Inspect()
			if !seen[key] {
				seen[key] = true
				result = append(result, elem)
			}
		}
		return &object.Array{Elements: result}

	case "find":
		if len(args) < 1 {
			return newError("find requires 1 argument (function)")
		}
		for _, elem := range arr.Elements {
			cond := callFunc(args[0], elem, nil)
			if isTruthy(cond) {
				return elem
			}
		}
		return NULL

	case "any":
		if len(args) < 1 {
			return newError("any requires 1 argument (function)")
		}
		for _, elem := range arr.Elements {
			cond := callFunc(args[0], elem, nil)
			if isTruthy(cond) {
				return TRUE
			}
		}
		return FALSE

	case "all":
		if len(args) < 1 {
			return newError("all requires 1 argument (function)")
		}
		for _, elem := range arr.Elements {
			cond := callFunc(args[0], elem, nil)
			if !isTruthy(cond) {
				return FALSE
			}
		}
		return TRUE

	case "each", "forEach":
		if len(args) < 1 {
			return newError("each requires 1 argument (function)")
		}
		for _, elem := range arr.Elements {
			callFunc(args[0], elem, nil)
		}
		return NULL

	case "push":
		newElems := make([]object.Object, len(arr.Elements), len(arr.Elements)+len(args))
		copy(newElems, arr.Elements)
		newElems = append(newElems, args...)
		return &object.Array{Elements: newElems}

	case "pop":
		if len(arr.Elements) == 0 {
			return NULL
		}
		newElems := make([]object.Object, len(arr.Elements)-1)
		copy(newElems, arr.Elements[:len(arr.Elements)-1])
		return &object.Array{Elements: newElems}

	case "remove":
		if len(args) < 1 {
			return newError("remove requires 1 argument (index)")
		}
		idx, ok := args[0].(*object.Integer)
		if !ok {
			return newError("remove argument must be an integer")
		}
		i := int(idx.Value)
		if i < 0 || i >= len(arr.Elements) {
			return newError("index out of range")
		}
		newElems := make([]object.Object, 0, len(arr.Elements)-1)
		newElems = append(newElems, arr.Elements[:i]...)
		newElems = append(newElems, arr.Elements[i+1:]...)
		return &object.Array{Elements: newElems}

	case "contains":
		if len(args) < 1 {
			return newError("contains requires 1 argument")
		}
		for _, elem := range arr.Elements {
			if elem.Inspect() == args[0].Inspect() {
				return TRUE
			}
		}
		return FALSE

	case "index_of":
		if len(args) < 1 {
			return newError("index_of requires 1 argument")
		}
		for i, elem := range arr.Elements {
			if elem.Inspect() == args[0].Inspect() {
				return &object.Integer{Value: int64(i)}
			}
		}
		return &object.Integer{Value: -1}

	case "reverse":
		elems := make([]object.Object, len(arr.Elements))
		copy(elems, arr.Elements)
		for i, j := 0, len(elems)-1; i < j; i, j = i+1, j-1 {
			elems[i], elems[j] = elems[j], elems[i]
		}
		return &object.Array{Elements: elems}

	case "slice":
		start, end := 0, len(arr.Elements)
		if len(args) >= 1 {
			if s, ok := args[0].(*object.Integer); ok {
				start = int(s.Value)
			}
		}
		if len(args) >= 2 {
			if e, ok := args[1].(*object.Integer); ok {
				end = int(e.Value)
			}
		}
		if start < 0 {
			start = 0
		}
		if end > len(arr.Elements) {
			end = len(arr.Elements)
		}
		if start >= end {
			return &object.Array{Elements: []object.Object{}}
		}
		newElems := make([]object.Object, end-start)
		copy(newElems, arr.Elements[start:end])
		return &object.Array{Elements: newElems}

	case "concat":
		var all []object.Object
		all = append(all, arr.Elements...)
		for _, arg := range args {
			if a, ok := arg.(*object.Array); ok {
				all = append(all, a.Elements...)
			} else {
				all = append(all, arg)
			}
		}
		return &object.Array{Elements: all}

	default:
		return newError("unknown array method: %s", method)
	}
}

func dispatchStringMethod(str *object.String, method string, args []object.Object) object.Object {
	switch method {
	case "upper":
		return &object.String{Value: strings.ToUpper(str.Value)}
	case "lower":
		return &object.String{Value: strings.ToLower(str.Value)}
	case "split":
		sep := " "
		if len(args) >= 1 {
			if s, ok := args[0].(*object.String); ok {
				sep = s.Value
			}
		}
		parts := strings.Split(str.Value, sep)
		elems := make([]object.Object, len(parts))
		for i, p := range parts {
			elems[i] = &object.String{Value: p}
		}
		return &object.Array{Elements: elems}
	case "contains":
		if len(args) < 1 {
			return newError("contains requires 1 argument")
		}
		if s, ok := args[0].(*object.String); ok {
			return nativeBoolToBooleanObject(strings.Contains(str.Value, s.Value))
		}
		return FALSE
	case "replace":
		if len(args) < 2 {
			return newError("replace requires 2 arguments (old, new)")
		}
		old, _ := args[0].(*object.String)
		new, _ := args[1].(*object.String)
		if old == nil || new == nil {
			return newError("replace arguments must be strings")
		}
		return &object.String{Value: strings.ReplaceAll(str.Value, old.Value, new.Value)}
	case "trim":
		return &object.String{Value: strings.TrimSpace(str.Value)}
	case "starts_with":
		if len(args) < 1 {
			return newError("starts_with requires 1 argument")
		}
		if s, ok := args[0].(*object.String); ok {
			return nativeBoolToBooleanObject(strings.HasPrefix(str.Value, s.Value))
		}
		return FALSE
	case "ends_with":
		if len(args) < 1 {
			return newError("ends_with requires 1 argument")
		}
		if s, ok := args[0].(*object.String); ok {
			return nativeBoolToBooleanObject(strings.HasSuffix(str.Value, s.Value))
		}
		return FALSE
	case "repeat":
		n := 1
		if len(args) >= 1 {
			if i, ok := args[0].(*object.Integer); ok {
				n = int(i.Value)
			}
		}
		return &object.String{Value: strings.Repeat(str.Value, n)}
	case "reverse":
		runes := []rune(str.Value)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return &object.String{Value: string(runes)}
	case "charAt", "char_at":
		idx := 0
		if len(args) >= 1 {
			if i, ok := args[0].(*object.Integer); ok {
				idx = int(i.Value)
			}
		}
		runes := []rune(str.Value)
		if idx < 0 || idx >= len(runes) {
			return &object.String{Value: ""}
		}
		return &object.String{Value: string(runes[idx])}
	case "substring", "substr":
		start, end := 0, len([]rune(str.Value))
		if len(args) >= 1 {
			if s, ok := args[0].(*object.Integer); ok {
				start = int(s.Value)
			}
		}
		if len(args) >= 2 {
			if e, ok := args[1].(*object.Integer); ok {
				end = int(e.Value)
			}
		}
		runes := []rune(str.Value)
		if start < 0 {
			start = 0
		}
		if end > len(runes) {
			end = len(runes)
		}
		if start >= end {
			return &object.String{Value: ""}
		}
		return &object.String{Value: string(runes[start:end])}
	case "indexOf", "index_of":
		if len(args) < 1 {
			return newError("index_of requires 1 argument")
		}
		if s, ok := args[0].(*object.String); ok {
			idx := strings.Index(str.Value, s.Value)
			return &object.Integer{Value: int64(idx)}
		}
		return &object.Integer{Value: -1}
	case "toInt":
		n := 0
		fmt.Sscanf(str.Value, "%d", &n)
		return &object.Integer{Value: int64(n)}
	case "toFloat":
		f := 0.0
		fmt.Sscanf(str.Value, "%f", &f)
		return &object.Float{Value: f}
	case "toStr":
		return str
	case "isEmpty":
		return nativeBoolToBooleanObject(len(str.Value) == 0)
	default:
		return newError("unknown string method: %s", method)
	}
}

func dispatchHashMethod(hash *object.Hash, method string, args []object.Object) object.Object {
	switch method {
	case "keys":
		keys := []object.Object{}
		for _, pair := range hash.Pairs {
			keys = append(keys, pair.Key)
		}
		return &object.Array{Elements: keys}
	case "values":
		vals := []object.Object{}
		for _, pair := range hash.Pairs {
			vals = append(vals, pair.Value)
		}
		return &object.Array{Elements: vals}
	case "size", "len":
		return &object.Integer{Value: int64(len(hash.Pairs))}
	case "has", "contains":
		if len(args) < 1 {
			return newError("has requires 1 argument (key)")
		}
		keyHash := object.HashKey(args[0])
		_, ok := hash.Pairs[keyHash]
		return nativeBoolToBooleanObject(ok)
	case "entries":
		var result []object.Object
		for _, pair := range hash.Pairs {
			entry := &object.Array{Elements: []object.Object{pair.Key, pair.Value}}
			result = append(result, entry)
		}
		return &object.Array{Elements: result}
	case "filter":
		if len(args) < 1 {
			return newError("filter requires 1 argument (function)")
		}
		pairs := map[uint64]object.HashPair{}
		for k, pair := range hash.Pairs {
			cond := callFunc(args[0], pair.Value, []object.Object{pair.Key})
			if isTruthy(cond) {
				pairs[k] = pair
			}
		}
		return &object.Hash{Pairs: pairs}
	case "map":
		if len(args) < 1 {
			return newError("map requires 1 argument (function)")
		}
		pairs := map[uint64]object.HashPair{}
		for k, pair := range hash.Pairs {
			newVal := callFunc(args[0], pair.Value, []object.Object{pair.Key})
			pairs[k] = object.HashPair{Key: pair.Key, Value: newVal}
		}
		return &object.Hash{Pairs: pairs}
	case "each", "forEach":
		if len(args) < 1 {
			return newError("each requires 1 argument (function)")
		}
		for _, pair := range hash.Pairs {
			callFunc(args[0], pair.Value, []object.Object{pair.Key})
		}
		return NULL
	default:
		return newError("unknown hash method: %s", method)
	}
}
