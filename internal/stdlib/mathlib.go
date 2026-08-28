package stdlib

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

func mathModule() *module.Module {
	m := module.New("math")
	// constants
	m.Set("pi", func(args ...object.Object) object.Object {
		return &object.Float{Value: math.Pi}
	}).Set("e", func(args ...object.Object) object.Object {
		return &object.Float{Value: math.E}
	}).Set("inf", func(args ...object.Object) object.Object {
		return &object.Float{Value: math.Inf(1)}
	})
	// trigonometry
	m.Set("sin", func(args ...object.Object) object.Object {
		f, ok := requireNum(args)
		if !ok {
			return errObj("math::sin expects a number")
		}
		return &object.Float{Value: math.Sin(f)}
	}).Set("cos", func(args ...object.Object) object.Object {
		f, ok := requireNum(args)
		if !ok {
			return errObj("math::cos expects a number")
		}
		return &object.Float{Value: math.Cos(f)}
	}).Set("tan", func(args ...object.Object) object.Object {
		f, ok := requireNum(args)
		if !ok {
			return errObj("math::tan expects a number")
		}
		return &object.Float{Value: math.Tan(f)}
	}).Set("atan2", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("math::atan2 expects 2 arguments")
		}
		y, ok1 := asFloat(args[0])
		x, ok2 := asFloat(args[1])
		if !ok1 || !ok2 {
			return errObj("math::atan2 expects numbers")
		}
		return &object.Float{Value: math.Atan2(y, x)}
	})
	// logs and powers
	m.Set("log", func(args ...object.Object) object.Object {
		f, ok := requireNum(args)
		if !ok {
			return errObj("math::log expects a number")
		}
		return &object.Float{Value: math.Log(f)}
	}).Set("log10", func(args ...object.Object) object.Object {
		f, ok := requireNum(args)
		if !ok {
			return errObj("math::log10 expects a number")
		}
		return &object.Float{Value: math.Log10(f)}
	}).Set("pow", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("math::pow expects 2 arguments")
		}
		base, ok1 := asFloat(args[0])
		exp, ok2 := asFloat(args[1])
		if !ok1 || !ok2 {
			return errObj("math::pow expects numbers")
		}
		return &object.Float{Value: math.Pow(base, exp)}
	}).Set("exp", func(args ...object.Object) object.Object {
		f, ok := requireNum(args)
		if !ok {
			return errObj("math::exp expects a number")
		}
		return &object.Float{Value: math.Exp(f)}
	})
	// rounding
	m.Set("round", func(args ...object.Object) object.Object {
		f, ok := requireNum(args)
		if !ok {
			return errObj("math::round expects a number")
		}
		return &object.Float{Value: math.Round(f)}
	})
	// random extras
	m.Set("random_int", func(args ...object.Object) object.Object {
		if len(args) == 0 {
			return &object.Integer{Value: rand.Int63()}
		}
		if len(args) == 1 {
			max, ok := args[0].(*object.Integer)
			if !ok {
				return errObj("math::random_int expects INTEGER")
			}
			return &object.Integer{Value: rand.Int63n(max.Value)}
		}
		min, ok1 := args[0].(*object.Integer)
		max, ok2 := args[1].(*object.Integer)
		if !ok1 || !ok2 {
			return errObj("math::random_int expects INTEGERs")
		}
		return &object.Integer{Value: min.Value + rand.Int63n(max.Value-min.Value+1)}
	})
	m.Desc = "Math functions (wraps Go math package)"
	return m
}

// timelib provides date/time formatting on top of builtin time::* helpers.
func timeModule() *module.Module {
	m := module.New("time")
	m.Set("now", func(args ...object.Object) object.Object {
		now := time.Now()
		pairs := make(map[uint64]object.HashPair)
		add := func(k *object.String, v object.Object) {
			pairs[object.HashKey(k)] = object.HashPair{Key: k, Value: v}
		}
		add(&object.String{Value: "timestamp"}, &object.Integer{Value: now.Unix()})
		add(&object.String{Value: "string"}, &object.String{Value: now.Format("2006-01-02 15:04:05")})
		add(&object.String{Value: "date"}, &object.String{Value: now.Format("2006-01-02")})
		add(&object.String{Value: "time"}, &object.String{Value: now.Format("15:04:05")})
		return &object.Hash{Pairs: pairs}
	}).Set("unix", func(args ...object.Object) object.Object {
		return &object.Integer{Value: time.Now().Unix()}
	}).Set("millis", func(args ...object.Object) object.Object {
		return &object.Integer{Value: time.Now().UnixMilli()}
	}).Set("format", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("time::format expects 1 argument")
		}
		layout, ok := requireString("format", args[0])
		if !ok {
			return errObj("time::format expects STRING")
		}
		layout = strings.ReplaceAll(layout, "YYYY", "2006")
		layout = strings.ReplaceAll(layout, "MM", "01")
		layout = strings.ReplaceAll(layout, "DD", "02")
		layout = strings.ReplaceAll(layout, "HH", "15")
		layout = strings.ReplaceAll(layout, "mm", "04")
		layout = strings.ReplaceAll(layout, "ss", "05")
		return &object.String{Value: time.Now().Format(layout)}
	})
	m.Desc = "Time and date helpers (wraps Go time package)"
	return m
}

var seeded = false

func init() {
	if !seeded {
		rand.Seed(time.Now().UnixNano())
		seeded = true
	}
}

func requireNum(args []object.Object) (float64, bool) {
	if len(args) != 1 {
		return 0, false
	}
	return asFloat(args[0])
}

func asFloat(o object.Object) (float64, bool) {
	switch t := o.(type) {
	case *object.Integer:
		return float64(t.Value), true
	case *object.Float:
		return t.Value, true
	default:
		return 0, false
	}
}
