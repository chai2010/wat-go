package watvalidate

import (
	"strconv"
	"strings"

	"github.com/chai2010/wat-go/pkg/token"
)

func unreachable() {
	panic("unreachable")
}

func toCType(typ token.Token) string {
	switch typ {
	case token.I32:
		return "int32_t"
	case token.I64:
		return "int64_t"
	case token.F32:
		return "float"
	case token.F64:
		return "double"
	default:
		return "void"
	}
}

// 生成C语言的标识符
func toCGlobalName(name string) string {
	return "g_" + toCName(name)
}
func toCFuncName(name string) string {
	return "f_" + toCName(name)
}
func toCFuncArgName(name string) string {
	return "a_" + toCName(name)
}
func toCFuncLocalName(name string) string {
	return "v_" + toCName(name)
}
func toCFuncLabelName(name string) string {
	return "L_" + toCName(name)
}

func toCName(name string) string {
	if name == "" {
		return name
	}
	if c := name[0]; c >= '0' && c <= '9' {
		return name
	}

	var sb strings.Builder
	for _, c := range ([]rune)(name) {
		switch {
		case c == '_':
			sb.WriteRune(c)
		case 'a' <= c && c <= 'z':
			sb.WriteRune(c)
		case 'A' <= c && c <= 'Z':
			sb.WriteRune(c)
		case c == '$':
			// $ 是保留字符，转义为 _$$_
			sb.WriteRune('_')
			sb.WriteRune('$')
			sb.WriteRune('$')
			sb.WriteRune('_')
		default:
			sb.WriteRune('_')
			sb.WriteRune('$')
			sb.WriteString(strconv.FormatInt(int64(c), 16))
			sb.WriteRune('_')
		}
	}

	return sb.String()
}
