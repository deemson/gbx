// Package clilog is the stdout/stderr logger for the cobra/config layer only.
// The running TUI owns stdout and logs via zerolog to a file, so do not use this
// from internal/tui.
package clilog

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

func log(w io.Writer, level string, message string) {
	fmt.Fprintln(w, strings.Join([]string{
		color.HiBlackString(time.Now().Format(time.TimeOnly)),
		level,
		message,
	}, " "))
}

func Info(message string) {
	log(os.Stdout, color.GreenString("I"), message)
}

func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}

func Error(message string) {
	log(os.Stderr, color.RedString("E"), message)
}

func Errorf(format string, args ...any) {
	Error(fmt.Sprintf(format, args...))
}
