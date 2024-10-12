// 版权 @2024 wat-go 作者。保留所有权利。

package scanner_test

import (
	"fmt"

	"github.com/chai2010/wat-go/scanner"
	"github.com/chai2010/wat-go/token"
)

func ExampleScanner_Scan() {
	var src = []byte("(module $__walang__)")
	var file = token.NewFile("", len(src))

	var s scanner.Scanner
	s.Init(file, src, nil, scanner.ScanComments)

	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		fmt.Printf("%s %q\n", tok, lit)
	}

	// output:
	// ( ""
	// module "module"
	// IDENT "__walang__"
	// ) ""
}
