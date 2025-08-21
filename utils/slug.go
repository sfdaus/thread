package utils

import (
	"strings"
	"unicode"
)

func Slugify(title string) string {
	title = strings.ToLower(title)
	var sb strings.Builder
	for _, r := range title {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else if unicode.IsSpace(r) {
			sb.WriteRune('-')
		}
	}
	return sb.String()
}
