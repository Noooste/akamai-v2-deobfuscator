package deobfuscator

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja/ast"
)

func loadScript(rawScript string) (prg *ast.Program, err error) {
	script, err := goja.Parse("script.js", rawScript)
	if err != nil {
		return nil, err
	}
	return script, nil
}

func cleanScript(script *ast.Program) (reducedScript string, enterIdentifier string, arguments []string, isCall bool) {
	inside := script.Body[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)

	function := inside.Callee.(*ast.FunctionLiteral)
	reducedScript = function.Source[function.Body.LeftBrace-1 : function.Body.RightBrace-2]

	for i := 0; i < len(function.Body.List); i++ {
		switch function.Body.List[i].(type) {
		case *ast.ReturnStatement:
			statement := function.Body.List[i].(*ast.ReturnStatement).Argument.(*ast.CallExpression)
			dotStatement := statement.Callee.(*ast.DotExpression)
			if dotStatement.Identifier.Name.String() == "call" {
				isCall = true
			}
			enterIdentifier = dotStatement.Left.(*ast.Identifier).Name.String()
			for _, argument := range statement.ArgumentList {
				switch argument.(type) {
				case *ast.ThisExpression:
					arguments = append(arguments, "this")
				case *ast.Identifier:
					arguments = append(arguments, argument.(*ast.Identifier).Name.String())
				}
			}
		}
	}
	script.File.Base()
	return
}
