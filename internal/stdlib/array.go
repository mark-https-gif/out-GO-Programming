package stdlib

import (
	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

var callFunc func(fn object.Object, args []object.Object) object.Object

func SetFuncCaller(fn func(object.Object, []object.Object) object.Object) {
	callFunc = fn
}

func arrayModule() *module.Module {
	m := module.New("array")

	m.Set("filter", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("array::filter expects 2 arguments (array, function)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("array::filter expects ARRAY as first argument")
		}
		fn := args[1]
		if callFunc == nil {
			return errObj("array::filter: function caller not initialized")
		}
		result := make([]object.Object, 0)
		for _, elem := range arr.Elements {
			ret := callFunc(fn, []object.Object{elem})
			if ret != nil {
				if b, ok := ret.(*object.Boolean); ok && b.Value {
					result = append(result, elem)
				}
			}
		}
		return &object.Array{Elements: result}
	})

	m.Set("map", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("array::map expects 2 arguments (array, function)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("array::map expects ARRAY as first argument")
		}
		fn := args[1]
		if callFunc == nil {
			return errObj("array::map: function caller not initialized")
		}
		result := make([]object.Object, len(arr.Elements))
		for i, elem := range arr.Elements {
			ret := callFunc(fn, []object.Object{elem})
			if ret != nil {
				result[i] = ret
			} else {
				result[i] = elem
			}
		}
		return &object.Array{Elements: result}
	})

	m.Set("reduce", func(args ...object.Object) object.Object {
		if len(args) != 3 {
			return errObj("array::reduce expects 3 arguments (array, function, initial)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("array::reduce expects ARRAY")
		}
		fn := args[1]
		if callFunc == nil {
			return errObj("array::reduce: function caller not initialized")
		}
		acc := args[2]
		for _, elem := range arr.Elements {
			acc = callFunc(fn, []object.Object{acc, elem})
			if acc == nil {
				acc = &object.Null{}
			}
		}
		return acc
	})

	m.Set("sort", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("array::sort expects 1 argument (array)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("array::sort expects ARRAY")
		}
		elems := make([]object.Object, len(arr.Elements))
		copy(elems, arr.Elements)
		for i := 0; i < len(elems)-1; i++ {
			for j := i + 1; j < len(elems); j++ {
				if elems[i].Inspect() > elems[j].Inspect() {
					elems[i], elems[j] = elems[j], elems[i]
				}
			}
		}
		return &object.Array{Elements: elems}
	})

	m.Set("unique", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("array::unique expects 1 argument (array)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("array::unique expects ARRAY")
		}
		seen := make(map[string]bool)
		result := make([]object.Object, 0)
		for _, elem := range arr.Elements {
			key := elem.Inspect()
			if !seen[key] {
				seen[key] = true
				result = append(result, elem)
			}
		}
		return &object.Array{Elements: result}
	})

	m.Set("find", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("array::find expects 2 arguments (array, function)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("array::find expects ARRAY")
		}
		fn := args[1]
		if callFunc == nil {
			return errObj("array::find: function caller not initialized")
		}
		for _, elem := range arr.Elements {
			ret := callFunc(fn, []object.Object{elem})
			if ret != nil {
				if b, ok := ret.(*object.Boolean); ok && b.Value {
					return elem
				}
			}
		}
		return &object.Null{}
	})

	m.Set("any", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("array::any expects 2 arguments (array, function)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("array::any expects ARRAY")
		}
		fn := args[1]
		if callFunc == nil {
			return errObj("array::any: function caller not initialized")
		}
		for _, elem := range arr.Elements {
			ret := callFunc(fn, []object.Object{elem})
			if ret != nil {
				if b, ok := ret.(*object.Boolean); ok && b.Value {
					return TRUE
				}
			}
		}
		return FALSE
	})

	m.Set("all", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("array::all expects 2 arguments (array, function)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("array::all expects ARRAY")
		}
		fn := args[1]
		if callFunc == nil {
			return errObj("array::all: function caller not initialized")
		}
		for _, elem := range arr.Elements {
			ret := callFunc(fn, []object.Object{elem})
			if ret != nil {
				if b, ok := ret.(*object.Boolean); ok && !b.Value {
					return FALSE
				}
			} else {
				return FALSE
			}
		}
		return TRUE
	})

	m.Desc = "Array operations: filter, map, reduce, sort, unique, find, any, all"
	return m
}
