// 版权 @2024 wat-go 作者。保留所有权利。

package watvalidate

import (
	"github.com/chai2010/wat-go/pkg/ast"
	"github.com/chai2010/wat-go/pkg/parser"
)

const DebugMode = false

func WatValidate(path string, src []byte) (fn *ast.Func, ins []ast.Instruction, err error) {
	m, err := parser.ParseModule(path, src)
	if err != nil {
		return nil, nil, err
	}

	checker := newWatChecker(m)
	defer func() {
		if r := recover(); r != nil {
			if errx, ok := r.(*watError); ok {
				fn = checker.fn
				ins = checker.fnCheckedInsList
				err = errx
			} else {
				panic(r)
			}
		}
	}()

	if err = checker.checkFuncs(); err != nil {
		return checker.fn, checker.fnCheckedInsList, err
	}
	return nil, nil, nil
}

type watChecker struct {
	m *ast.Module

	fn               *ast.Func         // 当前检查的函数
	fnCheckedInsList []ast.Instruction // 当前检查的指令索引

	scopeLabels     []string // 嵌套的label查询, if/block/loop
	scopeStackBases []int    // if/block/loop, 开始的栈位置

	trace bool // 调试开关
}

func newWatChecker(mWat *ast.Module) *watChecker {
	return &watChecker{m: mWat, trace: DebugMode}
}

func (p *watChecker) checkFuncs() error {
	if len(p.m.Funcs) == 0 {
		return nil
	}

	for _, f := range p.m.Funcs {
		if err := p.checkFunc_body(f); err != nil {
			return err
		}
	}

	return nil
}
