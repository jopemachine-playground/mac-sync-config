package utils

import "log"

// TODO: Select proper error handling way on every error handling logic.
func FatalExitIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
