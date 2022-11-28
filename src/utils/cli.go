package utils

import (
	"os"
	"strings"

	"github.com/eiannone/keyboard"
)

func ScanChar() string {
	allowedKeys := []string{"y", "n", "p", "q"}
	for {
		char, key, err := keyboard.GetSingleKey()
		PanicIfErr(err)

		response := strings.ToLower(string(char))
		if key == keyboard.KeyEsc || key == keyboard.KeyCtrlC || response == "q" {
			os.Exit(0)
		}

		if key == keyboard.KeyEnter {
			response = "y"
		}

		if StringContains(allowedKeys, response) {
			return response
		}
	}
}

func EnterYesNoQuestion() bool {
	response := ScanChar()
	return strings.ToLower(response) == "y"
}

func WaitResponse() {
	ScanChar()
}

type ConfigAddQuestionResult string

const (
	PATCH  = ConfigAddQuestionResult("PATCH")
	ADD    = ConfigAddQuestionResult("ADD")
	IGNORE = ConfigAddQuestionResult("IGNORE")
)

func ConfigAddQuestion() ConfigAddQuestionResult {
	response := ScanChar()

	if strings.ToLower(response) == "y" {
		return ADD
	} else if strings.ToLower(response) == "p" {
		return PATCH
	}

	return IGNORE
}
