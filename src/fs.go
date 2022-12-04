package src

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	Utils "github.com/jopemachine/mac-sync-config/utils"
)

func ReadJSON(filePath string) map[string]string {
	absPath := RelativePathToAbs(filePath)

	if _, err := os.Stat(absPath); errors.Is(err, os.ErrNotExist) {
		return make(map[string]string)
	}

	dat, err := ioutil.ReadFile(absPath)
	Utils.FatalExitIfError(err)

	var jsonData map[string]string

	Utils.FatalExitIfError(json.Unmarshal(dat, &jsonData))

	return jsonData
}

func WriteJSON(filePath string, jsonData map[string]string) {
	absPath := RelativePathToAbs(filePath)

	bytesToWrite, err := json.Marshal(jsonData)
	Utils.FatalExitIfError(err)

	Utils.FatalExitIfError(ioutil.WriteFile(absPath, bytesToWrite, os.ModePerm))
}

func EditFile(filePath string) {
	VimCmd := exec.Command("vim", filePath)
	VimCmd.Stdin = os.Stdin
	VimCmd.Stdout = os.Stdout
	VimCmd.Stderr = os.Stderr
	Utils.FatalExitIfError(VimCmd.Run())
}

func CopyFiles(srcPath string, dstPath string) {
	if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
		Utils.FatalExitIfError(os.RemoveAll(dstPath))
	}

	mkdirCmd := exec.Command("mkdir", "-p", filepath.Dir(dstPath))
	_, err := mkdirCmd.CombinedOutput()
	Utils.FatalExitIfError(err)

	// a: Archive mode. Preserves structure and attributes of files but not directory structure.
	// f: If the destination file cannot be opened, remove it and create a new file.
	// R: If source_file designates a directory, cp copies the directory and the entire subtree connected at that point.
	// H: Symbolic links on the command line are followed.
	// L: All symbolic links are followed.
	cpCmd := exec.Command("cp", "-afRHL", srcPath, dstPath)
	_, err = cpCmd.CombinedOutput()
	Utils.FatalExitIfError(err)
}
