package src

import "os"

func GetRemoteConfigFolderName() string {
	if folderName := os.Getenv("MAC_SYNC_CONFIG_FOLDER"); folderName != "" {
		return folderName
	}

	return ".mac-sync-configs"
}

func GetGitBranchName() string {
	if branchEnv := os.Getenv("MAC_SYNC_CONFIG_BRANCH"); branchEnv != "" {
		return branchEnv
	}
	return "main"
}
