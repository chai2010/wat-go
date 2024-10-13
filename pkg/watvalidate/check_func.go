// 版权 @2024 wat-go 作者。保留所有权利。

package watvalidate

import (
	"fmt"

	"github.com/chai2010/wat-go/pkg/ast"
	"github.com/chai2010/wat-go/pkg/token"
)

func (p *watChecker) checkFunc_body(fn *ast.Func) error {
	p.Tracef("buildFunc_body: %s\n", fn.Name)

	p.fn = fn
	p.fnCheckedInsList = nil

	var stk valueTypeStack

	p.scopeLabels = nil
	p.scopeStackBases = nil

	assert(stk.Len() == 0)
	for i, ins := range fn.Body.Insts {
		if err := p.checkFunc_ins(fn, &stk, ins, 1); err != nil {
			return err
		}
		// 手动补充最后一个 return
		if i == len(fn.Body.Insts)-1 && ins.Token() != token.INS_RETURN {
			insReturn := ast.Ins_Return{OpToken: ast.OpToken(token.INS_RETURN)}
			if err := p.checkFunc_ins(fn, &stk, insReturn, 1); err != nil {
				return err
			}
		}
	}
	assert(stk.Len() == 0)
	return nil
}

func (p *watChecker) checkFunc_ins(fn *ast.Func, stk *valueTypeStack, i ast.Instruction, level int) error {
	p.Tracef("checkFunc_ins: %s begin: %v\n", i.Token(), stk.String())
	defer func() { p.Tracef("checkFunc_ins: %s end: %v\n", i.Token(), stk.String()) }()

	p.fnCheckedInsList = append(p.fnCheckedInsList, i)

	switch tok := i.Token(); tok {
	case token.INS_UNREACHABLE:
	case token.INS_NOP:

	case token.INS_BLOCK:
		i := i.(ast.Ins_Block)

		stkBase := stk.Len()
		defer func() { assert(stk.Len() == stkBase+len(i.Results)) }()

		p.enterLabelScope(stkBase, i.Label)
		defer p.leaveLabelScope()

		for _, ins := range i.List {
			if err := p.checkFunc_ins(fn, stk, ins, level+1); err != nil {
				return err
			}
		}

	case token.INS_LOOP:
		i := i.(ast.Ins_Loop)

		stkBase := stk.Len()
		defer func() { assert(stk.Len() == stkBase+len(i.Results)) }()

		p.enterLabelScope(stkBase, i.Label)
		defer p.leaveLabelScope()

		for _, ins := range i.List {
			if err := p.checkFunc_ins(fn, stk, ins, level+1); err != nil {
				return err
			}
		}

	case token.INS_IF:
		i := i.(ast.Ins_If)

		stk.Pop(token.I32)

		stkBase := stk.Len()
		defer func() { assert(stk.Len() == stkBase+len(i.Results)) }()

		p.enterLabelScope(stkBase, i.Label)
		defer p.leaveLabelScope()

		for _, ins := range i.Body {
			if err := p.checkFunc_ins(fn, stk, ins, level+1); err != nil {
				return err
			}
		}

		if len(i.Else) > 0 {
			// 这是静态分析, 需要消除 if 分支对栈分配的影响
			for _, retType := range i.Results {
				stk.Pop(retType)
			}

			// 重新开始
			assert(stk.Len() == stkBase)

			for _, ins := range i.Else {
				if err := p.checkFunc_ins(fn, stk, ins, level+1); err != nil {
					return err
				}
			}
		}

	case token.INS_ELSE:
		unreachable()
	case token.INS_END:
		unreachable()

	case token.INS_BR:
		i := i.(ast.Ins_Br)

		labelIdx := p.findLabelIndex(i.X)
		p.findLabelName(i.X)

		stkBase := p.scopeStackBases[len(p.scopeLabels)-labelIdx-1]
		assert(stk.Len() == stkBase)

	case token.INS_BR_IF:
		i := i.(ast.Ins_BrIf)
		labelIdx := p.findLabelIndex(i.X)
		labelName := p.findLabelName(i.X)

		stkBase := p.scopeStackBases[len(p.scopeLabels)-labelIdx-1]

		sp0 := stk.Pop(token.I32)
		assert(stk.Len() == stkBase)

		_ = labelName
		_ = sp0

	case token.INS_BR_TABLE:
		i := i.(ast.Ins_BrTable)
		assert(len(i.XList) > 1)

		sp0 := stk.Pop(token.I32)
		_ = sp0

	case token.INS_RETURN:
		for _, xType := range fn.Type.Results {
			spi := stk.Pop(xType)
			_ = spi
		}
		assert(stk.Len() == 0)

	case token.INS_CALL:
		i := i.(ast.Ins_Call)

		fnCallType := p.findFuncType(i.X)

		// 返回值
		argList := make([]int, len(fnCallType.Params))
		for k, x := range fnCallType.Params {
			argList[k] = stk.Pop(x.Type)
		}

		// 需要定义临时变量保存返回值
		switch len(fnCallType.Results) {
		case 0:
		case 1:
			ret0 := stk.Push(fnCallType.Results[0])
			_ = ret0
		}

		// 复制到当前stk
		if len(fnCallType.Results) > 1 {
			for _, retType := range fnCallType.Results {
				reti := stk.Push(retType)
				_ = reti
			}
		}

	case token.INS_CALL_INDIRECT:
		i := i.(ast.Ins_CallIndirect)

		sp0 := stk.Pop(token.I32)
		fnCallType := p.findType(i.TypeIdx)

		_ = sp0

		// 返回值
		argList := make([]int, len(fnCallType.Params))
		for k, x := range fnCallType.Params {
			argList[k] = stk.Pop(x.Type)
		}

		{
			for _, x := range fnCallType.Params {

				argi := stk.Pop(x.Type)
				_ = argi
			}

			// 保存返回值
			if len(fnCallType.Results) > 1 {
				for _, retType := range fnCallType.Results {
					reti := stk.Push(retType)
					_ = reti
				}
			}
		}

	case token.INS_DROP:
		sp0 := stk.DropAny()
		_ = sp0
	case token.INS_SELECT:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		sp2 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = sp2
		_ = ret0

	case token.INS_LOCAL_GET:
		i := i.(ast.Ins_LocalGet)
		xType := p.findLocalType(fn, i.X)
		ret0 := stk.Push(xType)
		_ = ret0

	case token.INS_LOCAL_SET:
		i := i.(ast.Ins_LocalSet)
		xType := p.findLocalType(fn, i.X)
		sp0 := stk.Pop(xType)
		_ = sp0

	case token.INS_LOCAL_TEE:
		i := i.(ast.Ins_LocalTee)
		xType := p.findLocalType(fn, i.X)
		sp0 := stk.Top(xType)
		_ = sp0
	case token.INS_GLOBAL_GET:
		i := i.(ast.Ins_GlobalGet)
		xType := p.findGlobalType(i.X)
		ret0 := stk.Push(xType)
		_ = ret0
	case token.INS_GLOBAL_SET:
		i := i.(ast.Ins_GlobalSet)
		xType := p.findGlobalType(i.X)
		sp0 := stk.Pop(xType)
		_ = sp0
	case token.INS_TABLE_GET:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.FUNCREF) // funcref
		_ = sp0
		_ = ret0
	case token.INS_TABLE_SET:
		sp0 := stk.Pop(token.FUNCREF) // funcref
		sp1 := stk.Pop(token.I32)
		_ = sp0
		_ = sp1
	case token.INS_I32_LOAD:
		i := i.(ast.Ins_I32Load)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = i
		_ = sp0
		_ = ret0

	case token.INS_I64_LOAD:
		i := i.(ast.Ins_I64Load)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_F32_LOAD:
		i := i.(ast.Ins_F32Load)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.F32)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_F64_LOAD:
		i := i.(ast.Ins_I32Load)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I32_LOAD8_S:
		i := i.(ast.Ins_I32Load8S)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I32_LOAD8_U:
		i := i.(ast.Ins_I32Load8U)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I32_LOAD16_S:
		i := i.(ast.Ins_I32Load16S)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I32_LOAD16_U:
		i := i.(ast.Ins_I32Load16U)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I64_LOAD8_S:
		i := i.(ast.Ins_I64Load8S)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I64_LOAD8_U:
		i := i.(ast.Ins_I64Load8U)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I64_LOAD16_S:
		i := i.(ast.Ins_I64Load16S)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I64_LOAD16_U:
		i := i.(ast.Ins_I64Load16U)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I64_LOAD32_S:
		i := i.(ast.Ins_I64Load32S)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I64_LOAD32_U:
		i := i.(ast.Ins_I64Load32U)
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = i
		_ = sp0
		_ = ret0
	case token.INS_I32_STORE:
		i := i.(ast.Ins_I32Store)
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		_ = i
		_ = sp0
		_ = sp1
	case token.INS_I64_STORE:
		i := i.(ast.Ins_I64Store)
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I32)
		_ = i
		_ = sp0
		_ = sp1
	case token.INS_F32_STORE:
		i := i.(ast.Ins_F32Store)
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.I32)
		_ = i
		_ = sp0
		_ = sp1
	case token.INS_F64_STORE:
		i := i.(ast.Ins_F64Store)
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.I32)
		_ = i
		_ = sp0
		_ = sp1
	case token.INS_I32_STORE8:
		i := i.(ast.Ins_I32Store8)
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		_ = i
		_ = sp0
		_ = sp1
	case token.INS_I32_STORE16:
		i := i.(ast.Ins_I32Store16)
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		_ = i
		_ = sp0
		_ = sp1
	case token.INS_I64_STORE8:
		i := i.(ast.Ins_I64Store8)
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I32)
		_ = i
		_ = sp0
		_ = sp1
	case token.INS_I64_STORE16:
		i := i.(ast.Ins_I64Store16)
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I32)
		_ = i
		_ = sp0
		_ = sp1
	case token.INS_I64_STORE32:
		i := i.(ast.Ins_I64Store32)
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I32)
		_ = i
		_ = sp0
		_ = sp1
	case token.INS_MEMORY_SIZE:
		sp0 := stk.Push(token.I32)
		_ = sp0
	case token.INS_MEMORY_GROW:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I32_CONST:
		i := i.(ast.Ins_I32Const)
		sp0 := stk.Push(token.I32)
		_ = i
		_ = sp0
	case token.INS_I64_CONST:
		i := i.(ast.Ins_I64Const)
		sp0 := stk.Push(token.I64)
		_ = i
		_ = sp0
	case token.INS_F32_CONST:
		i := i.(ast.Ins_F32Const)
		sp0 := stk.Push(token.F32)
		_ = i
		_ = sp0
	case token.INS_F64_CONST:
		i := i.(ast.Ins_F64Const)
		sp0 := stk.Push(token.F64)
		_ = i
		_ = sp0
	case token.INS_I32_EQZ:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I32_EQ:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_NE:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_LT_S:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_LT_U:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_GT_S:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_GT_U:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_LE_S:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_LE_U:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_GE_S:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_GE_U:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_EQZ:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I64_EQ:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_NE:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_LT_S:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_LT_U:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_GT_S:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_GT_U:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_LE_S:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_LE_U:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_GE_S:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_GE_U:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_EQ:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_NE:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_LT:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_GT:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_LE:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_GE:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_EQ:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_NE:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_LT:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_GT:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_LE:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_GE:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_CLZ:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I32_CTZ:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I32_POPCNT:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I32_ADD:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_SUB:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_MUL:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_DIV_S:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_DIV_U:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_REM_S:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_REM_U:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_AND:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_OR:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_XOR:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_SHL:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_SHR_S:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_SHR_U:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_ROTL:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_ROTR:
		sp0 := stk.Pop(token.I32)
		sp1 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_CLZ:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I64_CTZ:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I64_POPCNT:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I64_ADD:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_SUB:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_MUL:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_DIV_S:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_DIV_U:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_REM_S:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_REM_U:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_AND:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_OR:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_XOR:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_SHL:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_SHR_S:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_SHR_U:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_ROTL:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I64_ROTR:
		sp0 := stk.Pop(token.I64)
		sp1 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_ABS:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_NEG:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_CEIL:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_FLOOR:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_TRUNC:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_NEAREST:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_SQRT:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_ADD:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_SUB:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_MUL:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_DIV:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_MIN:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_MAX:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F32_COPYSIGN:
		sp0 := stk.Pop(token.F32)
		sp1 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_ABS:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_NEG:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_CEIL:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_FLOOR:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_TRUNC:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_NEAREST:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_SQRT:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_ADD:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_SUB:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_MUL:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_DIV:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_MIN:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_MAX:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_F64_COPYSIGN:
		sp0 := stk.Pop(token.F64)
		sp1 := stk.Pop(token.F64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = sp1
		_ = ret0
	case token.INS_I32_WRAP_I64:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I32_TRUNC_F32_S:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I32_TRUNC_F32_U:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I32_TRUNC_F64_S:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I32_TRUNC_F64_U:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I32)
		_ = sp0
		_ = ret0
	case token.INS_I64_EXTEND_I32_S:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = ret0
	case token.INS_I64_EXTEND_I32_U:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = ret0
	case token.INS_I64_TRUNC_F32_S:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = ret0
	case token.INS_I64_TRUNC_F32_U:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = ret0
	case token.INS_I64_TRUNC_F64_S:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = ret0
	case token.INS_I64_TRUNC_F64_U:
		sp0 := stk.Pop(token.F64)
		ret0 := stk.Push(token.I64)
		_ = sp0
		_ = ret0
	case token.INS_F32_CONVERT_I32_S:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_CONVERT_I32_U:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_CONVERT_I64_S:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_CONVERT_I64_U:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F32_DEMOTE_F64:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F64_CONVERT_I32_S:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_CONVERT_I32_U:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_CONVERT_I64_S:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_CONVERT_I64_U:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F64_PROMOTE_F32:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_I32_REINTERPRET_F32:
		sp0 := stk.Pop(token.F32)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_I64_REINTERPRET_F64:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	case token.INS_F32_REINTERPRET_I32:
		sp0 := stk.Pop(token.I32)
		ret0 := stk.Push(token.F32)
		_ = sp0
		_ = ret0
	case token.INS_F64_REINTERPRET_I64:
		sp0 := stk.Pop(token.I64)
		ret0 := stk.Push(token.F64)
		_ = sp0
		_ = ret0
	default:
		panic(fmt.Sprintf("unreachable: %T", i))
	}
	return nil
}
