package deobfuscator

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"log"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
)

type rotateFunctionStruct struct {
	script        []byte
	scriptReduced string

	functionName string
	function     string

	pusherName string
	objectName string

	windowName string

	webhookRotateFunction string
	operationFunctions    map[string][2]string

	nativeFunction string
	runInVm        func(string) (goja.Value, error)
}

var mathMethods = []string{
	"E", "LN10", "LN2", "LOG10E", "LOG2E", "PI", "SQRT1_2", "SQRT2", "abs", "acos", "acosh", "asin", "asinh", "atan",
	"atan2", "atanh", "cbrt", "ceil", "clz32", "cos", "cosh", "exp", "expm1", "floor", "fround", "hypot", "imul", "log",
	"log10", "log1p", "log2", "max", "min", "pow", "random", "round", "sign", "sin", "sinh", "sqrt", "tan", "tanh", "trunc",
}

func contains(element string, list []string) bool {
	for _, el := range list {
		if el == element {
			return true
		}
	}
	return false
}

var isWord = regexp.MustCompile(`^([;|,:]*|\w{2,}|\d+)$`)
var rotateFunctionCallReg = regexp.MustCompile(`(\w{2,}|})\(\)\(`)

func ParseRotateFunction(script []byte, runInVm func(string) (goja.Value, error)) (string, string, error, []Error) {
	index := rotateFunctionCallReg.FindIndex(script)

	if index == nil {
		return "", "", errors.New("rotate function not found"), nil
	}

	var start, end int
	if script[index[0]] == '}' {
		start, end = getFunctionBounds(script, index[0], true)
	} else {
		var name []byte
		var i = index[0]
		for script[i] != '(' {
			name = append(name, script[i])
			i++
		}
		index2 := regexp.MustCompile(`(function ` + string(name) + `\(|` + string(name) + `=)`).FindIndex(script)
		start, end = getFunctionBounds(script, index2[0], false)
	}

	functionMinified := minifyAll.ReplaceAll(script[start:end], []byte{})
	hash := fmt.Sprintf("%x", md5.Sum(functionMinified))

	if value := GetSavedFunction(hash); value != "" {
		return value, hash, nil, nil
	}

	pushIndex := deviceIdReg.FindSubmatch(script)
	length := len(pushIndex)

	pushName := string(pushIndex[length-2])

	windowName := regexp.MustCompile(`(\w+)=window`).FindSubmatch(script)[1]

	var randomSub = regexp.MustCompile(`return \['(\w{2,3})',`).FindSubmatch(script)[1]

	rf := &rotateFunctionStruct{
		script:     script,
		pusherName: pushName,
		function:   "function rf()" + string(script[start:end]),
		windowName: string(windowName),
		runInVm:    runInVm,
		objectName: string(regexp.MustCompile(`(\w+)\.` + string(randomSub)).FindSubmatch(script)[1]),
	}

	function, err, stack := rf.deobRotateFunction()
	if err == nil {
		SaveFunction(hash, function)
	}

	return function, hash, err, stack
}

var minifyAll = regexp.MustCompile(`(\.\w+(\.call|\.apply|)\((([!\s]*)(null|\w+|,|\((\w+(\((\w+|{}|\[])\))?|\{}|\[])\)|[\[\]]|[{}]|\s))+\))|\w{1,3}`)

func (rf *rotateFunctionStruct) deobRotateFunction() (string, error, []Error) {
	rf.getDependencies(map[string]bool{})

	rf.replaceStrings()

	rf.addNativeFunctionsDependencies()

	var err, stack = rf.completeDependencies()

	return "var window;" + rf.function, err, stack
}

