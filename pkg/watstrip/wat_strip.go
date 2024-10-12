// 版权 @2024 wat-go 作者。保留所有权利。

package watstrip

import (
	"bytes"

	"github.com/chai2010/wat-go/pkg/parser"
	"github.com/chai2010/wat-go/pkg/printer"
)

func WatStrip(path string, src []byte) (watBytes []byte, err error) {
	m, err := parser.ParseModule(path, src)
	if err != nil {
		return nil, err
	}

	// 删除未使用对象
	pass := new_RemoveUnusedPass(m)
	m = pass.DoPass()

	var buf bytes.Buffer
	if err := printer.FprintIndent(&buf, m, "  "); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
