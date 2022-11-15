package utils

import (
	"log"
	"os"
	"os/user"
)

func IsRootUser() bool {
	return os.Geteuid() == 0
}

func GetCurrentUserName() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	return user.Username
}
