package src

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

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