func (rf *rotateFunctionStruct) replaceStrings() {
	var pushReg = regexp.MustCompile(rf.pusherName + `\.push\([\w\d]+\);`)
	var allPush = pushReg.FindAllString(rf.function, -1)

	var objectCallReg = regexp.MustCompile(rf.objectName + `\.\w+(\.call|\.apply|)\((([!\s]*)(null|\w+|\((\w+(\((\w+|{}|\[])\))?|\{}|\[])\)|[\[\]]|[{}]|,|\s))+\)`)
	var objectCallReplaced []string

	allCalls := objectCallReg.FindAllString(rf.function, -1)
	objectCallReplaced = make([]string, len(allCalls))

	lock := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(allPush) + 2)

	for _, push := range allPush {
		go func(push string) {
			defer func() {
				if r := recover(); r != nil {
					log.Print(r)
					debug.PrintStack()
				}
			}()

			vm := goja.New()
			_, _ = vm.RunString(tryPrefix + string(rf.script) + trySuffix)

			listName := pushValueRegex.FindSubmatch(rf.script)[1]
			_, _ = vm.RunString("var " + string(listName) + "=[];" + push)

			for i, el := range allCalls {
				if objectCallReplaced[i] != "" {
					continue
				}
				value, err := vm.RunString(el)
				if err != nil {
					panic(err)
				}
				var word = value.String()

				if isWord.MatchString(word) {
					if i > 0 && objectCallReplaced[i-1] == "Math" {
						if !contains(word, mathMethods) {
							continue
						}
					}

					objectCallReplaced[i] = word

					unescaped := `"` + stringUnescape.ReplaceAllStringFunc(word, func(s string) string {
						if strings.Count(s, `\`)%2 == 0 {
							s = `\` + s
						}
						if s[len(s)-1] == '\\' {
							s += `\`
						}

						return s
					}) + `"`

					lock.Lock()
					rf.function = strings.Replace(rf.function, el, unescaped, -1)
					lock.Unlock()
				}
			}

			wg.Done()
		}(push)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Print(r)
				debug.PrintStack()
			}
		}()

		lock.Lock()
		defer lock.Unlock()
		rf.function = pushReg.ReplaceAllString(rf.function, "")
		wg.Done()
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Print(r)
				debug.PrintStack()
			}
		}()

		lock.Lock()
		defer lock.Unlock()
		rf.function = regexp.MustCompile(rf.pusherName+`\.pop\(\)[,;]\s?`).ReplaceAllString(rf.function, "")
		wg.Done()
	}()

	wg.Wait()
}

var functionCallReg = regexp.MustCompile(`(function |\.|)(\w+)\(`)

func (rf *rotateFunctionStruct) getDependencies(functionSet map[string]bool) map[string]bool {
	var allFunctions = functionCallReg.FindAllStringSubmatch(rf.function, -1)
	var script = []byte(rf.script)

	for _, f := range allFunctions {
		if f[1] != "" || f[2] == "if" || f[2] == "for" || f[2] == "while" {
			continue
		}

		var name = f[2]

		if _, ok := functionSet[name]; ok {
			continue
		}

		functionSet[name] = true
	}

	nbFunctions := len(functionSet)
	var functions = make([]string, nbFunctions)
	var i = 0
	var lock = &sync.Mutex{}
	var group = &sync.WaitGroup{}
	group.Add(nbFunctions)

	for name, value := range functionSet {
		if !value {
			group.Done()
			continue
		}
		go func(name string) {
			defer func() {
				if r := recover(); r != nil {
					log.Print(r)
					debug.PrintStack()
				}
			}()

			defer group.Done()
			var index int
			var indexes []int

			if indexes = regexp.MustCompile(`function ` + name + `\(`).FindIndex(rf.script); len(indexes) > 0 {
				index = indexes[0] + 10 + len(name)
			} else if indexes = regexp.MustCompile(name + `=function\(`).FindIndex(rf.script); len(indexes) > 0 {
				index = indexes[0] + 10 + len(name)
			}

			var _, end = getFunctionBounds(script, index, false)
			rf2 := &rotateFunctionStruct{
				script:   rf.script,
				function: `function ` + name + `(` + string(script[index:end]),
			}

			tmp := map[string]bool{}
			for k, _ := range functionSet {
				tmp[k] = false
			}
			rf2.getDependencies(tmp)

			lock.Lock()
			functions[i] = rf2.function
			i++
			lock.Unlock()
		}(name)
	}

	group.Wait()

	for _, function := range functions {
		rf.function = function + rf.function
	}

	return functionSet
}

func (rf *rotateFunctionStruct) addNativeFunctionsDependencies() {
	rf.nativeFunction = fmt.Sprintf(window, "")
	if strings.Contains(rf.function, "btoa") || strings.Contains(rf.function, "atob") {
		rf.nativeFunction += base64
	}
}

