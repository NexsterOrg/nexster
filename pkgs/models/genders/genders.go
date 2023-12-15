package genders

import "fmt"

func MkGenderId(key string) string {
	return fmt.Sprintf("genders/%s", key)
}
