// 版权 @2024 wat-go 作者。保留所有权利。

package watstrip

import (
	"errors"
	"os"
)

func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		}
		return nil, errors.New("invalid source")
	}

	return os.ReadFile(filename)
}
