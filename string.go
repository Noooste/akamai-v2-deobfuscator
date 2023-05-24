package deobfuscator

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja/ast"
	"log"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

var pushValueRegex = regexp.MustCompile(`(\w+).push\((\w+)\)`)
var stringUnescape = regexp.MustCompile(`\\?(?:\\{2})*"`)

const INFINITE = 0xffffffff

type Virtual struct {
	vm      *goja.Runtime
	program *ast.Program
	lock    *sync.Mutex

	script string

	pusherName string

	listName                 string
	objectName               string
	functionBigOperationName string

	changeOperators bool
	windowName      string

	maxIndex int

	scriptId               chan string
	deviceInformationOrder chan []string

	verbose          bool
	verboseTime      time.Time
	verboseDeltaTime time.Time

	operationFunctions map[string][2]string

	deobScript   bool
	deobedScript string
}

func (v *Virtual) evalString() {
	defer func() {
		if r := recover(); r != nil {
			log.Print(r)
			debug.PrintStack()
		}
	}()

	if v.maxIndex == 0 {
		v.maxIndex = INFINITE
	}

	if v.listName == "" {
		v.listName = pushValueRegex.FindStringSubmatch(v.script)[1]
		v.runInVm("var " + v.listName + "=[];")
	}

	if v.objectName == "" {
		var randomSub = regexp.MustCompile(`return \['(\w{2,3})',`).FindStringSubmatch(v.script)[1]
		v.objectName = regexp.MustCompile(`(\w+)\.` + randomSub).FindStringSubmatch(v.script)[1]
	}

	//I really don't know why it fixes the parsing, but anyway it works so
	var _, err = v.runInVm(v.listName + ".push(0);")
	log.Print(err)

	for _, statement := range v.program.Body {
		v.evalInto(statement)
	}
}

func (v *Virtual) ifStatement(expression *ast.IfStatement) any {
	switch expression.Test.(type) {
	case *ast.SequenceExpression, *ast.CallExpression, *ast.BracketExpression, *ast.ArrayPattern, *ast.UnaryExpression:
		v.evalInto(expression.Test)
	}
	v.evalInto(expression.Consequent)
	v.evalInto(expression.Alternate)
	return nil
}

func (v *Virtual) switchStatement(expression *ast.SwitchStatement) any {
	for _, switchCase := range expression.Body {
		for _, consequent := range switchCase.Consequent {
			switch consequent.(type) {
			case *ast.BlockStatement:
				for _, statement := range consequent.(*ast.BlockStatement).List {
					v.evalInto(statement)
				}
			}
		}
	}
	return nil
}

func isFine(s string) bool {
	for _, r := range s {
		if r < 33 || r > 127 || r == '\\' || r == '\'' || r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\v' || r == '\f' {
			return false
		}
	}
	return true
}

