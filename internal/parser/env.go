package parser

import (
	"os"
	"regexp"

	"github.com/marcelriegr/draide/pkg/ui"
)

// Env replaces all variable starting with a $ (dollar sign) or # (number sign) character inside a string with the corresponding environment variable
func Env(str string) string {
	pattern := regexp.MustCompile(`[\$#]\w+`)
	return pattern.ReplaceAllStringFunc(str, func(envVar string) string {
		val := os.Getenv(envVar[1:])

		if val == "" {
			ui.ErrorAndExit(1, "Cannot resolve environment variable %s", envVar)
		}

		return val
	})
}
