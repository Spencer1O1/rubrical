package extract

import (
	"fmt"
	"unicode/utf8"
)

func Text(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty file")
	}
	if utf8.Valid(data) {
		return string(data), nil
	}
	return "", fmt.Errorf("could not decode as UTF-8 text")
}
