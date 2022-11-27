package utils

import (
	"os"
	"strings"

	"github.com/eiannone/keyboard"
)

func ScanValue() string {
	char, key, err := keyboard.GetSingleKey()
	PanicIfErr(err)
	response := string(char)
	if key == keyboard.KeyEsc || strings.ToLower(response) == "q" || key == keyboard.KeyCtrlC {
		os.Exit(0)
	}

	return response
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
