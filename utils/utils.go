package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func Capitalize(str string) string {
	return strings.ToUpper(str[0:1]) + str[1:]
}

func HumanSize(size int64) string {
	i := 0
	units := []string{"B", "K", "M", "G", "T", "P", "E", "Z", "Y"}
	for size >= 1024 {
		size = size / 1024
		i++
	}
	return fmt.Sprintf("%d%s", size, units[i])
}

// Parses a human-readable string representing an amount of RAM
// in bytes, kibibytes, mebibytes or gibibytes, and returns the
// number of bytes, or -1 if the string is unparseable.
// Units are case-insensitive, and the 'b' suffix is optional.
func RAMInBytes(size string) (bytes int64, err error) {
	re, error := regexp.Compile("^(\\d+)([kKmMgG])?[bB]?$")
	if error != nil {
		return -1, error
	}

	matches := re.FindStringSubmatch(size)

	if len(matches) != 3 {
		return -1, fmt.Errorf("Invalid size: '%s'", size)
	}

	memLimit, error := strconv.ParseInt(matches[1], 10, 0)
	if error != nil {
		return -1, error
	}

	unit := strings.ToLower(matches[2])

	if unit == "k" {
		memLimit *= 1024
	} else if unit == "m" {
		memLimit *= 1024 * 1024
	} else if unit == "g" {
		memLimit *= 1024 * 1024 * 1024
	}

	return memLimit, nil
}
