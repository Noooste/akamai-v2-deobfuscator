package deobfuscator

import (
	"bytes"
	"regexp"
	"strconv"
)

var mainFunctionName = regexp.MustCompile(`^\(function (\w+)\(\){`)
var hexReg = regexp.MustCompile(`(0x[a-fA-F0-9]+),`)
var seedReg = regexp.MustCompile(`\w+=\w+\(\w+,(\d{5,})\)`)

const (
	c132 uint32 = 0xcc9e2d51
	c232 uint32 = 0x1b873593
)

func mmh3(input []byte, seed uint32) uint32 {
	var index uint32 = 0

	for _, b := range input {
		if b == 10 || b == 13 || b == 32 {
			continue
		}
		char := uint32(b)
		char = (char&0xffff)*c132 + (((char>>16)*c132&0xffff)<<16)&0xffffffff
		char = (char<<15 | char>>17) & 0xffffffff
		char = (char&0xffff)*c232 + (((char>>16)*c232&0xffff)<<16)&0xffffffff
		seed ^= char
		seed = (seed<<13 | seed>>19) & 0xffffffff
		dwl := (seed&0xffff)*5 + (((seed>>16)*5&0xffff)<<16)&0xffffffff
		seed = (dwl & 0xffff) + 0x6b64 + (((dwl >> 16) + 0xe654&0xffff) << 16)
		index += 1
	}

	seed ^= index
	seed ^= seed >> 16
	seed = (seed&0xffff)*0x85ebca6b + (((seed>>16)*0x85ebca6b&0xffff)<<16)&0xffffffff
	seed ^= seed >> 13
	seed = (seed&0xffff)*0xc2b2ae35 + (((seed>>16)*0xc2b2ae35&0xffff)<<16)&0xffffffff
	seed ^= seed >> 16
	return seed & 0xffffffff
}

func findAndReplaceScriptHashValue(script []byte) []byte {
	var functionName string
	if t := mainFunctionName.FindSubmatch(script); t != nil {
		functionName = string(t[1])
	} else {
		return script
	}

	var scriptHashValue = regexp.MustCompile(`(\w+)=\w+\(\w+\(` + functionName + `\),"` + functionName + `","[x\d\\]+"\)`)

	function := script[1 : len(script)-4]

	hexList := hexReg.FindSubmatch(function)
	if len(hexList) == 0 {
		return script
	}

	hex := hexList[1]

	seed := seedReg.FindSubmatch(function)[1]

	var index1 = bytes.Index(function, hex)
	var index2 = index1 + bytes.Index(function[index1:], []byte(";"))

	var builder = bytes.Buffer{}
	builder.Grow(len(function) - len(seed) + 10)

	builder.Write(function[:index1])
	builder.Write(function[index2+1:])
	builder.WriteString("undefined")

	seedValue, _ := strconv.ParseUint(string(seed), 10, 32)

	var number, _ = strconv.ParseUint(string(function[len(hex)+index1+1:index2]), 10, 32)
	var hashValue = mmh3(builder.Bytes(), uint32(seedValue))

	result := uint32(number) - hashValue

	return scriptHashValue.ReplaceAllFunc(script, func(s []byte) []byte {
		return []byte(string(bytes.Split(s, []byte("="))[0]) + "=" + strconv.FormatUint(uint64(result), 10))
	})
}
