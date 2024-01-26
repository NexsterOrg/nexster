package string

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
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

func InterfaceToStringArray(value interface{}) ([]string, error) {
	switch v := value.(type) {
	case []string:
		return v, nil
	case []interface{}:
		result := []string{}
		for i, elem := range v {
			strElem, ok := elem.(string)
			if !ok {
				log.Println(fmt.Errorf("converting interface to string arrary: element at index %d is not a string", i))
				continue
			}
			result = append(result, strElem)
		}
		return result, nil
	default:
		return []string{}, fmt.Errorf("unsupported type: %T", value)
	}
}

func StrToInt64(s string) (int64, error) {
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func ConvertBirthdayToCorrectFormat(bd string) string {
	arr := strings.Split(bd, "T")
	if len(arr) >= 1 {
		return arr[0]
	}
	return ""
}

func IsInArray(arr []string, id string) bool {
	for _, element := range arr {
		if element == id {
			return true
		}
	}
	return false
}

func MkCompletePath(rootDir, relPath string) string {
	return fmt.Sprintf("%s/%s", rootDir, relPath)
}

// Check valid mobile phone number or not.
func ConvertToValidMobileNo(phoneNo string) (string, error) {
	if len(phoneNo) != 10 {
		return "", fmt.Errorf("len should be 10")
	}

	if phoneNo[:2] != "07" {
		return "", fmt.Errorf("prefix should be 07")
	}

	// Check if the remaining 8 characters are numeric

	if !regexp.MustCompile(`^[0-9]+$`).MatchString(phoneNo[2:]) {
		return "", fmt.Errorf("all characters should be numberic")
	}
	return "94" + phoneNo[1:], nil
}
