package parser

import (
	"os"
	"regexp"
)

// Env replaces all variable starting with a $ (dollar sign) or # (number sign) character inside a string with the corresponding environment variable
func Env(str string) string {
	pattern := regexp.MustCompile(`[\$#]\w+`)
	return pattern.ReplaceAllStringFunc(str, func(s string) string {
		return os.Getenv(s[1:])
	})
}
