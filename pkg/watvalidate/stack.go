// 版权 @2024 wat-go 作者。保留所有权利。

package watvalidate

import (
	"fmt"
	"strings"

	"github.com/chai2010/wat-go/pkg/token"
)

type watError struct {
	msg string
}

func (e *watError) Error() string {
	return e.msg
}

func assert(condition bool, args ...interface{}) {
	if !condition {
		if msg := fmt.Sprint(args...); msg != "" {
			panic(&watError{fmt.Sprintf("assert failed, %s", msg)})
		} else {
			panic(&watError{fmt.Sprint("assert failed")})
		}
	}
}

// 函数内栈类型和深度计算
type valueTypeStack struct {
	stack           []token.Token // i32/i64/f32/f64/funcref
	maxStackPointer int
}

func (s *valueTypeStack) Len() int {
	return len(s.stack)
}

func (s *valueTypeStack) MaxDepth() int {
	return s.maxStackPointer
}

func (s *valueTypeStack) Top(expect token.Token) int {
	if len(s.stack) == 0 {
		panic(&watError{"unexpected stack empty"})
	}
	idx := len(s.stack) - 1
	if got := s.stack[idx]; got != expect {
		panic(&watError{"unexpected value type: got " + got.String() + ", expect " + expect.String()})
	}
	return idx
}

func (s *valueTypeStack) Push(v token.Token) int {
	switch v {
	case token.I32, token.I64, token.F32, token.F64, token.FUNCREF:
	default:
		panic(&watError{"unexpected value type"})
	}
	s.stack = append(s.stack, v)
	if sp := len(s.stack); sp > s.maxStackPointer {
		s.maxStackPointer = sp
	}
	return len(s.stack) - 1
}

func (s *valueTypeStack) Pop(expect token.Token) int {
	switch expect {
	case token.I32, token.I64, token.F32, token.F64, token.FUNCREF:
	default:
		panic(&watError{"unexpected value type"})
	}

	if len(s.stack) == 0 {
		panic(&watError{"unexpected stack empty"})
	}
	idx := len(s.stack) - 1
	if got := s.stack[idx]; got != expect {
		panic(&watError{"unexpected value type: got " + got.String() + ", expect " + expect.String()})
	}
	s.stack = s.stack[:len(s.stack)-1]
	return idx
}

func (s *valueTypeStack) DropAny() int {
	if len(s.stack) == 0 {
		panic(&watError{"unexpected stack empty"})
	}
	idx := len(s.stack) - 1
	s.stack = s.stack[:len(s.stack)-1]
	return idx
}

func (s *valueTypeStack) String() string {
	var sb strings.Builder
	sb.WriteString("stack: [")
	for i, v := range s.stack {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(v.String())
	}
	sb.WriteString("]")
	return sb.String()
}
