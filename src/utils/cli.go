package utils

import "fmt"

func EnterYesNoQuestion() bool {
	var response string
	_, err := fmt.Scanln(&response)
	PanicIfErr(err)
	ok := []string{"y", "Y", "yes", "Yes", "YES"}
	return StringContains(ok, response)
}
