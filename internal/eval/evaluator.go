package eval

import (
	"fmt"
	"os"

	"github.com/out-lang/out/internal/ast"
	"github.com/out-lang/out/internal/env"
	"github.com/out-lang/out/internal/lexer"
	"github.com/out-lang/out/internal/object"
	"github.com/out-lang/out/internal/parser"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, e *env.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, e)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, e)
	case *ast.BlockStatement:
		return evalBlockStatement(node, e)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, e)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, e)
		if isError(val) {
			return val
		}
		e.Set(node.Name.Value, val)
	case *ast.ImportStatement:
		return evalImport(node, e)
	case *ast.AssignExpression:
		val := Eval(node.Value, e)
		if isError(val) {
			return val
		}
		e.Set(node.Name.Value, val)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.NullLiteral:
		return NULL
	case *ast.PrefixExpression:
		right := Eval(node.Right, e)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, e)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, e)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		return evalIfExpression(node, e)
	case *ast.WhileExpression:
		return evalWhileExpression(node, e)
	case *ast.ForExpression:
		return evalForExpression(node, e)
	case *ast.Identifier:
		return evalIdentifier(node, e)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: e}
	case *ast.CallExpression:
		function := Eval(node.Function, e)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, e)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	case *ast.ArrayLiteral:
		elems := evalExpressions(node.Elements, e)
		if len(elems) == 1 && isError(elems[0]) {
			return elems[0]
		}
		return &object.Array{Elements: elems}
	case *ast.HashLiteral:
		return evalHashLiteral(node, e)
	case *ast.IndexExpression:
		left := Eval(node.Left, e)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, e)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	case *ast.MemberAccess:
		if ident, ok := node.Object.(*ast.Identifier); ok {
			if m := lookupModuleMember(ident.Value, node.Member); m != nil {
				return m
			}
			if builtin, ok := builtins[ident.Value+"::"+node.Member]; ok {
				return builtin
			}
		}
		left := Eval(node.Object, e)
		if isError(left) {
			return left
		}
		return evalMemberAccess(left, node.Member)
	}
	return nil
}

func evalProgram(program *ast.Program, e *env.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, e)
		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
		if _, ok := result.(*object.Error); ok {
			return result
		}
	}
	return result
}

// evalImport loads another .out file in the current environment, making the
// functions/variables it defines available to the importer. If the name
// matches a registered Go module, it resolves in place (no file load needed).
func evalImport(node *ast.ImportStatement, e *env.Environment) object.Object {
	path := node.Path
	if path == "" {
		return newError("import requires a file path")
	}
	if _, ok := registry.Lookup(path); ok {
		return NULL
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return newError("cannot import '%s': %s", path, err.Error())
	}
	l := lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return newError("cannot import '%s': %s", path, p.Errors()[0])
	}
	return Eval(program, e)
}

func evalBlockStatement(block *ast.BlockStatement, e *env.Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, e)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "-":
		return evalMinusPrefixExpression(right)
	case "!":
		return evalBangPrefixExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalMinusPrefixExpression(right object.Object) object.Object {
	switch right := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -right.Value}
	case *object.Float:
		return &object.Float{Value: -right.Value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

func evalBangPrefixExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		lv := float64(left.(*object.Integer).Value)
		rv := right.(*object.Float).Value
		return evalFloatInfixExpressionRaw(operator, lv, rv)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		lv := left.(*object.Float).Value
		rv := float64(right.(*object.Integer).Value)
		return evalFloatInfixExpressionRaw(operator, lv, rv)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case operator == "and":
		return nativeBoolToBooleanObject(isTruthy(left) && isTruthy(right))
	case operator == "or":
		return nativeBoolToBooleanObject(isTruthy(left) || isTruthy(right))
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	lv := left.(*object.Integer).Value
	rv := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: lv + rv}
	case "-":
		return &object.Integer{Value: lv - rv}
	case "*":
		return &object.Integer{Value: lv * rv}
	case "/":
		if rv == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: lv / rv}
	case "%":
		if rv == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: lv % rv}
	case "<":
		return nativeBoolToBooleanObject(lv < rv)
	case ">":
		return nativeBoolToBooleanObject(lv > rv)
	case "<=":
		return nativeBoolToBooleanObject(lv <= rv)
	case ">=":
		return nativeBoolToBooleanObject(lv >= rv)
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s", operator)
	}
}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	lv := left.(*object.Float).Value
	rv := right.(*object.Float).Value
	return evalFloatInfixExpressionRaw(operator, lv, rv)
}

