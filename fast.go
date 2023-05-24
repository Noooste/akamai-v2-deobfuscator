package deobfuscator

import (
	"github.com/Noooste/go-utils"
	"github.com/dop251/goja"
	"regexp"
	"strings"
	"sync"
)

var deviceIdReg = regexp.MustCompile(`var \w+=function\(\){\w+\+\+,\w+=\w+(\.call|\.apply|\w+\.)?\((([!\s]*)(null|\w+|\((\w+(\((\w+|{}|\[])\))?|\{}|\[])\)|[\[\]]|[{}]|,|\s))+\);};(\w+)\.push\((\w+)\);`)
var returnReg = regexp.MustCompile(`return \w+(\.(call|apply))?\(this,\w+\);.*$`)
var beginningReg = regexp.MustCompile(`^\(function(\s\w+)?\(\){`)

func GetScriptInformation(script []byte) (string, []any) {
	script, vm := CleanScriptAndRun(script)

	lock := &sync.Mutex{}
	runInVm := func(str string) (goja.Value, error) {
		lock.Lock()
		defer lock.Unlock()
		return vm.RunString(str)
	}

	order := make(chan []any)
	id := make(chan string)

	utils.SafeGoRoutine(func() {
		order <- GetDevice(script, runInVm)
	})

	utils.SafeGoRoutine(func() {
		id <- GetId(script, runInVm)
	})

	return <-id, <-order
}

func CleanScriptAndRun(script []byte) ([]byte, *goja.Runtime) {
	script = findAndReplaceScriptHashValue(script)

	script = returnReg.ReplaceAllFunc(script, func(match []byte) []byte {
		return match[7:]
	})

	script = beginningReg.ReplaceAll(script, []byte(""))
	script = script[:len(script)-5]

	vm := goja.New()
	_, _ = vm.RunString(tryPrefix + string(script) + trySuffix)

	return script, vm
}

var scriptIdUsageReg = regexp.MustCompile(`\),\w+\),\w+=\w+(\(\w+)+,(\w+)\),\w+\)`)

func GetId(script []byte, runInVm func(string) (goja.Value, error)) string {
	pushIndex := deviceIdReg.FindSubmatch(script)
	length := len(pushIndex)

	pushName := string(pushIndex[length-2])
	pushValue := string(pushIndex[length-1])

	scriptIdUsage := scriptIdUsageReg.FindSubmatch(script)
	scriptIdVarName := string(scriptIdUsage[len(scriptIdUsage)-1])
	scriptId := regexp.MustCompile(`(var |,)` + scriptIdVarName + "=(.*?);")

	value, err := runInVm(pushName + " = [" + pushValue + "];" + string(scriptId.FindSubmatch(script)[2]))

	if err != nil {
		return ""
	}

	return value.String()
}

var inDeviceFunctionReg = regexp.MustCompile(`function\(\){\w+\.push\(\w+\);var (\w+)=\[];`)
var shuffleDeviceReg = regexp.MustCompile(`return \[\w+\[\w+],`)
var endShuffleDeviceReg = regexp.MustCompile(`,\w+=\((\w+)`)

func GetDevice(script []byte, runInVm func(string) (goja.Value, error)) []any {
	var deviceData = make(chan string)
	var deviceOrder = make(chan string)
	var listName string

	indexes := shuffleDeviceReg.FindIndex(script)

	if indexes == nil {
		return nil
	}

	utils.SafeGoRoutine(func() {
		index := inDeviceFunctionReg.FindIndex(script)
		searchZone := script[index[1]:]

		pushes := regexp.MustCompile(`\(\[(\w+),[\w,]+]\)`)

		matches := pushes.FindAllSubmatch(searchZone, 19)
		var device = make([]string, len(matches))
		for i, match := range matches {
			device[i] = string(match[1])
		}

		deviceData <- "[" + strings.Join(device, ",") + "]"
	})

	utils.SafeGoRoutine(func() {
		start := indexes[0]
		endGroup := endShuffleDeviceReg.FindSubmatchIndex(script[start:])
		end := start + endGroup[0]
		listName = string(script[start+endGroup[2] : start+endGroup[3]])
		deviceOrder <- string(script[start+7 : end])
	})

	v, _ := runInVm("var " + listName + " = " + <-deviceData + ";" + <-deviceOrder)

	return v.Export().([]any)
}
