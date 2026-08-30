// Package stdlib bundles the standard OUT library.
// Every module here wraps Go's standard packages, exposing them
// to OUT scripts through the module::function syntax.
package stdlib

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/out-lang/out/internal/dev"
	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

// RegisterAll adds every built-in module to the registry.
func RegisterAll(reg *module.Registry) {
	reg.Register(osModule())
	reg.Register(strconvModule())
	reg.Register(jsonModule())
	reg.Register(httpModule())
	reg.Register(listModule())
	reg.Register(envModule())
	reg.Register(cryptoModule())
	reg.Register(filesModule())
	reg.Register(mathModule())
	reg.Register(timeModule())
	reg.Register(dev.Module())
	reg.Register(randomModule())
	reg.Register(consoleModule())
	reg.Register(shellModule())
	reg.Register(arrayModule())
	reg.Register(dictModule())
	reg.Register(loggingModule())
}

func osModule() *module.Module {
	m := module.New("os").Set("name", func(args ...object.Object) object.Object {
		return &object.String{Value: "windows"}
	}).Set("cwd", func(args ...object.Object) object.Object {
		wd, err := os.Getwd()
		if err != nil {
			return errObj("os::cwd: " + err.Error())
		}
		return &object.String{Value: wd}
	}).Set("mkdir", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("os::mkdir expects 1 argument")
		}
		s, ok := args[0].(*object.String)
		if !ok {
			return errObj("os::mkdir expects STRING")
		}
		if err := os.MkdirAll(s.Value, 0755); err != nil {
			return errObj("os::mkdir: " + err.Error())
		}
		return NULL
	}).Set("listdir", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("os::listdir expects 1 argument")
		}
		s, ok := args[0].(*object.String)
		if !ok {
			return errObj("os::listdir expects STRING")
		}
		entries, err := os.ReadDir(s.Value)
		if err != nil {
			return errObj("os::listdir: " + err.Error())
		}
		elems := make([]object.Object, len(entries))
		for i, e := range entries {
			elems[i] = &object.String{Value: e.Name()}
		}
		return &object.Array{Elements: elems}
	}).Set("getenv", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("os::getenv expects 1 argument")
		}
		s, ok := args[0].(*object.String)
		if !ok {
			return errObj("os::getenv expects STRING")
		}
		return &object.String{Value: os.Getenv(s.Value)}
	})
	m.Desc = "Operating system interface (wraps Go os package)"
	return m
}

func strconvModule() *module.Module {
	m := module.New("strconv").Set("atoi", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("strconv::atoi expects 1 argument")
		}
		s, ok := args[0].(*object.String)
		if !ok {
			return errObj("strconv::atoi expects STRING")
		}
		n, err := strconv.Atoi(s.Value)
		if err != nil {
			return errObj("strconv::atoi: " + err.Error())
		}
		return &object.Integer{Value: int64(n)}
	}).Set("itoa", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("strconv::itoa expects 1 argument")
		}
		iv, ok := args[0].(*object.Integer)
		if !ok {
			return errObj("strconv::itoa expects INTEGER")
		}
		return &object.String{Value: strconv.FormatInt(iv.Value, 10)}
	}).Set("parse_float", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("strconv::parse_float expects 1 argument")
		}
		s, ok := args[0].(*object.String)
		if !ok {
			return errObj("strconv::parse_float expects STRING")
		}
		f, err := strconv.ParseFloat(s.Value, 64)
		if err != nil {
			return errObj("strconv::parse_float: " + err.Error())
		}
		return &object.Float{Value: f}
	}).Set("format_float", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("strconv::format_float expects 1 argument")
		}
		fv, ok := args[0].(*object.Float)
		if !ok {
			return errObj("strconv::format_float expects FLOAT")
		}
		return &object.String{Value: strconv.FormatFloat(fv.Value, 'f', -1, 64)}
	})
	m.Desc = "String number conversions (wraps Go strconv package)"
	return m
}

