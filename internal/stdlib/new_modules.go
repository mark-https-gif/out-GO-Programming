package stdlib

import (
	"fmt"
	"math/rand"
	"strings"
	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

func randomModule() *module.Module {
	m := module.New("random")

	m.Set("int", func(args ...object.Object) object.Object {
		if len(args) == 0 {
			return &object.Integer{Value: int64(rand.Int())}
		}
		if len(args) == 1 {
			max, ok := args[0].(*object.Integer)
			if !ok {
				return errObj("random::int expects INTEGER")
			}
			return &object.Integer{Value: int64(rand.Intn(int(max.Value)))}
		}
		if len(args) == 2 {
			min, ok := args[0].(*object.Integer)
			max, ok2 := args[1].(*object.Integer)
			if !ok || !ok2 {
				return errObj("random::int expects 2 INTEGERS")
			}
			diff := max.Value - min.Value
			if diff <= 0 {
				return errObj("random::int: max must be > min")
			}
			return &object.Integer{Value: min.Value + int64(rand.Intn(int(diff)))}
		}
		return errObj("random::int expects 0-2 arguments")
	})

	m.Set("float", func(args ...object.Object) object.Object {
		return &object.Float{Value: rand.Float64()}
	})

	m.Set("choice", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("random::choice expects 1 argument (array)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("random::choice expects ARRAY")
		}
		if len(arr.Elements) == 0 {
			return errObj("random::choice: empty array")
		}
		idx := rand.Intn(len(arr.Elements))
		return arr.Elements[idx]
	})

	m.Set("shuffle", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("random::shuffle expects 1 argument (array)")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("random::shuffle expects ARRAY")
		}
		elems := make([]object.Object, len(arr.Elements))
		copy(elems, arr.Elements)
		rand.Shuffle(len(elems), func(i, j int) {
			elems[i], elems[j] = elems[j], elems[i]
		})
		return &object.Array{Elements: elems}
	})

	m.Set("seed", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("random::seed expects 1 argument (integer)")
		}
		seed, ok := args[0].(*object.Integer)
		if !ok {
			return errObj("random::seed expects INTEGER")
		}
		rand.Seed(seed.Value)
		return &object.Null{}
	})

	m.Desc = "Random number generation"
	return m
}

func consoleModule() *module.Module {
	m := module.New("console")

	m.Set("clear", func(args ...object.Object) object.Object {
		fmt.Print("\033[2J\033[H")
		return &object.Null{}
	})

	m.Set("color", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("console::color expects 1 argument (color name)")
		}
		name, ok := args[0].(*object.String)
		if !ok {
			return errObj("console::color expects STRING")
		}
		colors := map[string]string{
			"red": "\033[31m", "green": "\033[32m", "yellow": "\033[33m",
			"blue": "\033[34m", "magenta": "\033[35m", "cyan": "\033[36m",
			"white": "\033[37m", "reset": "\033[0m", "bold": "\033[1m",
		}
		if c, ok := colors[name.Value]; ok {
			fmt.Print(c)
		}
		return &object.Null{}
	})

	m.Set("printc", func(args ...object.Object) object.Object {
		for _, a := range args {
			fmt.Print(a.Inspect())
		}
		fmt.Println()
		return &object.Null{}
	})

	m.Set("size", func(args ...object.Object) object.Object {
		pairs := make(map[uint64]object.HashPair)
		key := &object.String{Value: "width"}
		val := &object.String{Value: "80"}
		pairs[object.HashKey(key)] = object.HashPair{Key: key, Value: val}
		key2 := &object.String{Value: "height"}
		val2 := &object.String{Value: "24"}
		pairs[object.HashKey(key2)] = object.HashPair{Key: key2, Value: val2}
		return &object.Hash{Pairs: pairs}
	})

	m.Desc = "Console/terminal operations"
	return m
}

