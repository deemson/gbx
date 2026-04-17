package clilog

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

func log(level string, message string) {
	fmt.Println(strings.Join([]string{
		color.HiBlackString("%s", time.Now().Format(time.TimeOnly)),
		level,
		message,
	}, " "))
}

func Info(message string) {
	log(color.GreenString("I"), message)
}

func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}

func Error(message string) {
	log(color.RedString("E"), message)
}

func Errorf(format string, args ...any) {
	Error(fmt.Sprintf(format, args...))
}
