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

func GetProfileName() string {
	if profileEnv := os.Getenv("MAC_SYNC_CONFIG_PROFILE"); profileEnv != "" {
		return profileEnv
	}

	if profilePreference := ReadLocalPreference()["profile"]; profilePreference != "" {
		return profilePreference
	}

	return "DEFAULT_USER_PROFILE"
}