func getFunctionBounds(buffer []byte, index int, backward bool) (start, end int) {
	var unclosedParentheses uint8 = 0
	for unclosedParentheses == 0 {
		if buffer[index] == '}' && backward ||
			buffer[index] == '{' && !backward {
			unclosedParentheses++
		}
		if backward {
			index--
		} else {
			index++
		}
	}

	if backward {
		end = index + 2
	} else {
		start = index - 1
	}

	for unclosedParentheses > 0 {
		switch buffer[index] {
		case '{':
			if backward {
				unclosedParentheses--
			} else {
				unclosedParentheses++
			}
			break
		case '}':
			if backward {
				unclosedParentheses++
			} else {
				unclosedParentheses--
			}
			break
		}
		if backward {
			index--
		} else {
			index++
		}
	}

	if backward {
		start = index + 1
	} else {
		end = index
	}

	return
}

var errRefErrName = regexp.MustCompile(`ReferenceError: (\w+) is not defined`)

type Error struct {
	Err    string
	Solved bool
}

func testFunction(rf RotateFunction) error {
	_, err := rf.GetResult(
		4444,
		"181:;:0:;:0:;:0,;,278:;:1440,;,139:;:12147,;,284:;:2256,;,257:;:8105,;,186:;:2560,;,246:;:0,;,119:;:3614607,;,132:;:412159,;,138:;:1096,;,179:;:,;,129:;:1440,;,183:;:Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36,;,224:;:837561806501.5,;,231:;:20030107:;:en-US:;:Gecko:;:5,;,265:;:2256,;,212:;:,cpen:0,i1:0,dm:0,cwen:0,non:1,opc:0,fc:0,sc:0,wrc:1,isc:0,vib:1,bat:1,x11:0,x12:1,;,171:;:0.917187746458,;,198:;:2560",
		"0,1,2875,2054,680;1,1,2881,2055,678;2,1,2888,2057,676;3,1,2895,2059,673;4,1,2901,2063,669;5,1,2907,2066,665;6,1,2914,2070,661;7,1,2920,2072,657;8,1,2927,2074,654;9,1,2933,2075,651;10,1,2939,2074,649;11,1,2945,2073,648;12,1,2951,2070,648;13,1,2957,2066,649;14,1,2964,2062,651;15,1,2971,2057,653;16,1,2978,2053,656;17,1,2985,2048,659;18,1,2992,2044,662;19,1,2999,2041,665;20,1,3006,2038,667;21,1,3012,2036,669;22,1,3018,2034,671;23,1,3024,2033,672;24,1,3031,2032,673;25,1,3037,2031,672;26,1,3043,2031,672;27,1,3050,2032,670;28,1,3057,2033,668;29,1,3064,2035,666;30,1,3070,2037,663;31,1,3077,2039,660;32,1,3084,2042,656;",
		4444,
		4444)
	return err
}

func (rf *rotateFunctionStruct) completeDependencies() (error, []Error) {
	rf.function = CleanFinalScript(rf.function, rf.windowName, false)

	rfTest := RotateFunction{}

	var vm = goja.New()
	rfTest.vm = vm

	_, _ = vm.RunString(rf.nativeFunction + rf.function)

	err := testFunction(rfTest)

	var finalValue goja.Value
	var errorStack []Error

	for err != nil {
		var currentError = Error{
			Err: err.(*goja.Exception).Value().String(),
		}

		if strings.Contains(currentError.Err, "ReferenceError") {
			undefinedName := errRefErrName.FindStringSubmatch(currentError.Err)[1]

			finalValue, err = rf.runInVm(`JSON.stringify(` + undefinedName + `)`)
			if err != nil {
				return errors.New("error when accessing to a variable : " + err.Error()), errorStack
			}

			rf.function = "var " + undefinedName + "=" + finalValue.String() + ";" + rf.function
			rf.webhookRotateFunction = "var " + undefinedName + "=" + finalValue.String() + ";" + rf.webhookRotateFunction

			_, _ = vm.RunString(rf.nativeFunction + rf.function)

		} else {
			return errors.New("non reference error found, error"), errorStack
		}

		err = testFunction(rfTest)
		if err == nil || err.Error() != currentError.Err {
			currentError.Solved = true
			errorStack = append(errorStack, currentError)
			continue
		}

		errorStack = append(errorStack, currentError)
		return errors.New("loop on same error : " + fmt.Sprint(currentError.Err)), errorStack
	}

	return nil, errorStack
}
