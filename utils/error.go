package utils

import "log"

func FatalExitIfError(err error) {
	if err != nil {
		if Flags.UsePanic {
			panic(err)
		} else {
			log.Fatal(err)
		}
	}
}
