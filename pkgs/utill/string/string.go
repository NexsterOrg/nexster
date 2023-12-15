package string

import (
	"fmt"
	"strconv"
	"unicode"
)

func FirstCharToLower(input string) string {
	if len(input) == 0 {
		return input
	}

	inputInrunes := []rune(input)
	if unicode.IsUpper(inputInrunes[0]) {
		inputInrunes[0] = unicode.ToLower(inputInrunes[0])
	}

	return string(inputInrunes)
}

func FirstCharToUpper(input string) string {
	if len(input) == 0 {
		return input
	}

	inputInrunes := []rune(input)
	if unicode.IsLower(inputInrunes[0]) {
		inputInrunes[0] = unicode.ToUpper(inputInrunes[0])
	}

	return string(inputInrunes)
}

func InterfaceToString(value interface{}) (string, error) {
	if value == nil {
		return "", fmt.Errorf("input is nil")
	}
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return fmt.Sprintf("%d", v), nil
	// Add cases for other types you want to handle
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func StrToInt64(s string) (int64, error) {
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}
