package deobfuscator

import (
	http "github.com/Noooste/fhttp"
	"github.com/Noooste/go-utils"
	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"testing"
	"time"
)

func TestScriptId(t *testing.T) {
	var script []byte

	if content, err := ioutil.ReadFile("input.js"); err != nil {
		var response, _ = http.Get("https://www.ihg.com/pjYhi/LLw3p/IAq-/-m5-/tv/X9YYbpGSaJS9/MngDAQ/dVogXDE/wRFA")
		script = utils.GetResponseBody(response)
	} else {
		script = content
	}

	start := time.Now()

	var id, order = GetScriptInformation(script)
	now := time.Since(start)

	log.Print("Finish in ", now)

	log.Print(id)
	log.Print(order)
}
func TestScriptId2(t *testing.T) {
	var script []byte

	var response, _ = http.Get("https://www.nike.com/h8r6ElR8B4Q6OG-YC53dZdAB1hU/7wacrNpthiat/RX44Qw/dT1rJV/RfAxY")
	script = utils.GetResponseBody(response)

	start := time.Now()

	var id, order = GetScriptInformation(script)
	now := time.Since(start)

	log.Print("Finish in ", now)

	log.Print(id)
	log.Print(order)
}

func TestScriptDeobRotateFunction(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Print(r)
			debug.PrintStack()
		}
	}()

	var script []byte

	var response, _ = http.Get("https://www.nike.com/h8r6ElR8B4Q6OG-YC53dZdAB1hU/7wacrNpthiat/RX44Qw/dT1rJV/RfAxY")
	script = utils.GetResponseBody(response)

	rf, _, _, _ := GetRotateFunction(script)

	time.Sleep(100 * time.Millisecond)

	beautified, _ := jsbeautifier.Beautify(&rf, jsbeautifier.DefaultOptions())
	os.WriteFile("rotate_function.js", []byte(beautified), 0644)
}

func TestScriptDeob(t *testing.T) {
	var response, _ = http.Get("https://www.nike.com/h8r6ElR8B4Q6OG-YC53dZdAB1hU/7wacrNpthiat/RX44Qw/dT1rJV/RfAxY")

	script := Deob(utils.GetResponseBody(response))

	_ = os.WriteFile("output.js", script, 0644)
}
