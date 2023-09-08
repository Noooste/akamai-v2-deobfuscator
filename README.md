# Akamai V2 Deobfuscator

### :warning: THIS PROJECT IS NOT STABLE AND NOT WORKING 100%. ONLY NEED TO BE USE AS EDUCATIONAL PURPOSE
This project aims to reverse Akamai script obfuscation.


## Use

Deob script :

```go
package main

import (
	http "github.com/Noooste/fhttp"
	"github.com/Noooste/go-utils"
    	"github.com/Noooste/akamai-v2-deobfuscator"
	"os"
)

func main() {
    var response, _ = http.Get("https://www.nike.com/h8r6ElR8B4Q6OG-YC53dZdAB1hU/7wacrNpthiat/RX44Qw/dT1rJV/RfAxY")
    
    script := deobfuscator.Deob(utils.GetResponseBody(response))
    
    _ = os.WriteFile("output.js", script, 0644)
}
```

Get script ID and DeviceData key order (only on dynamic scripts) : 
```go
package main

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

	log.Println("script id : ", id)
	log.Println("device data order :", order)
}
```

Get rotate function :

```go
package main

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

