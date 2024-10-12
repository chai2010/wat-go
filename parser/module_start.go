// 版权 @2024 wat-go 作者。保留所有权利。

package parser

import "github.com/chai2010/wat-go/token"

// start ::= (start funcidx)

func (p *parser) parseModuleSection_start() string {
	p.acceptToken(token.START)

	p.consumeComments()
	return p.parseIdent()
}