func jsonModule() *module.Module {
	m := module.New("json").Set("parse", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("json::parse expects 1 argument")
		}
		s, ok := args[0].(*object.String)
		if !ok {
			return errObj("json::parse expects STRING")
		}
		var raw interface{}
		if err := json.Unmarshal([]byte(s.Value), &raw); err != nil {
			return errObj("json::parse: " + err.Error())
		}
		return fromGo(raw)
	}).Set("stringify", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("json::stringify expects 1 argument")
		}
		g := toGo(args[0])
		bytes, err := json.Marshal(g)
		if err != nil {
			return errObj("json::stringify: " + err.Error())
		}
		return &object.String{Value: string(bytes)}
	})
	m.Desc = "JSON parsing and serialization (wraps Go encoding/json)"
	return m
}

type httpResult struct {
	Status  int
	Body    string
	OK      bool
}

func httpModule() *module.Module {
	m := module.New("http")
	m.Set("get", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("http::get expects 1 argument")
		}
		url, ok := args[0].(*object.String)
		if !ok {
			return errObj("http::get expects STRING url")
		}
		return httpGet(url.Value)
	})
	m.Desc = "Simple HTTP client (wraps Go net/http)"
	return m
}

func listModule() *module.Module {
	m := module.New("list").Set("new", func(args ...object.Object) object.Object {
		return &object.Array{Elements: []object.Object{}}
	}).Set("push", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("list::push expects 2 arguments")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("list::push expects ARRAY")
		}
		elems := make([]object.Object, len(arr.Elements)+1)
		copy(elems, arr.Elements)
		elems[len(arr.Elements)] = args[1]
		return &object.Array{Elements: elems}
	}).Set("contains", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("list::contains expects 2 arguments")
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return errObj("list::contains expects ARRAY")
		}
		for _, e := range arr.Elements {
			if toGo(e) == toGo(args[1]) {
				return TRUE
			}
		}
		return FALSE
	})
	m.Desc = "List helpers"
	return m
}

func envModule() *module.Module {
	m := module.New("env").Set("all", func(args ...object.Object) object.Object {
		envs := os.Environ()
		pairs := make(map[uint64]object.HashPair)
		for _, e := range envs {
			kv := strings.SplitN(e, "=", 2)
			if len(kv) == 2 {
				key := &object.String{Value: kv[0]}
				val := &object.String{Value: kv[1]}
				pairs[object.HashKey(key)] = object.HashPair{Key: key, Value: val}
			}
		}
		return &object.Hash{Pairs: pairs}
	})
	m.Desc = "Environment variables access"
	return m
}

// ---- go<->out conversion helpers ----

// fromGo converts a decoded json value into OUT objects.
func fromGo(v interface{}) object.Object {
	switch t := v.(type) {
	case nil:
		return NULL
	case bool:
		if t {
			return TRUE
		}
		return FALSE
	case float64:
		if t == float64(int64(t)) {
			return &object.Integer{Value: int64(t)}
		}
		return &object.Float{Value: t}
	case string:
		return &object.String{Value: t}
	case []interface{}:
		elems := make([]object.Object, len(t))
		for i, e := range t {
			elems[i] = fromGo(e)
		}
		return &object.Array{Elements: elems}
	case map[string]interface{}:
		pairs := make(map[uint64]object.HashPair)
		for k, val := range t {
			key := &object.String{Value: k}
			pairs[object.HashKey(key)] = object.HashPair{Key: key, Value: fromGo(val)}
		}
		return &object.Hash{Pairs: pairs}
	default:
		return &object.String{Value: fmt.Sprintf("%v", v)}
	}
}

// toGo converts OUT objects into plain Go values for JSON/marshaling.
func toGo(o object.Object) interface{} {
	switch t := o.(type) {
	case *object.Integer:
		return t.Value
	case *object.Float:
		return t.Value
	case *object.String:
		return t.Value
	case *object.Boolean:
		return t.Value
	case *object.Null:
		return nil
	case *object.Array:
		res := make([]interface{}, len(t.Elements))
		for i, e := range t.Elements {
			res[i] = toGo(e)
		}
		return res
	case *object.Hash:
		res := make(map[string]interface{})
		for _, p := range t.Pairs {
			k, ok := p.Key.(*object.String)
			if ok {
				res[k.Value] = toGo(p.Value)
			}
		}
		return res
	default:
		return t.Inspect()
	}
}

// small helpers
var (
	NULL = &object.Null{}
	TRUE = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func errObj(msg string) *object.Error {
	return &object.Error{Message: msg}
}
