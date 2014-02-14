package utils

import (
	"strings"
)

func Capitalize(str string) string {
	return strings.ToUpper(str[0:1]) + str[1:]
}
