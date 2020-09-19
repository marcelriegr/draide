package ui

import (
	"fmt"
	"os"

	au "github.com/logrusorgru/aurora/v3"
	"github.com/spf13/viper"
)

// Logf tbd
func Logf(format string, args ...interface{}) {
	if viper.GetBool("verbose") {
		fmt.Printf(format, args...)
	}
}

// Log tbd
func Log(format string, args ...interface{}) {
	Logf(format+"\n", args...)
}

// Info tbd
func Info(format string, args ...interface{}) {
	fmt.Println(au.Sprintf(au.BrightBlue(format), args...))
}

// Success tbd
func Success(format string, args ...interface{}) {
	fmt.Println(au.Sprintf(au.BrightGreen(format), args...))
}

// Warning tbd
func Warning(format string, args ...interface{}) {
	fmt.Println(au.Sprintf(au.Yellow(format), args...))
}

// Error tbd
func Error(format string, args ...interface{}) {
	fmt.Println(au.Sprintf(au.Red(format), args...))
}

// ErrorAndExit tbd
func ErrorAndExit(code int, format string, args ...interface{}) int {
	Error(format, args...)
	os.Exit(code)
	return code
}
