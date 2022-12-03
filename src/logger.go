package src

import (
	"fmt"

	"github.com/fatih/color"
)

type loggerType struct{}

var (
	Logger loggerType
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
	Logger.Log(fmt.Sprintf("%s %s", color.YellowString("⚠️"), msg))
}

func (logger loggerType) Question(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.GreenString("?"), msg))
}

// Warning: It might not be cross-platform
func (logger loggerType) ClearConsole() {
	fmt.Print("\033[H\033[2J")
}
