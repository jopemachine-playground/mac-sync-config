package utils

import (
	"fmt"
	"strings"
)

func ScanValue() string {
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		if strings.Contains(err.Error(), "unexpected newline") {
			return "y"
		} else {
			PanicIfErr(err)
		}
	}

	return response
}

func EnterYesNoQuestion() bool {
	response := ScanValue()
	ok := []string{"y", "Y", "yes", "Yes", "YES"}
	return StringContains(ok, response)
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
	ok := []string{"y", "Y", "yes", "Yes", "YES"}
	patch := []string{"p", "P", "Patch", "patch"}

	if StringContains(ok, response) {
		return ADD
	} else if StringContains(patch, response) {
		return PATCH
	}

	return IGNORE
}
