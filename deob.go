package deobfuscator

import (
	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"github.com/dop251/goja"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

var Verbose bool
var VerboseTime time.Time

const tryPrefix = `try {`
const trySuffix = `;}catch(e){}`

var cleanObjectCalls = regexp.MustCompile(`\["[$\w_][\w\d_]+"]`)

func (v *Virtual) runInVm(str string) (goja.Value, error) {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.vm.RunString(str)
}

func Deob(script []byte) []byte {
	script = findAndReplaceScriptHashValue(script)

	windowName := regexp.MustCompile(`(\w+)=window`).FindSubmatch(script)[1]

	var err error
	var v *Virtual
	var scriptString string

	scriptString, v, err = runMainFunction(string(script), true)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	v.deobedScript = scriptString

	if err = v.deob(false); err != nil {
		log.Fatalf("Error: %s", err)
	}

	return []byte(CleanFinalScript(v.deobedScript, string(windowName), true))
}

func runMainFunction(script string, fast bool) (string, *Virtual, error) {
	parsed, err := loadScript(script)
	if err != nil {
		return "", nil, err
	}

	var _, vm = CleanScriptAndRun([]byte(script))

	if err != nil {
		return "", nil, err
	}

	virtual := &Virtual{
		vm:           vm,
		lock:         &sync.Mutex{},
		script:       script,
		program:      parsed,
		deobedScript: script,
		deobScript:   true,
	}

	virtual.evalString()

	if !fast {
		virtual.evalInt()
		replaceOperations(&virtual.deobedScript)
	}

	return virtual.deobedScript, virtual, nil
}

func (v *Virtual) deob(fast bool) error {
	var _, vm = CleanScriptAndRun([]byte(v.script))
	v.vm = vm
	v.deobString(fast)
	return nil
}

func removeReturnStatement(script, identifier string, arguments []string, isCall bool) string {
	var expression = identifier
	if isCall {
		expression += ".call"
	}
	expression += "("
	for i, arg := range arguments {
		if i > 0 {
			expression += ","
		}
		expression += arg
	}
	expression += ")"
	return strings.Replace(script, "return "+expression, expression, 1)
}

func RunInVm(script string) (*goja.Runtime, error) {
	vm := goja.New()
	_, err := vm.RunString(tryPrefix + script + trySuffix)
	return vm, err
}

func CleanFinalScript(script string, windowName string, doBeautify bool) string {
	script = cleanObjectCalls.ReplaceAllStringFunc(script, func(s string) string {
		return "." + s[2:len(s)-2]
	})

	wn := regexp.MustCompile(`\b` + windowName + `\b\.`)

	script = wn.ReplaceAllString(script, "")

	if doBeautify {
		script, _ = beautify(&script)
	}

	return script
}

func beautify(src *string) (string, error) {
	options := jsbeautifier.DefaultOptions()
	return jsbeautifier.Beautify(src, options)
}
