package src

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/utils"
)

var User_Profile = "DEFAULT_USER_PROFILE"

func ReplaceUserName(path string) string {
	if userProfile := ReadLocalPreference()["profile"]; userProfile != "" {
		User_Profile = userProfile
	}

	// Overwrite the profile if env variable set
	if userProfileEnv := os.Getenv("MAC_SYNC_CONFIG_USER_PROFILE"); userProfileEnv != "" {
		User_Profile = userProfileEnv
	}

	if strings.HasPrefix(path, "/Users/") {
		return strings.Replace(path, fmt.Sprintf("/Users/%s", Utils.GetMacosUserName()), fmt.Sprintf("/Users/%s", User_Profile), 1)
	}

	return path
}

func RelativePathToAbs(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	}

	return path
}

func HandleWhiteSpaceInPath(path string) string {
	return strings.ReplaceAll(path, " ", "\\ ")
}
