// 版权 @2024 wat-go 作者。保留所有权利。

package parser

import (
	"github.com/chai2010/wat-go/ast"
	"github.com/chai2010/wat-go/token"
)

// table ::= (table id? tabletype)
// tabletype ::= lim:limits et:reftype
// limits ::= n:u32 | n:u32 m:u32
// reftype ::= funcref | externref

// (table 3 funcref)
func (p *parser) parseModuleSection_table() *ast.Table {
	p.acceptToken(token.TABLE)

	tab := &ast.Table{}

	p.consumeComments()
	if p.tok == token.IDENT {
		tab.Name = p.parseIdent()
	}

	p.consumeComments()
	tab.Size = p.parseIntLit()

	p.consumeComments()
	if p.tok == token.INT {
		tab.MaxSize = p.parseIntLit()
	}

	p.consumeComments()
	p.acceptToken(token.FUNCREF)
	tab.Type = token.FUNCREF

	// Note: 不支持 externref

	return tab
}
