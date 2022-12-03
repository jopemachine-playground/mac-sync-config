package src

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	MacSyncConfig "github.com/jopemachine/mac-sync-config/src"
	Utils "github.com/jopemachine/mac-sync-config/utils"
)

type PullPathInfo struct {
	originalPath string
	srcPath      string
	dstPath      string
}

func PullRemoteConfigs(nameFilter string) {
	MacSyncConfig.Logger.ClearConsole()
	remoteCommitHashId := MacSyncConfig.Github.GetRemoteConfigHashId()
	lastChangedConfig := MacSyncConfig.ReadLastChanged()

	if lastChangedConfig["remote-commit-hash-id"] == remoteCommitHashId {
		MacSyncConfig.Logger.Info("Config files already up to date.")
		return
	}

	tempPath := MacSyncConfig.Github.CloneConfigsRepository()
	configs, err := MacSyncConfig.ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempPath, MacSyncConfig.MAC_SYNC_CONFIGS_FILE))

	Utils.PanicIfErr(err)

	configPathsToSync := configs.ConfigPathsToSync
	selectedFilePaths := []PullPathInfo{}
	filteredConfigPathsToSync := []string{}

	for _, configPathToSync := range configPathsToSync {
		if nameFilter != "" && !strings.Contains(filepath.Base(configPathToSync), nameFilter) {
			continue
		}

		configRootPath := fmt.Sprintf("%s/%s", tempPath, MacSyncConfig.GetRemoteConfigFolderName())
		dstPath := fmt.Sprintf("%s%s", configRootPath, MacSyncConfig.ReplaceUserName(MacSyncConfig.RelativePathToAbs(configPathToSync)))

		if haveDiff := MacSyncConfig.Git.IsUpdated(tempPath, dstPath); haveDiff {
			filteredConfigPathsToSync = append(filteredConfigPathsToSync, configPathToSync)
		}
	}

	for configPathIdx, configPathToSync := range filteredConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, MacSyncConfig.GetRemoteConfigFolderName())

		absConfigPathToSync := MacSyncConfig.ReplaceUserName(MacSyncConfig.RelativePathToAbs(configPathToSync))
		srcPath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)

		if _, err := os.Stat(srcPath); errors.Is(err, os.ErrNotExist) {
			MacSyncConfig.Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"%s\", but the config file is not found on the remote repository.\nEnsure to push the config file before pulling.", configPathToSync, MacSyncConfig.MAC_SYNC_CONFIGS_FILE))
			MacSyncConfig.Logger.Log(MacSyncConfig.PRESS_ANYKEY_HELP)
			Utils.WaitResponse()
			MacSyncConfig.Logger.ClearConsole()
			continue
		}

		dstPath := MacSyncConfig.RelativePathToAbs(configPathToSync)

		// To show diff, copy dstPath file to srcPath.
		// This should be reset before copying from dstFile to srcPath.
		if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
			MacSyncConfig.CopyFiles(dstPath, srcPath)
		}

		if MacSyncConfig.Flag_OverWrite {
			if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
				err = os.RemoveAll(dstPath)
				Utils.PanicIfErr(err)
			}

			selectedFilePaths = append(selectedFilePaths, PullPathInfo{
				configPathToSync,
				srcPath,
				dstPath,
			})
		} else {
			progressStr := color.GreenString(fmt.Sprintf("[%d/%d]", configPathIdx+1, len(configPathsToSync)))
			MacSyncConfig.Logger.Info(fmt.Sprintf("%s %s", progressStr, color.MagentaString(path.Base(srcPath))))
			MacSyncConfig.Logger.Log(color.HiBlackString(fmt.Sprintf("Full path: %s", dstPath)))
			MacSyncConfig.Logger.Log(color.New(color.FgCyan, color.Bold).Sprintf(MacSyncConfig.PULL_HELP))

			shouldAdd := true
			userResp := Utils.MakeQuestion(Utils.PULL_CONFIG_ALLOWED_KEYS)

			if userResp == Utils.QUESTION_RESULT_EDIT {
				MacSyncConfig.EditFile(srcPath)
			} else if userResp == Utils.QUESTION_RESULT_SHOW_DIFF {
				MacSyncConfig.Git.ShowDiff(tempPath, srcPath)
				MacSyncConfig.Logger.Log(MacSyncConfig.PRESS_ANYKEY_HELP)
				shouldAdd = Utils.MakeYesNoQuestion()
			} else {
				shouldAdd = false
			}

			if shouldAdd {
				selectedFilePaths = append(selectedFilePaths, PullPathInfo{
					configPathToSync,
					srcPath,
					dstPath,
				})
			}

			MacSyncConfig.Logger.ClearConsole()
		}
	}

	for _, path := range selectedFilePaths {
		MacSyncConfig.Git.Reset(tempPath, path.srcPath)
		if _, err := os.Stat(path.dstPath); !errors.Is(err, os.ErrNotExist) {
			err = os.RemoveAll(path.dstPath)
			Utils.PanicIfErr(err)
		}
		MacSyncConfig.CopyFiles(path.srcPath, path.dstPath)
		MacSyncConfig.Logger.Success(fmt.Sprintf("\"%s\" updated.", path.originalPath))
	}

	if _, err := os.Stat(tempPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempPath, os.ModePerm)
	}

	// If 'nameFilter' is not empty, same commit hash id should not be ignored.
	if nameFilter == "" {
		lastChangedConfig["remote-commit-hash-id"] = remoteCommitHashId
		MacSyncConfig.WriteLastChangedConfigFile(lastChangedConfig)
	}

	if len(selectedFilePaths) > 0 {
		MacSyncConfig.Logger.Info("Local config files are updated successfully.\n  Note that Some changes might require to reboot to apply.")
	} else {
		MacSyncConfig.Logger.Info("Config files already up to date.")
	}
}
