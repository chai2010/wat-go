// 版权 @2023 凹语言 作者。保留所有权利。

package watstrip

import (
	"fmt"
	"os"

	"github.com/chai2010/wat-go/pkg/3rdparty/cli"
	"github.com/chai2010/wat-go/pkg/watstrip"
)

var CmdWatStrip = &cli.Command{
	Hidden:    false,
	Name:      "strip",
	Usage:     "remove unused func and global in WebAssembly text file",
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
		watBytes, err := watstrip.WatStrip(infile, source)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		os.Stdout.Write(watBytes)
		return nil
	},
}