func evalFloatInfixExpressionRaw(operator string, lv, rv float64) object.Object {
	switch operator {
	case "+":
		return &object.Float{Value: lv + rv}
	case "-":
		return &object.Float{Value: lv - rv}
	case "*":
		return &object.Float{Value: lv * rv}
	case "/":
		if rv == 0 {
			return newError("division by zero")
		}
		return &object.Float{Value: lv / rv}
	case "<":
		return nativeBoolToBooleanObject(lv < rv)
	case ">":
		return nativeBoolToBooleanObject(lv > rv)
	case "<=":
		return nativeBoolToBooleanObject(lv <= rv)
	case ">=":
		return nativeBoolToBooleanObject(lv >= rv)
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s", operator)
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	lv := left.(*object.String).Value
	rv := right.(*object.String).Value
	switch operator {
	case "+":
		return &object.String{Value: lv + rv}
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIfExpression(node *ast.IfExpression, e *env.Environment) object.Object {
	condition := Eval(node.Condition, e)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(node.Consequence, e)
	} else if node.Alternative != nil {
		return Eval(node.Alternative, e)
	}
	return NULL
}

func evalWhileExpression(node *ast.WhileExpression, e *env.Environment) object.Object {
	var result object.Object
	for {
		condition := Eval(node.Condition, e)
		if isError(condition) {
			return condition
		}
		if !isTruthy(condition) {
			break
		}
		result = Eval(node.Body, e)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalForExpression(node *ast.ForExpression, e *env.Environment) object.Object {
	start := Eval(node.Start, e)
	if isError(start) {
		return start
	}
	end := Eval(node.End, e)
	if isError(end) {
		return end
	}
	startInt, ok1 := start.(*object.Integer)
	endInt, ok2 := end.(*object.Integer)
	if !ok1 || !ok2 {
		return newError("range bounds must be integers")
	}
	var result object.Object
	for i := startInt.Value; i < endInt.Value; i++ {
		loopEnv := env.NewEnclosed(e)
		loopEnv.Set(node.Name.Value, &object.Integer{Value: i})
		result = evalBlockStatement(node.Body, loopEnv)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalIdentifier(node *ast.Identifier, e *env.Environment) object.Object {
	if val, ok := e.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("line %d:%d - identifier not found: %s", node.Token.Line, node.Token.Pos, node.Value)
}

func evalExpressions(exps []ast.Expression, e *env.Environment) []object.Object {
	var result []object.Object
	for _, exp := range exps {
		evaluated := Eval(exp, e)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extended := extendFunctionEnv(fn, args)
		evaluated := evalBlockStatement(fn.Body, extended)
		return unwrapReturnValue(evaluated)
	case *object.BuiltinFunction:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *env.Environment {
	parent, ok := fn.Env.(*env.Environment)
	if !ok {
		return env.New()
	}
	extended := env.NewEnclosed(parent)
	for paramIdx, param := range fn.Parameters {
		extended.Set(param.Value, args[paramIdx])
	}
	return extended
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}
	return arrayObject.Elements[idx]
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)
	if !object.IsHashable(index) {
		return newError("unusable as hash key: %s", index.Type())
	}
	key := object.HashKey(index)
	pair, ok := hashObject.Pairs[key]
	if !ok {
		return NULL
	}
	return pair.Value
}

func evalHashLiteral(node *ast.HashLiteral, e *env.Environment) object.Object {
	pairs := make(map[uint64]object.HashPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, e)
		if isError(key) {
			return key
		}
		if !object.IsHashable(key) {
			return newError("unusable as hash key: %s", key.Type())
		}
		value := Eval(valueNode, e)
		if isError(value) {
			return value
		}
		hashKey := object.HashKey(key)
		pairs[hashKey] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}

func evalMemberAccess(left object.Object, member string) object.Object {
	switch obj := left.(type) {
	case *object.Array:
		switch member {
		case "len":
			return &object.Integer{Value: int64(len(obj.Elements))}
		case "first":
			if len(obj.Elements) > 0 {
				return obj.Elements[0]
			}
			return NULL
		case "last":
			if len(obj.Elements) > 0 {
				return obj.Elements[len(obj.Elements)-1]
			}
			return NULL
		default:
			return newError("unknown array member: %s", member)
		}
	case *object.String:
		switch member {
		case "len":
			return &object.Integer{Value: int64(len(obj.Value))}
		default:
			return newError("unknown string member: %s", member)
		}
	case *object.Hash:
		switch member {
		case "keys":
			keys := []object.Object{}
			for _, pair := range obj.Pairs {
				keys = append(keys, pair.Key)
			}
			return &object.Array{Elements: keys}
		case "values":
			vals := []object.Object{}
			for _, pair := range obj.Pairs {
				vals = append(vals, pair.Value)
			}
			return &object.Array{Elements: vals}
		default:
			return newError("unknown hash member: %s", member)
		}
	default:
		return newError("member access not supported on %s", left.Type())
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
