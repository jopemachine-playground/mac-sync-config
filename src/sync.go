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

func CopyFiles(srcPath string, dstPath string) {
	dirPath := filepath.Dir(dstPath)

	mkdirCmd := exec.Command("mkdir", "-p", dirPath)
	output, err := mkdirCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)

	cpCmd := exec.Command("cp", "-fR", srcPath, dstPath)
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

		dstPath := fmt.Sprintf("%s%s", configRootPath, ReplaceUserName(RelativePathToAbs(configPathToSync)))

		// Delete files for update if the files already exist
		if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
			err := os.RemoveAll(dstPath)
			Utils.PanicIfErr(err)
		}

		if _, err := os.Stat(absSrcConfigPathToSync); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" not found in the local computer.", configPathToSync))
			Utils.WaitResponse()
			Logger.ClearConsole()
			continue
		}

		CopyFiles(absSrcConfigPathToSync, dstPath)
		Utils.PanicIfErr(err)

		if haveDiff := Git.IsUpdated(tempPath, dstPath); haveDiff {
			updatedFilePaths = append(updatedFilePaths, PushPath{configPathToSync, dstPath})
		}
	}

	if Flag_OverWrite {
		Git.AddAll(tempPath)
		selectedUpdatedFilePaths = updatedFilePaths
	} else {
		for fileIdx, updatedFilePath := range updatedFilePaths {
			progressStr := color.GreenString(fmt.Sprintf("[%d/%d]", fileIdx+1, len(updatedFilePaths)))
			Logger.Info(fmt.Sprintf("%s Diff of %s\n", progressStr, color.MagentaString(path.Base(updatedFilePath.convertedPath))))
			Git.ShowDiff(tempPath, updatedFilePath.convertedPath)

			Logger.Question(color.New(color.FgCyan, color.Bold).Sprint("Press 'y' for adding the file, 'n' to ignore, 'p' for patching."))
			userRes := Utils.CreateQuestion(Utils.PUSH_CONFIG_ALLOWED_KEYS)

			if userRes != Utils.QUESTION_RESULT_IGNORE {
				selectedUpdatedFilePaths = append(selectedUpdatedFilePaths, updatedFilePath)
			}

			if userRes == Utils.QUESTION_RESULT_PATCH {
				Git.PatchFile(tempPath, updatedFilePath.convertedPath)
			} else if userRes == Utils.QUESTION_RESULT_ADD {
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

func PullRemoteConfigs(nameFilter string) {
	remoteCommitHashId := Git.GetRemoteConfigHashId()
	lastChangedConfig := ReadLastChanged()

	if lastChangedConfig["remote-commit-hash-id"] == remoteCommitHashId {
		Logger.Info("Config files already up to date.")
		return
	}

	tempPath := Git.CloneConfigsRepository()
	configs, err := ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempPath, MAC_SYNC_CONFIGS_FILE))

	Utils.PanicIfErr(err)

	configPathsToSync := configs.ConfigPathsToSync
	selectedFilePaths := []PullPath{}
	filteredConfigPathsToSync := []string{}

	for _, configPathToSync := range configPathsToSync {
		if nameFilter != "" && !strings.Contains(filepath.Base(configPathToSync), nameFilter) {
			continue
		}

		filteredConfigPathsToSync = append(filteredConfigPathsToSync, configPathToSync)
	}

	for configPathIdx, configPathToSync := range filteredConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

		absConfigPathToSync := ReplaceUserName(RelativePathToAbs(configPathToSync))
		srcPath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)

		if _, err := os.Stat(srcPath); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"%s\", but the config file is not found on the remote repository.\nEnsure to push the config file before pulling.", configPathToSync, MAC_SYNC_CONFIGS_FILE))
			Utils.WaitResponse()
			Logger.ClearConsole()
			continue
		}

		dstPath := RelativePathToAbs(configPathToSync)

		// To show diff, copy dstPath file to srcPath.
		// This should be reset before copying from dstFile to srcPath.
		if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
			CopyFiles(dstPath, srcPath)
		}

		if Flag_OverWrite {
			if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
				err = os.RemoveAll(dstPath)
				Utils.PanicIfErr(err)
			}

			selectedFilePaths = append(selectedFilePaths, PullPath{
				configPathToSync,
				srcPath,
				dstPath,
			})
		} else {
			progressStr := color.GreenString(fmt.Sprintf("[%d/%d]", configPathIdx+1, len(configPathsToSync)))
			Logger.Info(fmt.Sprintf("%s Diff of %s\n", progressStr, color.MagentaString(path.Base(srcPath))))

			// Git.ShowDiff(tempPath, srcPath)
			Logger.Question(color.New(color.FgCyan, color.Bold).Sprintf(
				"Press 'y' to update '%s', 'n' to ignore.", path.Base(dstPath)))

			Logger.Log(color.HiBlackString(fmt.Sprintf("Full path: %s", dstPath)))

			if yes := Utils.EnterYesNoQuestion(); yes {
				selectedFilePaths = append(selectedFilePaths, PullPath{
					configPathToSync,
					srcPath,
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

	// If 'nameFilter' is not empty, same commit hash id should not be ignored.
	if nameFilter == "" {
		lastChangedConfig["remote-commit-hash-id"] = remoteCommitHashId
		WriteLastChangedConfigFile(lastChangedConfig)
	}

	Logger.NewLine()
	Logger.Info("Local config files are updated successfully.\n  Note that Some changes might require to reboot to apply.")
}
