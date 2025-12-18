package log

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Info prints an informational message to Stderr
func Info(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "[INFO] "+format+"\n", a...)
}

// Success prints a success message to Stderr
func Success(format string, a ...interface{}) {
	color.New(color.FgGreen).Fprintf(os.Stderr, "[SUCCESS] "+format+"\n", a...)
}

// Warn prints a warning log
func Warn(format string, a ...interface{}) {
	color.New(color.FgYellow).Fprintf(os.Stderr, "[WARN] "+format+"\n", a...)
}

// Error prints an error message to Stderr
func Error(format string, a ...interface{}) {
	color.New(color.FgRed).Fprintf(os.Stderr, "[ERROR] "+format+"\n", a...)
}

// Debug prints a debug message to Stderr if verbose mode is enabled (caller needs to check verbose)
func Debug(format string, a ...interface{}) {
	color.New(color.FgMagenta).Fprintf(os.Stderr, "[DEBUG] "+format+"\n", a...)
}

// Fatal prints an error message to Stderr and exits
func Fatal(format string, a ...interface{}) {
	Error(format, a...)
	os.Exit(1)
}

// PrintJSON outputs raw JSON to Stdout
// This is the ONLY function that should write to Stdout when using --format=json
func PrintJSON(jsonData []byte) {
	fmt.Fprintf(os.Stdout, "%s\n", string(jsonData))
}
