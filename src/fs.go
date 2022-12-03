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
	Utils.PanicIfErr(err)

	var jsonData map[string]string

	Utils.PanicIfErr(json.Unmarshal(dat, &jsonData))

	return jsonData
}

func WriteJSON(filePath string, jsonData map[string]string) {
	absPath := RelativePathToAbs(filePath)

	bytesToWrite, err := json.Marshal(jsonData)
	Utils.PanicIfErr(err)

	Utils.PanicIfErr(ioutil.WriteFile(absPath, bytesToWrite, os.ModePerm))
}

func EditFile(filePath string) {
	VimCmd := exec.Command("vim", filePath)
	VimCmd.Stdin = os.Stdin
	VimCmd.Stdout = os.Stdout
	VimCmd.Stderr = os.Stderr
	Utils.PanicIfErr(VimCmd.Run())
}

func CopyFiles(srcPath string, dstPath string) {
	dirPath := filepath.Dir(dstPath)

	mkdirCmd := exec.Command("mkdir", "-p", dirPath)
	output, err := mkdirCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)

	cpCmd := exec.Command("cp", "-fR", srcPath, dstPath)
	output, err = cpCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}
