// 版权 @2024 wat-go 作者。保留所有权利。

package parser_test

import (
	"os"
	"testing"

	"github.com/chai2010/wat-go/pkg/parser"
)

func TestParseModule(t *testing.T) {
	const filename = "../../testdata/hello.wat"
	src := tReadFile(t, filename)

	_, err := parser.ParseModule(filename, src)
	if err != nil {
		t.Fatal(err)
	}
}

func tReadFile(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
