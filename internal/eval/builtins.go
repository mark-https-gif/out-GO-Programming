package eval

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/out-lang/out/internal/object"
)

var builtins = map[string]*object.BuiltinFunction{
	"print":  {Fn: builtinPrint},
	"input":  {Fn: builtinInput},
	"type":   {Fn: builtinType},
	"len":    {Fn: builtinLen},
	"str":    {Fn: builtinStr},
	"int":    {Fn: builtinInt},
	"float":  {Fn: builtinFloat},
	"append": {Fn: builtinAppend},
	"vibe":   {Fn: builtinVibe},
	"yeet":   {Fn: builtinYeet},
	"sus":    {Fn: builtinSus},
	"flex":   {Fn: builtinFlex},
	"bruh":   {Fn: builtinBruh},
	"howl":   {Fn: builtinHowl},
	"echo":   {Fn: builtinEcho},
	"math::abs":    {Fn: builtinMathAbs},
	"math::sqrt":   {Fn: builtinMathSqrt},
	"math::max":    {Fn: builtinMathMax},
	"math::min":    {Fn: builtinMathMin},
	"math::random": {Fn: builtinMathRandom},
	"math::random_int": {Fn: builtinMathRandomInt},
	"math::floor":  {Fn: builtinMathFloor},
	"math::ceil":   {Fn: builtinMathCeil},
	"strings::upper":     {Fn: builtinStringsUpper},
	"strings::lower":     {Fn: builtinStringsLower},
	"strings::split":     {Fn: builtinStringsSplit},
	"strings::join":      {Fn: builtinStringsJoin},
	"strings::contains":  {Fn: builtinStringsContains},
	"strings::replace":   {Fn: builtinStringsReplace},
	"strings::trim":      {Fn: builtinStringsTrim},
	"strings::starts_with": {Fn: builtinStringsStartsWith},
	"strings::ends_with":   {Fn: builtinStringsEndsWith},
	"strings::repeat":     {Fn: builtinStringsRepeat},
	"time::now":       {Fn: builtinTimeNow},
	"time::clock":     {Fn: builtinTimeClock},
	"time::sleep":     {Fn: builtinTimeSleep},
	"files::read":     {Fn: builtinFilesRead},
	"files::write":    {Fn: builtinFilesWrite},
	"files::exists":   {Fn: builtinFilesExists},
}

func builtinPrint(args ...object.Object) object.Object {
	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return NULL
}

func builtinEcho(args ...object.Object) object.Object {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg.Inspect())
	}
	fmt.Println()
	return NULL
}

func builtinHowl(args ...object.Object) object.Object {
	for _, arg := range args {
		fmt.Println(strings.ToUpper(arg.Inspect()))
	}
	return NULL
}

func builtinInput(args ...object.Object) object.Object {
	if len(args) > 0 {
		fmt.Print(args[0].Inspect())
	}
	var input string
	fmt.Scanln(&input)
	return &object.String{Value: input}
}

func builtinType(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("type() takes exactly 1 argument (%d given)", len(args))
	}
	return &object.String{Value: string(args[0].Type())}
}

func builtinLen(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("len() takes exactly 1 argument (%d given)", len(args))
	}
	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to len() not supported, got %s", args[0].Type())
	}
}

func builtinStr(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("str() takes exactly 1 argument (%d given)", len(args))
	}
	return &object.String{Value: args[0].Inspect()}
}

func builtinInt(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("int() takes exactly 1 argument (%d given)", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return arg
	case *object.Float:
		return &object.Integer{Value: int64(arg.Value)}
	case *object.String:
		var val int64
		_, err := fmt.Sscanf(arg.Value, "%d", &val)
		if err != nil {
			return newError("cannot convert %q to int", arg.Value)
		}
		return &object.Integer{Value: val}
	case *object.Boolean:
		if arg.Value {
			return &object.Integer{Value: 1}
		}
		return &object.Integer{Value: 0}
	default:
		return newError("cannot convert %s to int", args[0].Type())
	}
}

func builtinFloat(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("float() takes exactly 1 argument (%d given)", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Float:
		return arg
	case *object.Integer:
		return &object.Float{Value: float64(arg.Value)}
	case *object.String:
		var val float64
		_, err := fmt.Sscanf(arg.Value, "%f", &val)
		if err != nil {
			return newError("cannot convert %q to float", arg.Value)
		}
		return &object.Float{Value: val}
	default:
		return newError("cannot convert %s to float", args[0].Type())
	}
}