func (v *Virtual) dotStatement(expression *ast.DotExpression, parentExpression any) any {
	identifier := expression.Identifier.Name.String()

	switch expression.Left.(type) {
	case *ast.DotExpression:
		return v.dotStatement(expression.Left.(*ast.DotExpression), parentExpression)

	case *ast.BracketExpression:
		return v.bracketStatement(expression.Left.(*ast.BracketExpression))

	case *ast.FunctionLiteral:
		return v.evalInto(expression.Left.(*ast.FunctionLiteral))

	default:
		name := expression.Left.(*ast.Identifier).Name.String()
		var element string

		switch parentExpression.(type) {

		case *ast.CallExpression:
			callExpression := parentExpression.(*ast.CallExpression)
			element = v.script[callExpression.Idx0()-1 : callExpression.Idx1()-1]

		case *ast.BracketExpression:
			bracketExpression := parentExpression.(*ast.BracketExpression)
			element = v.script[bracketExpression.Idx0()-1 : bracketExpression.Idx1()-1]
		}

		if name == v.objectName {
			value, err := v.runInVm(element)
			if err != nil {
				log.Print(err)
			}

			if err == nil {
				var toString = value.String()
				if v.deobScript {
					if isFine(toString) {
						result := `"` + stringUnescape.ReplaceAllStringFunc(toString, func(s string) string {
							if strings.Count(s, `\`)%2 == 0 {
								s = `\` + s
							}
							if s[len(s)-1] == '\\' {
								s += `\`
							}

							return s
						}) + `"`

						v.deobedScript = strings.Replace(v.deobedScript, element, result, -1)
					}
				} else if strings.HasSuffix(toString, "==") {
					v.scriptId <- toString
				}
			}

		} else if identifier == "push" {
			//if we detect push list (the list to decrypt string)
			if v.pusherName == "" || name == v.pusherName {
				//if push name is not set, set it
				if v.pusherName == "" {
					v.pusherName = name
				}

				return element
			}
		}

		return nil
	}
}

func (v *Virtual) forStatement(statement *ast.ForStatement) any {
	v.evalInto(statement.Test)
	v.evalInto(statement.Update)
	v.evalInto(statement.Initializer)
	switch statement.Body.(type) {
	case *ast.BlockStatement:
		for _, expression := range statement.Body.(*ast.BlockStatement).List {
			v.evalInto(expression)
		}
	default:
		return v.evalInto(statement.Body)
	}
	return nil
}

func (v *Virtual) forInStatement(statement *ast.ForInStatement) any {
	v.evalInto(statement.Source)
	switch statement.Body.(type) {
	case *ast.BlockStatement:
		for _, expression := range statement.Body.(*ast.BlockStatement).List {
			v.evalInto(expression)
		}
	default:
		return v.evalInto(statement.Body)
	}
	return nil
}

func (v *Virtual) bracketStatement(expression *ast.BracketExpression) any {
	return v.evalInto(expression.Member)
}

func (v *Virtual) whileStatement(statement *ast.WhileStatement) any {
	switch statement.Body.(type) {
	case *ast.BlockStatement:
		v.blockStatement(statement.Body.(*ast.BlockStatement))
	default:
		return v.evalInto(statement.Body)
	}
	return nil
}

func (v *Virtual) arrayStatement(statement *ast.ArrayLiteral) any {
	for _, expression := range statement.Value {
		v.evalInto(expression)
	}
	return nil
}

func (v *Virtual) blockStatement(statement *ast.BlockStatement) any {
	var lastValue string
	result, err := v.runInVm(fmt.Sprintf("%s[%s.length-1]", v.listName, v.listName))
	if err != nil {
		panic(err)
	}
	lastValue = result.String()

	var pushValue string
	var valueName string

	var ok goja.Value
	for _, expression := range statement.List {
		if pushValue != "" {
			if ok, err = v.runInVm(fmt.Sprintf("%s[%s.length-1] !== %s", v.listName, v.listName, valueName)); err == nil && ok.ToBoolean() {
				_, _ = v.runInVm(pushValue)
			}
		}

		var value = v.evalInto(expression)
		switch value.(type) {
		case string:
			pushValue = value.(string)
			if pushValue == "" {
				continue
			}
			find := pushValueRegex.FindStringSubmatch(pushValue)
			if len(find) < 3 {
				continue
			}
			valueName = find[2]
		}
	}

	_, _ = v.runInVm(fmt.Sprintf("%s.push(%s)", v.listName, lastValue))

	return nil
}

func (v *Virtual) assigmentExpression(statement *ast.AssignExpression) any {
	switch statement.Left.(type) {
	case *ast.CallExpression, *ast.BracketExpression, *ast.ArrayPattern:
		v.evalInto(statement.Left)
	}
	return v.evalInto(statement.Right)
}

func (v *Virtual) sequenceExpression(statement *ast.SequenceExpression) any {
	for _, expression := range statement.Sequence {
		v.evalInto(expression)
	}
	return nil
}

func (v *Virtual) evalInto(anything ...any) any {
	node := anything[0]

	switch node.(type) {
	case *ast.VariableStatement:
		variable := node.(*ast.VariableStatement)
		if int(variable.Idx0()) > v.maxIndex {
			return nil
		}

		for _, e := range variable.List {
			if _, ok := e.Initializer.(*ast.FunctionLiteral); ok {
				v.evalInto(e, e.Target.(*ast.Identifier).Name.String())
			} else {
				v.evalInto(e)
			}
		}

	case *ast.Binding:
		expression := node.(*ast.Binding)
		if int(expression.Idx0()) > v.maxIndex {
			return nil
		}
		var callExpression any
		if len(anything) > 1 {
			callExpression = anything[1]
		}
		v.evalInto(expression.Target, callExpression)
		v.evalInto(expression.Initializer, callExpression)

	case *ast.ExpressionStatement:
		return v.evalInto(node.(*ast.ExpressionStatement).Expression)

	case *ast.FunctionDeclaration:
		function := node.(*ast.FunctionDeclaration)
		if int(function.Idx0()) > v.maxIndex {
			return nil
		}
		if v.functionBigOperationName == "" {
			if expression, ok := function.Function.Body.List[0].(*ast.ExpressionStatement); ok {
				if sequence, ok := expression.Expression.(*ast.SequenceExpression); ok {
					if len(sequence.Sequence) > 500 {
						v.functionBigOperationName = function.Function.Name.Name.String()
					}
				}
			}
		}

		for _, expression := range function.Function.ParameterList.List {
			v.evalInto(expression)
		}
		return v.evalInto(function.Function.Body)

	case *ast.FunctionLiteral:
		function := node.(*ast.FunctionLiteral)
		for _, expression := range function.ParameterList.List {
			v.evalInto(expression)
		}
		if len(function.Body.List) == 1 && len(anything) > 1 {
			if value, ok := anything[1].(string); ok {
				addFunctionToOperations(function, value)
			}
		}
		return v.evalInto(function.Body)

	case *ast.IfStatement:
		return v.ifStatement(node.(*ast.IfStatement))

	case *ast.TryStatement:
		return v.evalInto(node.(*ast.TryStatement).Body)

	case *ast.CallExpression:
		expression := node.(*ast.CallExpression)
		for _, arg := range expression.ArgumentList {
			v.evalInto(arg, expression)
		}
		return v.evalInto(expression.Callee, expression)

	case *ast.ObjectLiteral:
		object := node.(*ast.ObjectLiteral)
		for _, value := range object.Value {
			v.evalInto(value)
		}

	case *ast.ForStatement:
		return v.forStatement(node.(*ast.ForStatement))

	case *ast.ForInStatement:
		return v.forInStatement(node.(*ast.ForInStatement))

	case *ast.DotExpression:
		var parentExpression any
		if len(anything) > 1 {
			parentExpression = anything[1]
		}
		return v.dotStatement(node.(*ast.DotExpression), parentExpression)

	case *ast.SwitchStatement:
		return v.switchStatement(node.(*ast.SwitchStatement))

	case *ast.DoWhileStatement:
		v.evalInto(node.(*ast.DoWhileStatement).Body)

	case *ast.WhileStatement:
		v.evalInto(node.(*ast.WhileStatement).Body)

	case *ast.BlockStatement:
		return v.blockStatement(node.(*ast.BlockStatement))

	case *ast.ArrayLiteral:
		return v.arrayStatement(node.(*ast.ArrayLiteral))

	case *ast.AssignExpression:
		return v.assigmentExpression(node.(*ast.AssignExpression))

	case *ast.ReturnStatement:
		return v.evalInto(node.(*ast.ReturnStatement).Argument)

	case *ast.SequenceExpression:
		return v.sequenceExpression(node.(*ast.SequenceExpression))

	case *ast.ConditionalExpression:
		expression := node.(*ast.ConditionalExpression)
		v.evalInto(expression.Test)
		v.evalInto(expression.Consequent)
		v.evalInto(expression.Alternate)
		return nil

	case *ast.BinaryExpression:
		expression := node.(*ast.BinaryExpression)
		v.evalInto(expression.Left)
		v.evalInto(expression.Right)

	case *ast.ArrayPattern:
		expression := node.(*ast.ArrayPattern)
		var callExpression any
		if len(anything) > 1 {
			callExpression = anything[1]
		}
		for _, arg := range expression.Elements {
			v.evalInto(arg, callExpression)
		}

	case *ast.UnaryExpression:
		expression := node.(*ast.UnaryExpression)
		var callExpression any
		if len(anything) > 1 {
			callExpression = anything[1]
		}
		v.evalInto(expression.Operand, callExpression)

	case *ast.PrivateDotExpression:
		expression := node.(*ast.PrivateDotExpression)
		var callExpression any
		if len(anything) > 1 {
			callExpression = anything[1]
		}
		return v.evalInto(expression.Left, callExpression)

	case *ast.NewExpression:
		expression := node.(*ast.NewExpression)
		v.evalInto(expression.Callee)
		for _, arg := range expression.ArgumentList {
			v.evalInto(arg)
		}

	case *ast.ThrowStatement:
		expression := node.(*ast.ThrowStatement)
		v.evalInto(expression.Argument)

	case *ast.BracketExpression:
		expression := node.(*ast.BracketExpression)
		v.evalInto(expression.Member, expression)
		v.evalInto(expression.Left, expression)

	case *ast.ArrowFunctionLiteral:
		expression := node.(*ast.ArrowFunctionLiteral)
		v.evalInto(expression.ParameterList.Rest)
		for _, arg := range expression.ParameterList.List {
			v.evalInto(arg)
		}

	case *ast.PropertyKeyed:
		expression := node.(*ast.PropertyKeyed)
		v.evalInto(expression.Value)

	case *ast.ForLoopInitializerVarDeclList:
		expression := node.(*ast.ForLoopInitializerVarDeclList)
		for _, arg := range expression.List {
			v.evalInto(arg)
		}

	case *ast.ForLoopInitializerExpression:
		expression := node.(*ast.ForLoopInitializerExpression)
		v.evalInto(expression.Expression)

	case *ast.Identifier, *ast.StringLiteral, *ast.NumberLiteral, *ast.EmptyStatement,
		*ast.BranchStatement, *ast.NullLiteral, *ast.ThisExpression, *ast.RegExpLiteral, *ast.LexicalDeclaration, nil:
		break
	}

	return nil
}
