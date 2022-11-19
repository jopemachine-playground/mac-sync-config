package src

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/src/utils"
	"gopkg.in/yaml.v3"
)

type path struct {
	originalPath  string
	convertedPath string
}

func CopyConfigs(targetFilePath string, dstFilePath string) {
	dirPath := filepath.Dir(dstFilePath)

	mkdirCmd := exec.Command("mkdir", "-p", dirPath)
	output, err := mkdirCmd.CombinedOutput()
	Utils.PanicIfErrWithOutput(string(output), err)

	cpCmd := exec.Command("cp", "-fR", targetFilePath, dstFilePath)
	output, err = cpCmd.CombinedOutput()
	Utils.PanicIfErrWithOutput(string(output), err)
}

func ReadMacSyncConfigFile(filepath string) (ConfigInfo, error) {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		return ConfigInfo{}, err
	}

	dat, err := ioutil.ReadFile(filepath)

	Utils.PanicIfErr(err)

	var config ConfigInfo

	err = yaml.Unmarshal(dat, &config)
	Utils.PanicIfErr(err)

	return config, nil
}

func PullRemoteConfigs(argFilter string) {
	remoteCommitHashId := FetchRemoteConfigCommitHashId()
	configFileLastChanged := ReadConfigFileLastChanged()

	if configFileLastChanged["remote-commit-hash-id"] == remoteCommitHashId {
		Logger.Info("Config files already up to date.")
	}

	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))

	Utils.PanicIfErr(err)

	configPathsToSync := configs.ConfigPathsToSync

	for _, configPathToSync := range configPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

		if argFilter != "" && strings.Contains(filepath.Base(configPathToSync), argFilter) == false {
			continue
		}

		absConfigPathToSync := HandleRelativePath(configPathToSync, true)
		srcFilePath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)

		if _, err := os.Stat(srcFilePath); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"%s\", but the config file is not found on the remote repository.\nEnsure to push the config file before pulling.", configPathToSync, MacSyncConfigsFile))
			continue
		}

		dstPath := HandleRelativePath(configPathToSync, false)
		selectedFilePaths := []string{}

		if Flag_OverWrite {
			if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
				err = os.RemoveAll(dstPath)
				Utils.PanicIfErr(err)
			}

			CopyConfigs(srcFilePath, dstPath)
			selectedFilePaths = append(selectedFilePaths, configPathToSync)
		} else {
			if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
				CopyConfigs(dstPath, srcFilePath)
				ShowDiff(tempPath, srcFilePath)
				Logger.Question(fmt.Sprintf("\"%s\" Update? (Y/N)", dstPath))

				if yes := Utils.EnterYesNoQuestion(); yes {
					GitReset(tempPath, srcFilePath)
					err = os.RemoveAll(dstPath)
					Utils.PanicIfErr(err)
					CopyConfigs(srcFilePath, dstPath)
					selectedFilePaths = append(selectedFilePaths, configPathToSync)
				}
			} else {
				CopyConfigs(srcFilePath, dstPath)
				selectedFilePaths = append(selectedFilePaths, configPathToSync)
			}
		}

		for _, selectedFilePath := range selectedFilePaths {
			Logger.Success(fmt.Sprintf("\"%s\" updated.", selectedFilePath))
		}
	}

	if _, err := os.Stat(tempPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempPath, os.ModePerm)
	}

	if argFilter == "" {
		configFileLastChanged["remote-commit-hash-id"] = remoteCommitHashId
		WriteConfigFileLastChanged(configFileLastChanged)
		Logger.Info("Local config files are updated. Some changes might require to reboot to apply.")
	}
}

func PushConfigFiles() {
	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))
	Utils.PanicIfErr(err)

	var selectedFilePaths = []path{}
	var updatedFilePaths = []path{}

	for _, configPathToSync := range configs.ConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())
		absConfigPathToSync := HandleRelativePath(configPathToSync, false)

		dstFilePath := fmt.Sprintf("%s%s", configRootPath, HandleRelativePath(configPathToSync, false))

		// Delete files for update if the files already exist
		if _, err := os.Stat(dstFilePath); !errors.Is(err, os.ErrNotExist) {
			err := os.RemoveAll(dstFilePath)
			Utils.PanicIfErr(err)
		}

		if _, err := os.Stat(absConfigPathToSync); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" not found in the local computer.", configPathToSync))
			continue
		}

		CopyConfigs(absConfigPathToSync, dstFilePath)
		Utils.PanicIfErr(err)
		updatedFilePaths = append(updatedFilePaths, path{configPathToSync, dstFilePath})
	}

	if Flag_OverWrite {
		GitAddCwd(tempPath)
		selectedFilePaths = updatedFilePaths
	} else {
		for _, updatedFilePath := range updatedFilePaths {
			// if haveDiff := IsUpdated(tempPath, updatedFilePath.convertedPath); haveDiff {
			// 	ShowDiff(tempPath, updatedFilePath.convertedPath)

			// 	Logger.NewLine()
			// 	Logger.Question("Update? (Y/N)")
			// 	if yes := Utils.EnterYesNoQuestion(); yes {
		// 		GitAddFile(tempPath, updatedFilePath.convertedPath)
			// 		selectedFilePaths = append(selectedFilePaths, updatedFilePath)
			// 	}
			// }

			GitAddPatchFile(tempPath, updatedFilePath.convertedPath)
			selectedFilePaths = append(selectedFilePaths, updatedFilePath)

			Logger.NewLine()
		}
	}

	for _, selectedFilePath := range selectedFilePaths {
		Logger.Success(fmt.Sprintf("\"%s\" updated.", selectedFilePath.originalPath))
	}

	GitCommit(tempPath)
	GitPush(tempPath)

	Logger.Info("Config files updated successfully.")
	os.RemoveAll(tempPath)
}
