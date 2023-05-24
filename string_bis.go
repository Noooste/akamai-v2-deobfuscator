package deobfuscator

import (
	"encoding/json"
	"regexp"
	"strings"
)

func (v *Virtual) deobString(fast bool) {
	if v.listName == "" {
		v.listName = pushValueRegex.FindStringSubmatch(v.script)[1]
		v.runInVm("var " + v.listName + "=[];")
	}

	var evilRegex = regexp.MustCompile(`(\w+\.\w+(\.call|\.apply|)\((([!\s]*)(null|\w+|,|\((\w+(\((\w+|{}|\[])\))?|\{}|\[])\)|[\[\]]|[{}]|\s))+\))|(` + v.listName + `\.(push|pop\(\w+\)))`)

	var lastWasPop bool

	v.deobedScript = string(evilRegex.ReplaceAllFunc([]byte(v.deobedScript), func(match []byte) []byte {
		if lastWasPop && strings.HasPrefix(string(match), v.listName+".pop") {
			return match
		}

		if strings.HasPrefix(string(match), v.listName+".pop") {
			lastWasPop = true
		} else {
			lastWasPop = false
		}

		r, _ := v.runInVm(string(match))

		if !strings.HasPrefix(string(match), v.listName+".push") && !strings.HasPrefix(string(match), v.listName+".pop") {
			var r2, _ = json.Marshal(r)
			return r2
		}

		return match
	}))

	if !fast {
		v.evalInt()
		replaceOperations(&v.deobedScript)
	}
}
