# Akamai V2 Deobfuscator

This project aims to reverse Akamai script obfuscation.

## Use

Deob script :

```go
import (
	http "github.com/Noooste/fhttp"
	"github.com/Noooste/go-utils"
	"log"
	"os"
	"regexp"
)

func main() {
	var response, _ = http.Get("https://www.nike.com/h8r6ElR8B4Q6OG-YC53dZdAB1hU/7wacrNpthiat/RX44Qw/dT1rJV/RfAxY")
	script := string(utils.GetResponseBody(response))

	windowName := regexp.MustCompile(`(\w+)=window`).FindStringSubmatch(script)[1]

	var err error
	var v *Virtual
	script, v, err = runMainFunction(script, true)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	v.deobedScript = script

	if err = v.deob(false); err != nil {
		log.Fatalf("Error: %s", err)
	}

	script = CleanFinalScript(v.deobedScript, windowName, true)

	_ = os.WriteFile("output.js", []byte(script), 0644)
}
```

Get script ID and DeviceData key order (only on dynamic scripts) : 
```go
import (
	"github.com/Noooste/akamai-v2-deobfuscator"
	"github.com/Noooste/go-utils"
	"log"
	"time"
)

func main() {
	var response, _ = http.Get("https://www.ihg.com/pjYhi/LLw3p/IAq-/-m5-/tv/X9YYbpGSaJS9/MngDAQ/dVogXDE/wRFA")
	script := utils.GetResponseBody(response)

	var id, order = deobfuscator.GetScriptInformation(script)
	now := time.Since(start)

	log.Println("script id : ", id)
	log.Println("device data order :", order)
}
```

Get rotate function :

```go
import (
	"github.com/Noooste/akamai-v2-deobfuscator"
	http "github.com/Noooste/fhttp"
	"github.com/Noooste/go-utils"
	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"log"
	"os"
	"runtime/debug"
)


func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Print(r)
			debug.PrintStack()
		}
	}()

	var response, _ = http.Get("https://www.nike.com/h8r6ElR8B4Q6OG-YC53dZdAB1hU/7wacrNpthiat/RX44Qw/dT1rJV/RfAxY")
	script := utils.GetResponseBody(response)

	rf, _, _, _ := deobfuscator.GetRotateFunction(script)

	beautified, _ := jsbeautifier.Beautify(&rf, jsbeautifier.DefaultOptions())
	os.WriteFile("rotate_function.js", []byte(beautified), 0644)
}
```

