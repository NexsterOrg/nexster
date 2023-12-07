package string

import "unicode"

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
