package utils

import "log"

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
