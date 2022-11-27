package utils

import (
	"strings"

	"github.com/eiannone/keyboard"
)

func ScanValue() string {
	char, _, err := keyboard.GetSingleKey()
	PanicIfErr(err)
	return string(char)
}

func EnterYesNoQuestion() bool {
	response := ScanValue()
	return strings.ToLower(response) == "y"
}

func WaitResponse() {
	ScanValue()
}

type ConfigAddQuestionResult string

const (
	PATCH = ConfigAddQuestionResult("PATCH")
	ADD = ConfigAddQuestionResult("ADD")
	IGNORE = ConfigAddQuestionResult("IGNORE")
)

func ConfigAddQuestion() ConfigAddQuestionResult {
	response := ScanValue()

	if strings.ToLower(response) == "y" {
		return ADD
	} else if strings.ToLower(response) == "p" {
		return PATCH
	}

	return IGNORE
}
