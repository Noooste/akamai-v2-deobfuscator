package deobfuscator

import (
	"io/ioutil"
	"log"
	"os"
)

var SavePath = "./test"

func GetSavedFunction(hash string) string {
	_ = os.Mkdir(SavePath, 0666)
	if _, err := os.Stat(SavePath + "/" + hash); err == nil {
		content, _ := ioutil.ReadFile(SavePath + "/" + hash)
		return string(content)
	}
	return ""
}

func SaveFunction(hash, script string) {
	_ = os.Mkdir(SavePath, 0666)
	if _, err := os.Stat(SavePath + "/" + hash); err != nil {
		_ = ioutil.WriteFile(SavePath+"/"+hash, []byte(script), 0666)
	} else {
		log.Print(err)
	}
}
