// 版权 @2024 wat-go 作者。保留所有权利。

package wat2c

import (
	"fmt"
	"io"

	"github.com/chai2010/wat-go/ast"
	"github.com/chai2010/wat-go/token"
)

func (p *wat2cWorker) buildHeader(w io.Writer) error {
	if len(p.m.Exports) == 0 {
		return nil
	}
	var funcs []*ast.ExportSpec
	for _, e := range p.m.Exports {
		switch e.Kind {
		case token.FUNC:
			funcs = append(funcs, e)
		}
	}
	if len(funcs) == 0 {
		return nil
	}

	for _, e := range funcs {
		fmt.Fprintf(w, "// extern void %s();\n", toCName(e.Name))
	}

	return nil
}
