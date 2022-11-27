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
	Utils.PanicIfErrWithMsg(string(output), err)

	cpCmd := exec.Command("cp", "-fR", targetFilePath, dstFilePath)
	output, err = cpCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
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
	remoteCommitHashId := GitGetRemoteConfigHashId()
	configFileLastChanged := ReadConfigFileLastChanged()

	if configFileLastChanged["remote-commit-hash-id"] == remoteCommitHashId {
		Logger.Info("Config files already up to date.")
	}

	tempPath := GitCloneConfigsRepository()
	configs, err := ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))

	Utils.PanicIfErr(err)

	configPathsToSync := configs.ConfigPathsToSync

	for _, configPathToSync := range configPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

		if argFilter != "" && strings.Contains(filepath.Base(configPathToSync), argFilter) == false {
			continue
		}

		absConfigPathToSync := RelativePathToAbs(configPathToSync, true)
		srcFilePath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)

		if _, err := os.Stat(srcFilePath); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"%s\", but the config file is not found on the remote repository.\nEnsure to push the config file before pulling.", configPathToSync, MacSyncConfigsFile))
			Utils.WaitResponse()
			Logger.ClearConsole()
			continue
		}

		dstPath := RelativePathToAbs(configPathToSync, false)
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
				GitShowDiff(tempPath, srcFilePath)
				Logger.Question(fmt.Sprintf("\"%s\" Press 'y' for adding the file, 'n' to ignore", dstPath))

				if yes := Utils.EnterYesNoQuestion(); yes {
					GitReset(tempPath, srcFilePath)
					err = os.RemoveAll(dstPath)
					Utils.PanicIfErr(err)
					CopyConfigs(srcFilePath, dstPath)
					selectedFilePaths = append(selectedFilePaths, configPathToSync)
				}

				Logger.ClearConsole()
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
	}

	Logger.Info("Local config files are updated successfully.\nNote that Some changes might require to reboot to apply.")
}

func PushConfigFiles() {
	tempPath := GitCloneConfigsRepository()
	configs, err := ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))
	Utils.PanicIfErr(err)

	var selectedFilePaths = []path{}
	var updatedFilePaths = []path{}

	for _, configPathToSync := range configs.ConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())
		absSrcConfigPathToSync := RelativePathToAbs(configPathToSync, false)

		dstFilePath := fmt.Sprintf("%s%s", configRootPath, RelativePathToAbs(configPathToSync, true))

		// Delete files for update if the files already exist
		if _, err := os.Stat(dstFilePath); !errors.Is(err, os.ErrNotExist) {
			err := os.RemoveAll(dstFilePath)
			Utils.PanicIfErr(err)
		}

		if _, err := os.Stat(absSrcConfigPathToSync); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" not found in the local computer.", configPathToSync))
			Utils.WaitResponse()
			Logger.ClearConsole()
			continue
		}

		CopyConfigs(absSrcConfigPathToSync, dstFilePath)
		Utils.PanicIfErr(err)
		updatedFilePaths = append(updatedFilePaths, path{configPathToSync, dstFilePath})
	}

	if Flag_OverWrite {
		GitAddAll(tempPath)
		selectedFilePaths = updatedFilePaths
	} else {
		for _, updatedFilePath := range updatedFilePaths {
			if haveDiff := IsUpdated(tempPath, updatedFilePath.convertedPath); haveDiff {
				GitShowDiff(tempPath, updatedFilePath.convertedPath)

				Logger.Question("Press 'y' for adding the file, 'n' to ignore, 'p' for patching")
				userRes := Utils.ConfigAddQuestion()

				if userRes != Utils.IGNORE {
					selectedFilePaths = append(selectedFilePaths, updatedFilePath)
				}

				if userRes == Utils.PATCH {
					GitPatchFile(tempPath, updatedFilePath.convertedPath)
				} else if userRes == Utils.ADD {
					GitAddFile(tempPath, updatedFilePath.convertedPath)
				}

				Logger.ClearConsole()
			}
		}
	}

	Logger.NewLine()

	for _, selectedFilePath := range selectedFilePaths {
		Logger.Success(fmt.Sprintf("\"%s\" updated.", selectedFilePath.originalPath))
	}

	Logger.NewLine()
	GitCommit(tempPath)
	GitPush(tempPath)

	Logger.Info("Config files pushed successfully.")
	os.RemoveAll(tempPath)
}
