package validator

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/exp/constraints"
)

// EmailRegex is a regular expression used for sanity checking the format of email addresses
var EmailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// NotBlank returns true if string is not an empty string
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MinCharacters returns true if a string has more characters than the specified min value
func MinCharacters(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// MaxCharacters returns true if a string has less characters than the specified max value
func MaxCharacters(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// Between returns true if a value is between a given range of values
func Between[T constraints.Ordered](value, min, max T) bool {
	return value >= min && value <= max
}

// Matches returns true if a string value matches a specific regexp pattern
func Matches(value string, regex *regexp.Regexp) bool {
	return regex.MatchString(value)
}

// In returns true if a specific value is in a list of strings
func In[T comparable](value T, safelist ...T) bool {
	for i := range safelist {
		if value == safelist[i] {
			return true
		}
	}

	return false
}

// AllIn returns true if all values are in a list of strings
func AllIn[T comparable](values []T, safelist ...T) bool {
	for i := range values {
		if !In(values[i], safelist...) {
			return false
		}
	}

	return true
}

// NotIn returns true if a specific value is not in a list of strings
func NotIn[T comparable](value T, blocklist ...T) bool {
	for i := range blocklist {
		if value == blocklist[i] {
			return false
		}
	}

	return true
}

// NoDuplicates returns true if all string values in a slice are unique
func NoDuplicates[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

// IsEmail returns true if input is an email address
func IsEmail(value string) bool {
	if len(value) > 254 {
		return false
	}

	return EmailRegex.MatchString(value)
}

// IsURL returns true if input is a valid URL
func IsURL(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}
