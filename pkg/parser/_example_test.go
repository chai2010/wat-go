// 版权 @2024 wat-go 作者。保留所有权利。

package parser_test

import (
	"fmt"
	"log"
	"os"

	"github.com/chai2010/wat-go/pkg/parser"
	"wa-lang.org/wa/api"
)

func ExampleParseModule() {
	src := `(module $hello)`

	m, err := parser.ParseModule("hello.wat", []byte(src))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(m)
	fmt.Println(m.Name)

	// output:
	// (module $hello)
	// hello
}

func ExampleParseModule_wa() {
	const filename = "a.out.wa"
	const src = `func main { println("hello wa") }`

	_, watBytes, err := api.BuildFile(api.DefaultConfig(), filename, src)
	if err != nil {
		log.Fatal(err)
	}

	m, err := parser.ParseModule("a.out.wat", watBytes)
	if err != nil {
		os.WriteFile("a.out.wat", watBytes, 0666)
		fmt.Println(err)
		return
	}

	fmt.Println(m.Name)

	// output:
	// __walang__
}
