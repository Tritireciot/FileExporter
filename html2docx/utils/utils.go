package utils

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func Contains[T comparable](list []T, elem T) bool {
	for _, e := range list {
		if elem == e {
			return true
		}
	}
	return false
}

func Capitalize(s string) string {
	if s == "" {
		return ""
	}

	r, size := utf8.DecodeRuneInString(s)

	return string(unicode.ToUpper(r)) + s[size:]
}

func UpdateText(text string, transform string) string {
	switch transform {
	case "uppercase":
		return strings.ToUpper(text)
	case "lowercase":
		return strings.ToLower(text)
	case "capitalize":
		words := strings.Split(text, " ")
		for i, word := range words {
			words[i] = Capitalize(word)
		}
		return strings.Join(words, " ")
	}
	return text
}

func CopyMap[K comparable, V any](originMap map[K]V) map[K]V {
	copyMap := make(map[K]V)
	for key, value := range originMap {
		copyMap[key] = value
	}
	return copyMap
}

func Pointer[T any](variable T) *T {
	return &variable
}
