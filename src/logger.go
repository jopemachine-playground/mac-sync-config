package src

import (
	"bytes"
	"fmt"
	"io/ioutil"

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

func (logger loggerType) Success(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.GreenString("✔ "), msg))
}

func (logger loggerType) Error(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.RedString("✖ "), msg))
}

func (logger loggerType) Info(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.BlueString("ℹ "), msg))
}

func (logger loggerType) Warning(msg string) {
	Logger.Log(fmt.Sprintf("%s %s", color.BlueString("⚠️ "), msg))
}

func (logger loggerType) FileLogAppend(msg string) {
	logFile.WriteString(msg)
}

func (logger loggerType) WriteFileLog() {
	err := ioutil.WriteFile("/logs", logFile.Bytes(), 0644)

	if err != nil {
		panic(err)
	}
}
