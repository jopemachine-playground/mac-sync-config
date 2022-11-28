package src

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/src/utils"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

type PushPath struct {
	originalPath  string
	convertedPath string
}

type PullPath struct {
	originalPath string
	srcPath      string
	dstPath      string
}

func CopyFiles(targetFilePath string, dstFilePath string) {
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

func PushConfigFiles() {
	tempPath := Git.CloneConfigsRepository()
	configs, err := ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempPath, MAC_SYNC_CONFIGS_FILE))
	Utils.PanicIfErr(err)

	var updatedFilePaths = []PushPath{}
	var selectedUpdatedFilePaths = []PushPath{}

	for _, configPathToSync := range configs.ConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())
		absSrcConfigPathToSync := RelativePathToAbs(configPathToSync)

		dstFilePath := fmt.Sprintf("%s%s", configRootPath, ReplaceUserName(RelativePathToAbs(configPathToSync)))

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

		CopyFiles(absSrcConfigPathToSync, dstFilePath)
		Utils.PanicIfErr(err)

		if haveDiff := Git.IsUpdated(tempPath, dstFilePath); haveDiff {
			updatedFilePaths = append(updatedFilePaths, PushPath{configPathToSync, dstFilePath})
		}
	}

	if Flag_OverWrite {
		Git.AddAll(tempPath)
		selectedUpdatedFilePaths = updatedFilePaths
	} else {
		for fileIdx, updatedFilePath := range updatedFilePaths {
			progressStr := fmt.Sprintf("[%d/%d]", fileIdx+1, len(updatedFilePaths))
			Logger.Info(fmt.Sprintf("%s Diff of %s\n", progressStr, color.MagentaString(path.Base(updatedFilePath.convertedPath))))
			Git.ShowDiff(tempPath, updatedFilePath.convertedPath)

			Logger.Question(color.CyanString("Press 'y' for adding the file, 'n' to ignore, 'p' for patching."))
			userRes := Utils.ConfigAddQuestion()

			if userRes != Utils.IGNORE {
				selectedUpdatedFilePaths = append(selectedUpdatedFilePaths, updatedFilePath)
			}

			if userRes == Utils.PATCH {
				Git.PatchFile(tempPath, updatedFilePath.convertedPath)
			} else if userRes == Utils.ADD {
				Git.AddFile(tempPath, updatedFilePath.convertedPath)
			}

			Logger.ClearConsole()
		}
	}

	Logger.NewLine()

	for _, selectedFilePath := range selectedUpdatedFilePaths {
		Logger.Success(fmt.Sprintf("\"%s\" updated.", selectedFilePath.originalPath))
	}

	Logger.NewLine()

	if len(selectedUpdatedFilePaths) > 0 {
		Git.Commit(tempPath)
		Git.Push(tempPath)

		Logger.Info("Config files pushed successfully.")
	} else {
		Logger.Info("No file pushed.")
	}

	os.RemoveAll(tempPath)
}

func PullRemoteConfigs(argFilter string) {
	remoteCommitHashId := Git.GetRemoteConfigHashId()
	lastChangedConfig := ReadLastChanged()

	if lastChangedConfig["remote-commit-hash-id"] == remoteCommitHashId {
		Logger.Info("Config files already up to date.")
	}

	tempPath := Git.CloneConfigsRepository()
	configs, err := ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempPath, MAC_SYNC_CONFIGS_FILE))

	Utils.PanicIfErr(err)

	configPathsToSync := configs.ConfigPathsToSync
	selectedFilePaths := []PullPath{}
	filteredConfigPathsToSync := []string{}

	for _, configPathToSync := range configPathsToSync {
		if argFilter != "" && !strings.Contains(filepath.Base(configPathToSync), argFilter) {
			continue
		}

		filteredConfigPathsToSync = append(filteredConfigPathsToSync, configPathToSync)
	}

	for configPathIdx, configPathToSync := range filteredConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

		absConfigPathToSync := ReplaceUserName(RelativePathToAbs(configPathToSync))
		srcFilePath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)

		if _, err := os.Stat(srcFilePath); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"%s\", but the config file is not found on the remote repository.\nEnsure to push the config file before pulling.", configPathToSync, MAC_SYNC_CONFIGS_FILE))
			Utils.WaitResponse()
			Logger.ClearConsole()
			continue
		}

		dstPath := RelativePathToAbs(configPathToSync)

		// To show diff, copy dstPath file to srcFilePath.
		// This should be reset before copying from dstFile to srcFilePath.
		CopyFiles(dstPath, srcFilePath)

		if Flag_OverWrite {
			if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
				err = os.RemoveAll(dstPath)
				Utils.PanicIfErr(err)
			}

			selectedFilePaths = append(selectedFilePaths, PullPath{
				configPathToSync,
				srcFilePath,
				dstPath,
			})
		} else {
			progressStr := fmt.Sprintf("[%d/%d]", configPathIdx+1, len(configPathsToSync))
			Logger.Info(fmt.Sprintf("%s Diff of %s\n", progressStr, color.MagentaString(path.Base(srcFilePath))))

			Git.ShowDiff(tempPath, srcFilePath)
			Logger.Question(color.CyanString(fmt.Sprintf("Press 'y' to update '%s', 'n' to ignore.", path.Base(dstPath))))
			Logger.Log(color.HiBlackString(fmt.Sprintf("Full path: %s", dstPath)))

			if yes := Utils.EnterYesNoQuestion(); yes {
				selectedFilePaths = append(selectedFilePaths, PullPath{
					configPathToSync,
					srcFilePath,
					dstPath,
				})
			}

			Logger.ClearConsole()
		}
	}

	for _, path := range selectedFilePaths {
		Git.Reset(tempPath, path.srcPath)
		if _, err := os.Stat(path.dstPath); !errors.Is(err, os.ErrNotExist) {
			err = os.RemoveAll(path.dstPath)
			Utils.PanicIfErr(err)
		}
		CopyFiles(path.srcPath, path.dstPath)
		Logger.Success(fmt.Sprintf("\"%s\" updated.", path.originalPath))
	}

	if _, err := os.Stat(tempPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempPath, os.ModePerm)
	}

	// If 'argFilter' is not empty, same commit hash id should not be ignored.
	if argFilter == "" {
		lastChangedConfig["remote-commit-hash-id"] = remoteCommitHashId
		WriteLastChangedConfigFile(lastChangedConfig)
	}

	Logger.Info("Local config files are updated successfully.\nNote that Some changes might require to reboot to apply.")
}