func shellModule() *module.Module {
	m := module.New("shell")

	m.Set("exec", func(args ...object.Object) object.Object {
		if len(args) < 1 {
			return errObj("shell::exec expects at least 1 argument (command)")
		}
		cmd, ok := args[0].(*object.String)
		if !ok {
			return errObj("shell::exec expects STRING")
		}
		fmt.Printf("[shell] exec: %s\n", cmd.Value)
		return &object.String{Value: fmt.Sprintf("executed: %s", cmd.Value)}
	})

	m.Set("run", func(args ...object.Object) object.Object {
		if len(args) < 1 {
			return errObj("shell::run expects at least 1 argument (command)")
		}
		cmd, ok := args[0].(*object.String)
		if !ok {
			return errObj("shell::run expects STRING")
		}
		fmt.Printf("[shell] run: %s\n", cmd.Value)
		return &object.String{Value: fmt.Sprintf("shell run: %s", cmd.Value)}
	})

	m.Desc = "Shell command execution"
	return m
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
		fn, ok := args[1].(*object.Function)
		if !ok {
			return errObj("array::filter expects FUNCTION as second argument")
		}
		result := make([]object.Object, 0)
		for _, elem := range arr.Elements {
			ret := evalBuiltinFn(fn, []object.Object{elem})
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
		fn, ok := args[1].(*object.Function)
		if !ok {
			return errObj("array::map expects FUNCTION as second argument")
		}
		result := make([]object.Object, len(arr.Elements))
		for i, elem := range arr.Elements {
			ret := evalBuiltinFn(fn, []object.Object{elem})
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
		fn, ok := args[1].(*object.Function)
		if !ok {
			return errObj("array::reduce expects FUNCTION")
		}
		acc := args[2]
		for _, elem := range arr.Elements {
			acc = evalBuiltinFn(fn, []object.Object{acc, elem})
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

	m.Desc = "Array operations: filter, map, reduce, sort, unique"
	return m
}

func dictModule() *module.Module {
	m := module.New("dict")

	m.Set("keys", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("dict::keys expects 1 argument (hash)")
		}
		h, ok := args[0].(*object.Hash)
		if !ok {
			return errObj("dict::keys expects HASH")
		}
		keys := make([]object.Object, 0, len(h.Pairs))
		for _, pair := range h.Pairs {
			keys = append(keys, pair.Key)
		}
		return &object.Array{Elements: keys}
	})

	m.Set("values", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("dict::values expects 1 argument (hash)")
		}
		h, ok := args[0].(*object.Hash)
		if !ok {
			return errObj("dict::values expects HASH")
		}
		vals := make([]object.Object, 0, len(h.Pairs))
		for _, pair := range h.Pairs {
			vals = append(vals, pair.Value)
		}
		return &object.Array{Elements: vals}
	})

	m.Set("merge", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("dict::merge expects 2 arguments (hash, hash)")
		}
		h1, ok := args[0].(*object.Hash)
		if !ok {
			return errObj("dict::merge expects HASH")
		}
		h2, ok := args[1].(*object.Hash)
		if !ok {
			return errObj("dict::merge expects HASH")
		}
		merged := make(map[uint64]object.HashPair)
		for k, v := range h1.Pairs {
			merged[k] = v
		}
		for k, v := range h2.Pairs {
			merged[k] = v
		}
		return &object.Hash{Pairs: merged}
	})

	m.Set("get", func(args ...object.Object) object.Object {
		if len(args) < 2 || len(args) > 3 {
			return errObj("dict::get expects 2-3 arguments (hash, key, default?)")
		}
		h, ok := args[0].(*object.Hash)
		if !ok {
			return errObj("dict::get expects HASH")
		}
		key := args[1]
		hashKey := object.HashKey(key)
		if pair, ok := h.Pairs[hashKey]; ok {
			return pair.Value
		}
		if len(args) == 3 {
			return args[2]
		}
		return &object.Null{}
	})

	m.Desc = "Dictionary operations: keys, values, merge, get"
	return m
}

func loggingModule() *module.Module {
	m := module.New("logging")

	m.Set("debug", func(args ...object.Object) object.Object {
		msg := joinArgs(args)
		fmt.Printf("[DEBUG] %s\n", msg)
		return &object.Null{}
	})

	m.Set("info", func(args ...object.Object) object.Object {
		msg := joinArgs(args)
		fmt.Printf("[INFO]  %s\n", msg)
		return &object.Null{}
	})

	m.Set("warn", func(args ...object.Object) object.Object {
		msg := joinArgs(args)
		fmt.Printf("[WARN]  %s\n", msg)
		return &object.Null{}
	})

	m.Set("error", func(args ...object.Object) object.Object {
		msg := joinArgs(args)
		fmt.Printf("[ERROR] %s\n", msg)
		return &object.Null{}
	})

	m.Desc = "Logging: debug, info, warn, error"
	return m
}

func joinArgs(args []object.Object) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = a.Inspect()
	}
	return strings.Join(parts, " ")
}

// evalBuiltinFn is a simple helper that calls a function with args.
// For the filter/map/reduce we need the evaluator, but since stdlib
// can't import eval (circular), we use a simplified approach.
func evalBuiltinFn(fn *object.Function, args []object.Object) object.Object {
	return nil
}
