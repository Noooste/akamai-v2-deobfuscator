package deobfuscator

import (
	"github.com/Noooste/azuretls-client"
	http "github.com/Noooste/fhttp"
	"github.com/Noooste/go-utils"
	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"log"
	"os"
	"runtime/debug"
	"testing"
	"time"
)

func TestScriptId(t *testing.T) {
	var script []byte

	if content, err := os.ReadFile("input.js"); err != nil {
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

	rf, _, _, _ := GetRotateFunction(script)

	time.Sleep(100 * time.Millisecond)

	beautified, _ := jsbeautifier.Beautify(&rf, jsbeautifier.DefaultOptions())
	os.WriteFile("rotate_function.js", []byte(beautified), 0644)
}
func TestScriptId2(t *testing.T) {
	session := azuretls.NewSession()
	session.Header = http.Header{
		"content-type":       {"text/plain;charset=UTF-8"},
		"sec-ch-ua-mobile":   {"?0"},
		"user-agent":         {utils.UserAgent},
		"sec-ch-ua-platform": {"\"Windows\""},
		"accept":             {"*/*"},
		"sec-fetch-site":     {"same-origin"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-dest":     {"empty"},
		"accept-encoding":    {"gzip, deflate, br"},
		"accept-language":    {"fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7"},
		http.HeaderOrderKey:  {"content-type", "sec-ch-ua", "sec-ch-ua-mobile", "user-agent", "sec-ch-ua-platform", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "accept-encoding", "accept-language"},
	}

	var response, _ = session.Get("https://www.nike.com/46t9UissVWpfej4kcICy/7LatbJJmzO9J5c/dnwZFnQB/W1tgGG/5MU3cB")

	start := time.Now()

	var id, order = GetScriptInformation(response.Body)
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

	session := azuretls.NewSession()
	var response, _ = session.Get("https://www.nike.com/h8r6ElR8B4Q6OG-YC53dZdAB1hU/7wacrNpthiat/RX44Qw/dT1rJV/RfAxY")
	script = response.Body

	rf, _, _, _ := GetRotateFunction(script)

	time.Sleep(100 * time.Millisecond)

	beautified, _ := jsbeautifier.Beautify(&rf, jsbeautifier.DefaultOptions())
	os.WriteFile("rotate_function.js", []byte(beautified), 0644)
}

func TestScriptDeobRotateFunctionInput(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Print(r)
			debug.PrintStack()
		}
	}()

	var script []byte

	if content, err := os.ReadFile("input.js"); err != nil {
		var response, err = http.Get("https://www.ihg.com/pjYhi/LLw3p/IAq-/-m5-/tv/X9YYbpGSaJS9/MngDAQ/dVogXDE/wRFA")
		if err != nil {
			log.Print(err)
			return
		}
		script = utils.GetResponseBody(response)
	} else {
		script = content
	}

	rf, _, _, _ := GetRotateFunction(script)

	time.Sleep(100 * time.Millisecond)

	beautified, _ := jsbeautifier.Beautify(&rf, jsbeautifier.DefaultOptions())
	os.WriteFile("rotate_function_input.js", []byte(beautified), 0644)
}

func TestScriptDeob(t *testing.T) {
	session := azuretls.NewSession()
	var response, err = session.Get("https://www.nike.com/46t9UissVWpfej4kcICy/7LatbJJmzO9J5c/dnwZFnQB/W1tgGG/5MU3cB")

	if err != nil {
		log.Print(err)
		return

	}

	script := Deob(response.Body)

	_ = os.WriteFile("output.js", script, 0644)
}