func builtinAppend(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("append() takes exactly 2 arguments (%d given)", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("first argument to append() must be ARRAY, got %s", args[0].Type())
	}
	newElements := make([]object.Object, len(arr.Elements), len(arr.Elements)+1)
	copy(newElements, arr.Elements)
	newElements = append(newElements, args[1])
	return &object.Array{Elements: newElements}
}

func builtinVibe(args ...object.Object) object.Object {
	fmt.Println("good vibes only")
	return NULL
}

func builtinYeet(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("yeet() takes exactly 1 argument (%d given)", len(args))
	}
	return args[0]
}

func builtinSus(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("sus() takes exactly 1 argument (%d given)", len(args))
	}
	if !isTruthy(args[0]) {
		return newError("sus assertion failed")
	}
	return NULL
}

func builtinFlex(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("flex() takes exactly 1 argument (%d given)", len(args))
	}
	return &object.String{Value: args[0].Inspect()}
}

func builtinBruh(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("bruh() takes exactly 1 argument (%d given)", len(args))
	}
	msg, ok := args[0].(*object.String)
	if !ok {
		return newError("bruh() argument must be STRING, got %s", args[0].Type())
	}
	return &object.Error{Message: msg.Value}
}

func builtinMathAbs(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("math::abs() takes exactly 1 argument (%d given)", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		if arg.Value < 0 {
			return &object.Integer{Value: -arg.Value}
		}
		return arg
	case *object.Float:
		return &object.Float{Value: math.Abs(arg.Value)}
	default:
		return newError("argument to math::abs() must be INTEGER or FLOAT, got %s", args[0].Type())
	}
}

func builtinMathSqrt(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("math::sqrt() takes exactly 1 argument (%d given)", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.Float{Value: math.Sqrt(float64(arg.Value))}
	case *object.Float:
		return &object.Float{Value: math.Sqrt(arg.Value)}
	default:
		return newError("argument to math::sqrt() must be INTEGER or FLOAT, got %s", args[0].Type())
	}
}

func builtinMathMax(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("math::max() takes exactly 2 arguments (%d given)", len(args))
	}
	a, b, ok := getNumericPair(args[0], args[1])
	if !ok {
		return newError("arguments to math::max() must be numeric")
	}
	if a > b {
		return args[0]
	}
	return args[1]
}

func builtinMathMin(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("math::min() takes exactly 2 arguments (%d given)", len(args))
	}
	a, b, ok := getNumericPair(args[0], args[1])
	if !ok {
		return newError("arguments to math::min() must be numeric")
	}
	if a < b {
		return args[0]
	}
	return args[1]
}

func builtinMathRandom(args ...object.Object) object.Object {
	return &object.Float{Value: rand.Float64()}
}

func builtinMathRandomInt(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("math::random_int() takes exactly 2 arguments (%d given)", len(args))
	}
	min, ok1 := args[0].(*object.Integer)
	max, ok2 := args[1].(*object.Integer)
	if !ok1 || !ok2 {
		return newError("arguments to math::random_int() must be INTEGER")
	}
	return &object.Integer{Value: min.Value + rand.Int63n(max.Value-min.Value+1)}
}

func builtinMathFloor(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("math::floor() takes exactly 1 argument (%d given)", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return arg
	case *object.Float:
		return &object.Integer{Value: int64(math.Floor(arg.Value))}
	default:
		return newError("argument to math::floor() must be numeric")
	}
}

func builtinMathCeil(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("math::ceil() takes exactly 1 argument (%d given)", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return arg
	case *object.Float:
		return &object.Integer{Value: int64(math.Ceil(arg.Value))}
	default:
		return newError("argument to math::ceil() must be numeric")
	}
}

func builtinStringsUpper(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("strings::upper() takes exactly 1 argument (%d given)", len(args))
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to strings::upper() must be STRING")
	}
	return &object.String{Value: strings.ToUpper(s.Value)}
}

func builtinStringsLower(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("strings::lower() takes exactly 1 argument (%d given)", len(args))
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to strings::lower() must be STRING")
	}
	return &object.String{Value: strings.ToLower(s.Value)}
}

func builtinStringsSplit(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("strings::split() takes exactly 2 arguments (%d given)", len(args))
	}
	s, ok1 := args[0].(*object.String)
	sep, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to strings::split() must be STRING")
	}
	parts := strings.Split(s.Value, sep.Value)
	elems := make([]object.Object, len(parts))
	for i, p := range parts {
		elems[i] = &object.String{Value: p}
	}
	return &object.Array{Elements: elems}
}

