// 版权 @2023 wat-go 作者。保留所有权利。

package wat2ll

import (
	"fmt"
	"os"
	"strings"

	"github.com/chai2010/wat-go/pkg/3rdparty/cli"
	"github.com/chai2010/wat-go/pkg/wat2ll"
)

var CmdWat2ll = &cli.Command{
	Hidden:    false,
	Name:      "2ll",
	Usage:     "convert a WebAssembly text file to a llvm-ir file",
	ArgsUsage: "<file.wat>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "set code output file",
			Value:   "_a.out.ll",
		},
	},
	Action: func(c *cli.Context) error {
		if c.NArg() == 0 {
			fmt.Fprintf(os.Stderr, "no input file")
			os.Exit(1)
		}

		infile := c.Args().First()
		outfile := c.String("output")

		if outfile == "" {
			outfile = infile
			if n1, n2 := len(outfile), len(".wat"); n1 > n2 {
				if s := outfile[n1-n2:]; strings.EqualFold(s, ".wat") {
					outfile = outfile[:n1-n2]
				}
			}
			outfile += ".ll"
		}
		if !strings.HasSuffix(outfile, ".ll") {
			outfile += ".ll"
		}

		source, err := os.ReadFile(infile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		code, err := wat2ll.Wat2LL(infile, source)
		if err != nil {
			os.WriteFile(outfile, code, 0666)
			fmt.Println(err)
			os.Exit(1)
		}

		err = os.WriteFile(outfile, code, 0666)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		return nil
	},
}
