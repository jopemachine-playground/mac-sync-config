package src

import (
	"bytes"
	"fmt"

	"github.com/fatih/color"
)

type loggerType struct{}

var (
	Logger  loggerType
	logFile bytes.Buffer
)

func (logger loggerType) Log(msg string) {
	fmt.Println(msg)
}

func (logger loggerType) NewLine() {
	fmt.Println()
}

func (logger loggerType) Success(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.GreenString("✔"), msg))
}

func (logger loggerType) Error(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.RedString("✖"), color.RedString(msg)))
}

func (logger loggerType) Info(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.BlueString("ℹ"), msg))
}

func (logger loggerType) Warning(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.YellowString("⚠️"), color.RedString(msg)))
}

func (logger loggerType) Question(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.GreenString("?"), color.GreenString(msg)))
}
