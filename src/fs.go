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

	err = json.Unmarshal(dat, &jsonData)
	Utils.PanicIfErr(err)

	return jsonData
}

func WriteJSON(filePath string, jsonData map[string]string) {
	absPath := RelativePathToAbs(filePath)

	bytesToWrite, err := json.Marshal(jsonData)
	Utils.PanicIfErr(err)

	err = ioutil.WriteFile(absPath, bytesToWrite, os.ModePerm)
	Utils.PanicIfErr(err)
}

func EditFile(filePath string) {
	VimCmd := exec.Command("vim", filePath)
	VimCmd.Stdin = os.Stdin
	VimCmd.Stdout = os.Stdout
	VimCmd.Stderr = os.Stderr
	err := VimCmd.Run()
	Utils.PanicIfErr(err)
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
