package src

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/fatih/color"
	MacSyncConfig "github.com/jopemachine/mac-sync-config/src"
	Utils "github.com/jopemachine/mac-sync-config/utils"
)

type PullPathInfo struct {
	originalPath         string
	remoteConfigFilePath string
	localConfigFilePath  string
}

func PullRemoteConfigs(profileName string) {
	if profileName != "" {
		os.Setenv("MAC_SYNC_CONFIG_PROFILE", profileName)
	}

	MacSyncConfig.Logger.ClearConsole()

	tempConfigsRepoDirPath := MacSyncConfig.Github.CloneConfigsRepository()
	macSyncConfigs := MacSyncConfig.ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempConfigsRepoDirPath, MacSyncConfig.MAC_SYNC_CONFIGS_FILE))

	configPathsToSync := macSyncConfigs.ConfigPathsToSync
	selectedFilePaths := []PullPathInfo{}
	filteredConfigPathsToSync := []string{}

	for _, configPathToSync := range configPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempConfigsRepoDirPath, MacSyncConfig.GetRemoteConfigFolderName())
		absConfigPathToSync := MacSyncConfig.ReplaceMacOSUserName(MacSyncConfig.RelativePathToAbs(configPathToSync))

		remoteConfigFilePath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)
		localConfigFilePath := MacSyncConfig.RelativePathToAbs(configPathToSync)

		if _, err := os.Stat(remoteConfigFilePath); errors.Is(err, os.ErrNotExist) {
			MacSyncConfig.Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"%s\", but the config file is not found on the remote repository.", configPathToSync, MacSyncConfig.MAC_SYNC_CONFIGS_FILE))
			MacSyncConfig.Logger.Log(MacSyncConfig.PRESS_ANYKEY_HELP_MSG)
			Utils.WaitResponse()
			MacSyncConfig.Logger.ClearConsole()
			continue
		}

		// To find out if it is updated and to show diff, copy dstPath file to srcPath.
		// This should be reset before copying from dstFile to srcPath.
		if _, err := os.Stat(localConfigFilePath); !errors.Is(err, os.ErrNotExist) {
			MacSyncConfig.CopyFiles(localConfigFilePath, remoteConfigFilePath)
		}

		if diffExist := MacSyncConfig.Git.IsUpdated(tempConfigsRepoDirPath, remoteConfigFilePath); diffExist {
			filteredConfigPathsToSync = append(filteredConfigPathsToSync, configPathToSync)
		}
	}

	for configPathIdx, configPathToSync := range filteredConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempConfigsRepoDirPath, MacSyncConfig.GetRemoteConfigFolderName())
		absConfigPathToSync := MacSyncConfig.ReplaceMacOSUserName(MacSyncConfig.RelativePathToAbs(configPathToSync))

		remoteConfigFilePath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)
		localConfigFilePath := MacSyncConfig.RelativePathToAbs(configPathToSync)

		if Utils.Flags.Overwrite {
			selectedFilePaths = append(selectedFilePaths, PullPathInfo{
				configPathToSync,
				remoteConfigFilePath,
				localConfigFilePath,
			})
		} else {
			progressStr := color.GreenString(fmt.Sprintf("[%d/%d]", configPathIdx+1, len(filteredConfigPathsToSync)))
			MacSyncConfig.Logger.Info(color.New(color.Bold).Sprintf(
				fmt.Sprintf("%s %s", progressStr, color.MagentaString(path.Base(remoteConfigFilePath)))))

			MacSyncConfig.Logger.Log(color.HiBlackString(fmt.Sprintf("Full path: %s", localConfigFilePath)))
			MacSyncConfig.Logger.Log(color.New(color.FgCyan).Sprintf(MacSyncConfig.PULL_HELP_MSG))

			shouldAdd := true
			userResp := Utils.MakeQuestion(Utils.PULL_CONFIG_ALLOWED_KEYS)

			if userResp == Utils.QUESTION_RESULT_EDIT {
				MacSyncConfig.EditFile(remoteConfigFilePath)
			} else if userResp == Utils.QUESTION_RESULT_SHOW_DIFF {
				MacSyncConfig.Git.ShowDiff(tempConfigsRepoDirPath, remoteConfigFilePath)
				MacSyncConfig.Logger.Log(MacSyncConfig.PRESS_ANYKEY_HELP_MSG)
				shouldAdd = Utils.MakeYesNoQuestion()
			} else if userResp == Utils.QUESTION_RESULT_IGNORE {
				shouldAdd = false
			}

			if shouldAdd {
				selectedFilePaths = append(selectedFilePaths, PullPathInfo{
					configPathToSync,
					remoteConfigFilePath,
					localConfigFilePath,
				})
			}

			MacSyncConfig.Logger.ClearConsole()
		}
	}

	for _, path := range selectedFilePaths {
		MacSyncConfig.Git.Reset(tempConfigsRepoDirPath, path.remoteConfigFilePath)
		MacSyncConfig.CopyFiles(path.remoteConfigFilePath, path.localConfigFilePath)
		MacSyncConfig.Logger.Success(fmt.Sprintf("\"%s\" updated.", path.originalPath))
	}

	Utils.FatalExitIfError(os.RemoveAll(tempConfigsRepoDirPath))

	if len(selectedFilePaths) > 0 {
		MacSyncConfig.Logger.Info(color.New(color.FgCyan, color.Bold).Sprintf("Local config files are updated successfully."))
	} else {
		MacSyncConfig.Logger.Info("Config files already up to date.")
	}
}
