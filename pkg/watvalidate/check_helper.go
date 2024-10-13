package watvalidate

import (
	"fmt"
	"strconv"

	"github.com/chai2010/wat-go/pkg/ast"
	"github.com/chai2010/wat-go/pkg/token"
)

func (p *watChecker) Tracef(foramt string, a ...interface{}) {
	if p.trace {
		fmt.Printf(foramt, a...)
	}
}

func (p *watChecker) findGlobalType(ident string) token.Token {
	if ident == "" {
		panic("wat2c: empty ident")
	}

	// 不支持导入的全局变量
	if idx, err := strconv.Atoi(ident); err == nil {
		if idx < 0 || idx >= len(p.m.Globals) {
			panic(fmt.Sprintf("wat2c: unknown global %q", ident))
		}
		return p.m.Globals[idx].Type
	}
	for _, g := range p.m.Globals {
		if g.Name == ident {
			return g.Type
		}
	}
	panic("unreachable")
}

func (p *watChecker) findLocalType(fn *ast.Func, ident string) token.Token {
	if ident == "" {
		panic("wat2c: empty ident")
	}

	if idx, err := strconv.Atoi(ident); err == nil {
		if idx < 0 || idx >= len(fn.Type.Params)+len(fn.Body.Locals) {
			panic(fmt.Sprintf("wat2c: unknown local %q", ident))
		}
		if idx < len(fn.Type.Params) {
			return fn.Type.Params[idx].Type
		}
		return fn.Body.Locals[idx-len(fn.Type.Params)].Type
	}
	for _, arg := range fn.Type.Params {
		if arg.Name == ident {
			return arg.Type
		}
	}
	for _, local := range fn.Body.Locals {
		if local.Name == ident {
			return local.Type
		}
	}
	panic("unreachable")
}

func (p *watChecker) findType(ident string) *ast.FuncType {
	if ident == "" {
		panic("wat2c: empty ident")
	}

	if idx, err := strconv.Atoi(ident); err == nil {
		if idx < 0 || idx >= len(p.m.Types) {
			panic(fmt.Sprintf("wat2c: unknown type %q", ident))
		}
		return p.m.Types[idx].Type
	}
	for _, x := range p.m.Types {
		if x.Name == ident {
			return x.Type
		}
	}
	panic(fmt.Sprintf("wat2c: unknown type %q", ident))
}

func (p *watChecker) findFuncType(ident string) *ast.FuncType {
	if ident == "" {
		panic("wat2c: empty ident")
	}

	idx := p.findFuncIndex(ident)
	if idx < len(p.m.Imports) {
		return p.m.Imports[idx].FuncType
	}

	return p.m.Funcs[idx-len(p.m.Imports)].Type
}

func (p *watChecker) findFuncIndex(ident string) int {
	if ident == "" {
		panic("wat2c: empty ident")
	}

	if idx, err := strconv.Atoi(ident); err == nil {
		return idx
	}

	var importCount int
	for _, x := range p.m.Imports {
		if x.ObjKind == token.FUNC {
			if x.FuncName == ident {
				return importCount
			}
			importCount++
		}
	}
	for i, fn := range p.m.Funcs {
		if fn.Name == ident {
			return importCount + i
		}
	}
	panic(fmt.Sprintf("wat2c: unknown func %q", ident))
}

func (p *watChecker) findLabelName(label string) string {
	if label == "" {
		panic("wat2c: empty label")
	}

	idx := p.findLabelIndex(label)
	if idx < len(p.scopeLabels) {
		return p.scopeLabels[len(p.scopeLabels)-idx-1]
	}
	panic(fmt.Sprintf("wat2c: unknown label %q", label))
}

func (p *watChecker) findLabelIndex(label string) int {
	if label == "" {
		panic("wat2c: empty label")
	}

	if idx, err := strconv.Atoi(label); err == nil {
		return idx
	}
	for i := 0; i < len(p.scopeLabels); i++ {
		if s := p.scopeLabels[len(p.scopeLabels)-i-1]; s == label {
			return i
		}
	}
	panic(fmt.Sprintf("wat2c: unknown label %q", label))
}

func (p *watChecker) enterLabelScope(stkBase int, label string) {
	p.scopeLabels = append(p.scopeLabels, label)
	p.scopeStackBases = append(p.scopeStackBases, stkBase)
}
func (p *watChecker) leaveLabelScope() {
	p.scopeLabels = p.scopeLabels[:len(p.scopeLabels)-1]
}
