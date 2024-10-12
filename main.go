// 版权 @2024 wat-go 作者。保留所有权利。

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chai2010/wat-go/cmd/wat2c"
	"github.com/chai2010/wat-go/cmd/wat2ll"
	"github.com/chai2010/wat-go/cmd/wat2wasm"
	"github.com/chai2010/wat-go/cmd/watstrip"
	"github.com/chai2010/wat-go/cmd/watvalidate"
	"github.com/chai2010/wat-go/pkg/3rdparty/cli"
)

func main() {
	cliApp := cli.NewApp()
	cliApp.Name = "wat-go"
	cliApp.Usage = "wat-go is a tool for managing Wat source code."
	cliApp.Copyright = "Copyright 2024 The wat-go Authors. All rights reserved."
	cliApp.Version = "0.1.0"
	cliApp.EnableBashCompletion = true
	cliApp.HideHelpCommand = true

	cliApp.CustomAppHelpTemplate = cli.AppHelpTemplate +
		"\nSee \"https://github.com/chai2010/wat-go\" for more information.\n"

	// 没有参数时显示 help 信息
	cliApp.Action = func(c *cli.Context) error {
		if c.NArg() > 0 {
			fmt.Println("unknown command:", strings.Join(c.Args().Slice(), " "))
			os.Exit(1)
		}
		cli.ShowAppHelpAndExit(c, 0)
		return nil
	}

	cliApp.Commands = []*cli.Command{
		wat2c.CmdWat2c,             // 翻译到C语言
		wat2ll.CmdWat2ll,           // 翻译到 LLVM-IR(TODO)
		wat2wasm.CmdWat2Wasm,       // 翻译到 WebAssembly 二进制格式 (TODO)
		watstrip.CmdWatStrip,       // 删除未使用对象
		watvalidate.CmdWatValidate, // 验证 WebAssembly txt 文件(TODO)

		// TODO: objdump, fmt, run, ...
	}

	cliApp.Run(os.Args)
}
