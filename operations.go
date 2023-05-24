package deobfuscator

import (
	"github.com/dop251/goja/ast"
	"regexp"
)

// key : function name
// value : operator
var operationFunctions = map[string][2]string{}

func addFunctionToOperations(function *ast.FunctionLiteral, functionName string) {
	switch function.Body.List[0].(type) {
	case *ast.ReturnStatement:
		statement := function.Body.List[0].(*ast.ReturnStatement)
		switch statement.Argument.(type) {
		case *ast.UnaryExpression:
			operation := statement.Argument.(*ast.UnaryExpression)
			operationFunctions[functionName] = [2]string{operation.Operator.String(), "1"}
		case *ast.BinaryExpression:
			operation := statement.Argument.(*ast.BinaryExpression)
			operationFunctions[functionName] = [2]string{operation.Operator.String(), "2"}
		}
	}
}

func replaceOperations(script *string) {
	for name, operator := range operationFunctions {
		var functionReg *regexp.Regexp
		switch operator[1] {
		case "1":
			functionReg = regexp.MustCompile(name + `\(([\w\d._"]+)\)`)
			*script = functionReg.ReplaceAllString(*script, ""+operator[0]+" $1")
			break

		default:
			functionReg = regexp.MustCompile(name + `\(([\w\d._"]+),\s?([\w\d._"]+)\)`)
			*script = functionReg.ReplaceAllString(*script, "$1 "+operator[0]+" $2")
		}
	}
}
