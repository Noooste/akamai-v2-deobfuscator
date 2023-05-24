package deobfuscator

import (
	"fmt"
	"github.com/dop251/goja"
	"runtime/debug"
	"time"
)

type RotateFunction struct {
	DeviceDataOrder []string
	ScriptId        string

	vm *goja.Runtime

	LastCall time.Time
}

func GetRotateFunction(script []byte) (function string, hash string, err error, errors []Error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			debug.PrintStack()
		}
	}()

	scriptReduced, vm := CleanScriptAndRun(script)

	return ParseRotateFunction(scriptReduced, vm.RunString)
}

func (rf *RotateFunction) GetResult(startTimestamp int, deviceData, mouseMoveData string, totVel, deltaTimestamp int) (string, error) {
	rf.LastCall = time.Now()
	var information = fmt.Sprintf(
		`{"startTimestamp": %d, "deviceData": "%s", "mouseMoveData": "%s", "totVel": %d, "deltaTimestamp": %d}`,
		startTimestamp, deviceData, mouseMoveData, totVel, deltaTimestamp)

	result, err := rf.vm.RunString(`rf()(` + information + `);`)

	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (rf *RotateFunction) GetScriptId() string {
	rf.LastCall = time.Now()
	return rf.ScriptId
}
