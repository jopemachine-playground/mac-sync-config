package src

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/src/utils"
)

const userProfileMagicStr = "USER_PROFILE"

func replaceUserName(path string) string {
	if strings.HasPrefix(path, "/Users/") {
		return strings.Replace(path, fmt.Sprintf("/Users/%s", Utils.GetCurrentUserName()), fmt.Sprintf("/Users/%s", userProfileMagicStr), 1)
	}

	return path
}

// when pathHandlingType is true, it returns str replaced userName with magic string
func RelativePathToAbs(path string, shouldReplaceUserName bool) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if shouldReplaceUserName {
		dir = replaceUserName(dir)
	}

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
