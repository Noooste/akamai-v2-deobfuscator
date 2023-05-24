package deobfuscator

import (
	"regexp"
	"strings"
)

var varSequenceReg = regexp.MustCompile(`var ([\w\d]+,?){30,};`)
var varSequenceJSFuckReg = regexp.MustCompile(`([\w\d]+)=[!+\-\[\]\s]+[,;]`)
var checkOperation = regexp.MustCompile(`[!+\-]`)

func (v *Virtual) evalInt() {
	function := regexp.MustCompile(`function ` + v.functionBigOperationName + `\(\){[,=*+\w\d]+;}`).FindString(v.deobedScript)
	functionIndex := strings.Index(v.deobedScript, function)
	finalScript := (v.deobedScript)[:functionIndex] + (v.deobedScript)[functionIndex+len(function):]

	varLines := varSequenceReg.FindAllString(finalScript, 2)
	line1Index := strings.Index(finalScript, varLines[0])
	line1Length := len(varLines[0])
	line2Index := strings.Index(finalScript, varLines[1])
	line2Length := len(varLines[1])

	finalScript = finalScript[:line1Index] + finalScript[line1Index+line1Length:line2Index] + finalScript[line2Index+line2Length:]

	for _, jsFuck := range varSequenceJSFuckReg.FindAllStringSubmatch(finalScript, -1) {
		if checkOperation.MatchString(jsFuck[0]) {
			varLines = append(varLines, jsFuck[1])
		}
	}

	for _, line := range varLines {
		for _, variable := range strings.Split(line, ",") {
			result, err := v.vm.RunString(variable)
			if err != nil {
				panic(err)
			}
			resultString := result.String()
			finalScript = regexp.MustCompile(`\b`+variable+`\b`).ReplaceAllLiteralString(finalScript, resultString)
		}
	}

	v.deobedScript = finalScript
}
