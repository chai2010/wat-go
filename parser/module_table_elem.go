// 版权 @2024 wat-go 作者。保留所有权利。

package parser

import (
	"github.com/chai2010/wat-go/ast"
	"github.com/chai2010/wat-go/token"
)

// elem ::= (elem id? elemlist)
//       |  (elem id? x:tableuse (offset e:expr) elemlist)
//       |  (elem id? declare elemlist)
//
// elemlist ::= reftype vec(elemexpr)
// elemexpr ::= (item e:expr)
// tableuse ::= (table x:tableidx)

// (elem (i32.const 1) $$u8.$$block.$$onFree)
func (p *parser) parseModuleSection_elem() *ast.ElemSection {
	p.acceptToken(token.ELEM)

	elemSection := &ast.ElemSection{}

	if p.tok == token.IDENT {
		elemSection.Name = p.parseIdent()
	}

	p.acceptToken(token.LPAREN)
	p.acceptToken(token.INS_I32_CONST)
	elemSection.Offset = uint32(p.parseIntLit())
	p.acceptToken(token.RPAREN)

	elemSection.Values = p.parseIdentOrIndexList()

	return elemSection
}