func builtinStringsJoin(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("strings::join() takes exactly 2 arguments (%d given)", len(args))
	}
	arr, ok1 := args[0].(*object.Array)
	sep, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to strings::join() must be ARRAY and STRING")
	}
	parts := make([]string, len(arr.Elements))
	for i, e := range arr.Elements {
		s, ok := e.(*object.String)
		if !ok {
			return newError("elements in strings::join() must be STRING")
		}
		parts[i] = s.Value
	}
	return &object.String{Value: strings.Join(parts, sep.Value)}
}

func builtinStringsContains(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("strings::contains() takes exactly 2 arguments (%d given)", len(args))
	}
	s, ok1 := args[0].(*object.String)
	sub, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to strings::contains() must be STRING")
	}
	return nativeBoolToBooleanObject(strings.Contains(s.Value, sub.Value))
}

func builtinStringsReplace(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("strings::replace() takes exactly 3 arguments (%d given)", len(args))
	}
	s, ok1 := args[0].(*object.String)
	old, ok2 := args[1].(*object.String)
	new, ok3 := args[2].(*object.String)
	if !ok1 || !ok2 || !ok3 {
		return newError("arguments to strings::replace() must be STRING")
	}
	return &object.String{Value: strings.Replace(s.Value, old.Value, new.Value, -1)}
}

func builtinStringsTrim(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("strings::trim() takes exactly 1 argument (%d given)", len(args))
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to strings::trim() must be STRING")
	}
	return &object.String{Value: strings.TrimSpace(s.Value)}
}

func builtinStringsStartsWith(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("strings::starts_with() takes exactly 2 arguments (%d given)", len(args))
	}
	s, ok1 := args[0].(*object.String)
	sub, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to strings::starts_with() must be STRING")
	}
	return nativeBoolToBooleanObject(strings.HasPrefix(s.Value, sub.Value))
}

func builtinStringsEndsWith(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("strings::ends_with() takes exactly 2 arguments (%d given)", len(args))
	}
	s, ok1 := args[0].(*object.String)
	sub, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to strings::ends_with() must be STRING")
	}
	return nativeBoolToBooleanObject(strings.HasSuffix(s.Value, sub.Value))
}

func builtinStringsRepeat(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("strings::repeat() takes exactly 2 arguments (%d given)", len(args))
	}
	s, ok1 := args[0].(*object.String)
	n, ok2 := args[1].(*object.Integer)
	if !ok1 || !ok2 {
		return newError("arguments to strings::repeat() must be STRING and INTEGER")
	}
	return &object.String{Value: strings.Repeat(s.Value, int(n.Value))}
}

func builtinTimeNow(args ...object.Object) object.Object {
	return &object.Integer{Value: time.Now().Unix()}
}

func builtinTimeClock(args ...object.Object) object.Object {
	return &object.Float{Value: float64(time.Now().UnixNano()) / 1e9}
}

func builtinTimeSleep(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("time::sleep() takes exactly 1 argument (%d given)", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		time.Sleep(time.Duration(arg.Value) * time.Second)
	case *object.Float:
		time.Sleep(time.Duration(arg.Value * float64(time.Second)))
	default:
		return newError("argument to time::sleep() must be INTEGER or FLOAT")
	}
	return NULL
}

func builtinFilesRead(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("files::read() takes exactly 1 argument (%d given)", len(args))
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to files::read() must be STRING")
	}
	data, err := os.ReadFile(path.Value)
	if err != nil {
		return newError("cannot read file: %s", err.Error())
	}
	return &object.String{Value: string(data)}
}

func builtinFilesWrite(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("files::write() takes exactly 2 arguments (%d given)", len(args))
	}
	path, ok1 := args[0].(*object.String)
	content, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to files::write() must be STRING")
	}
	err := os.WriteFile(path.Value, []byte(content.Value), 0644)
	if err != nil {
		return newError("cannot write file: %s", err.Error())
	}
	return NULL
}

func builtinFilesExists(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("files::exists() takes exactly 1 argument (%d given)", len(args))
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to files::exists() must be STRING")
	}
	_, err := os.Stat(path.Value)
	return nativeBoolToBooleanObject(err == nil)
}

func getNumericPair(a, b object.Object) (float64, float64, bool) {
	var av, bv float64
	switch a := a.(type) {
	case *object.Integer:
		av = float64(a.Value)
	case *object.Float:
		av = a.Value
	default:
		return 0, 0, false
	}
	switch b := b.(type) {
	case *object.Integer:
		bv = float64(b.Value)
	case *object.Float:
		bv = b.Value
	default:
		return 0, 0, false
	}
	return av, bv, true
}
