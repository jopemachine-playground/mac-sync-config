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

type PushPathInfo struct {
	originalPath  string
	convertedPath string
}

func PushConfigFiles(profileName string) {
	if profileName != "" {
		os.Setenv("MAC_SYNC_CONFIG_USER_PROFILE", profileName)
	}

	MacSyncConfig.Logger.ClearConsole()

	tempPath := MacSyncConfig.Github.CloneConfigsRepository()
	configs, err := MacSyncConfig.ReadMacSyncConfigFile(fmt.Sprintf("%s/%s", tempPath, MacSyncConfig.MAC_SYNC_CONFIGS_FILE))
	Utils.PanicIfErr(err)

	var updatedFilePaths = []PushPathInfo{}
	var selectedUpdatedFilePaths = []PushPathInfo{}

	for _, configPathToSync := range configs.ConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, MacSyncConfig.GetRemoteConfigFolderName())
		absSrcConfigPathToSync := MacSyncConfig.RelativePathToAbs(configPathToSync)

		dstPath := fmt.Sprintf("%s%s", configRootPath, MacSyncConfig.ReplaceUserName(MacSyncConfig.RelativePathToAbs(configPathToSync)))

		// Delete files for update if the files already exist
		if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
			err := os.RemoveAll(dstPath)
			Utils.PanicIfErr(err)
		}

		if _, err := os.Stat(absSrcConfigPathToSync); errors.Is(err, os.ErrNotExist) {
			MacSyncConfig.Logger.Warning(fmt.Sprintf("\"%s\" not found in the local computer.", configPathToSync))
			Utils.WaitResponse()
			MacSyncConfig.Logger.Log(MacSyncConfig.PRESS_ANYKEY_HELP_MSG)
			MacSyncConfig.Logger.ClearConsole()
			continue
		}

		MacSyncConfig.CopyFiles(absSrcConfigPathToSync, dstPath)
		Utils.PanicIfErr(err)

		if haveDiff := MacSyncConfig.Git.IsUpdated(tempPath, dstPath); haveDiff {
			updatedFilePaths = append(updatedFilePaths, PushPathInfo{configPathToSync, dstPath})
		}
	}

	if MacSyncConfig.Flag_OverWrite {
		MacSyncConfig.Git.AddAllFiles(tempPath)
		selectedUpdatedFilePaths = updatedFilePaths
	} else {
		for fileIdx, updatedFilePath := range updatedFilePaths {
			progressStr := color.GreenString(fmt.Sprintf("[%d/%d]", fileIdx+1, len(updatedFilePaths)))
			MacSyncConfig.Logger.Info(fmt.Sprintf("%s %s", progressStr, color.MagentaString(path.Base(updatedFilePath.convertedPath))))
			MacSyncConfig.Logger.Log(color.New(color.FgCyan, color.Bold).Sprint(MacSyncConfig.PUSH_HELP_MSG))

			userResp := Utils.MakeQuestion(Utils.PUSH_CONFIG_ALLOWED_KEYS)
			shouldAdd := true
			partiallyPatched := false

			if userResp == Utils.QUESTION_RESULT_SHOW_DIFF {
				MacSyncConfig.Git.ShowDiff(tempPath, updatedFilePath.convertedPath)
				MacSyncConfig.Logger.Log(MacSyncConfig.PRESS_ANYKEY_HELP_MSG)
				shouldAdd = Utils.MakeYesNoQuestion()
			} else if userResp == Utils.QUESTION_RESULT_EDIT {
				MacSyncConfig.EditFile(updatedFilePath.convertedPath)
			} else if userResp == Utils.QUESTION_RESULT_PATCH {
				MacSyncConfig.Git.PatchFile(tempPath, updatedFilePath.convertedPath)
				partiallyPatched = true
			} else if userResp == Utils.QUESTION_RESULT_IGNORE {
				shouldAdd = false
			}

			if shouldAdd {
				if !partiallyPatched {
					MacSyncConfig.Git.AddFile(tempPath, updatedFilePath.convertedPath)
				}
				selectedUpdatedFilePaths = append(selectedUpdatedFilePaths, updatedFilePath)
			}

			MacSyncConfig.Logger.ClearConsole()
		}
	}

	MacSyncConfig.Logger.NewLine()

	if len(selectedUpdatedFilePaths) > 0 {
		MacSyncConfig.Git.Commit(tempPath)
		MacSyncConfig.Git.Push(tempPath)

		MacSyncConfig.Logger.NewLine()

		for _, selectedFilePath := range selectedUpdatedFilePaths {
			MacSyncConfig.Logger.Success(fmt.Sprintf("\"%s\" updated.", selectedFilePath.originalPath))
		}

		MacSyncConfig.Logger.NewLine()

		MacSyncConfig.Logger.Info("Config files pushed successfully.")
	} else {
		MacSyncConfig.Logger.Info("No file pushed.")
	}

	os.RemoveAll(tempPath)
}
