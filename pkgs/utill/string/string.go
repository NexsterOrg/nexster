package string

import (
	"fmt"
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
