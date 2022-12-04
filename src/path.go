package src

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/utils"
)

func ReplaceMacOSUserName(path string) string {
	if strings.HasPrefix(path, "/Users/") {
		return strings.Replace(path,
			fmt.Sprintf("/Users/%s", Utils.GetMacosUserName()),
			fmt.Sprintf("/Users/%s", GetProfileName()), 1)
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
