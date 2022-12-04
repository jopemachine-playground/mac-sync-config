package utils

import (
	"os"
	"strings"

	"github.com/eiannone/keyboard"
)

func ScanChar() string {
	// TODO: Refactoring below code.
	allowedKeys := []string{"y", "n", "p", "q", "d", "e"}
	for {
		char, key, err := keyboard.GetSingleKey()
		FatalIfError(err)

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

func WaitResponse() {
	ScanChar()
}

type QuestionResult string

const (
	QUESTION_RESULT_PATCH     = QuestionResult("PATCH")
	QUESTION_RESULT_ADD       = QuestionResult("ADD")
	QUESTION_RESULT_IGNORE    = QuestionResult("IGNORE")
	QUESTION_RESULT_SHOW_DIFF = QuestionResult("SHOW_DIFF")
	QUESTION_RESULT_EDIT      = QuestionResult("EDIT")
)

var PUSH_CONFIG_ALLOWED_KEYS = []string{"y", "p", "d", "e", "n"}
var PULL_CONFIG_ALLOWED_KEYS = []string{"y", "d", "n", "e"}

func MakeYesNoQuestion() bool {
	response := ScanChar()
	return strings.ToLower(response) == "y"
}

func MakeQuestion(allowedKeys []string) QuestionResult {
	response := ScanChar()

	if !StringContains(allowedKeys, response) {
		return QUESTION_RESULT_IGNORE
	}

	if strings.ToLower(response) == "y" {
		return QUESTION_RESULT_ADD
	} else if strings.ToLower(response) == "p" {
		return QUESTION_RESULT_PATCH
	} else if strings.ToLower(response) == "d" {
		return QUESTION_RESULT_SHOW_DIFF
	} else if strings.ToLower(response) == "e" {
		return QUESTION_RESULT_EDIT
	}

	return QUESTION_RESULT_IGNORE
}
