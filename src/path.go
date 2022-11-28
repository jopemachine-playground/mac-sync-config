package src

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/src/utils"
)

const USER_PROFILE_MAGIC_STR = "USER_PROFILE"

func ReplaceUserName(path string) string {
	if strings.HasPrefix(path, "/Users/") {
		return strings.Replace(path, fmt.Sprintf("/Users/%s", Utils.GetCurrentUserName()), fmt.Sprintf("/Users/%s", USER_PROFILE_MAGIC_STR), 1)
	}

	return path
}

// when pathHandlingType is true, it returns str replaced userName with magic string
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
