// 版权 @2023 wat-go 作者。保留所有权利。

package watvalidate

import (
	"fmt"
	"os"

	"github.com/chai2010/wat-go/pkg/3rdparty/cli"
	"github.com/chai2010/wat-go/pkg/watvalidate"
)

var CmdWatValidate = &cli.Command{
	Hidden:    false,
	Name:      "validate",
	Usage:     "validate a file in the WebAssembly txt format",
	ArgsUsage: "<file.wat>",
	Action: func(c *cli.Context) error {
		if c.NArg() == 0 {
			fmt.Fprintf(os.Stderr, "no input file")
			os.Exit(1)
		}

		infile := c.Args().First()

		source, err := os.ReadFile(infile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fn, ins, err := watvalidate.WatValidate(infile, source)
		if err != nil {
			if fn != nil {
				fmt.Println("func:", fn.Name)
			}
			for i, ins := range ins {
				fmt.Printf("  [%d] %v\n", i, ins)
			}
			fmt.Println(err)
			os.Exit(1)
		}

		return nil
	},
}
