package src

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/imroc/req/v3"
	Utils "github.com/jopemachine/mac-sync-config/utils"
)

type gitHubManipulator struct{}

var (
	Github gitHubManipulator
)

const GH_BOT_EMAIL = "github-actions[bot] <41898282+github-actions[bot]@users.noreply.github.com>"

func (github gitHubManipulator) CloneConfigsRepository() string {
	tempPath, err := os.MkdirTemp("", "mac-sync-config-temp-")
	Utils.FatalExitIfError(err)

	// Should fully clone repository for commit and push
	gitCloneArgs := strings.Fields(
		fmt.Sprintf("git clone -b %s --single-branch https://github.com/%s/%s %s",
			GetGitBranchName(),
			KeychainPreference.GithubId,
			KeychainPreference.MacSyncConfigGitRepositoryName,
			tempPath))

	gitCloneCmd := exec.Command(gitCloneArgs[0], gitCloneArgs[1:]...)
	_, err = gitCloneCmd.CombinedOutput()
	Utils.FatalExitIfError(err)

	tempConfigDirPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

	if _, err := os.Stat(tempConfigDirPath); errors.Is(err, os.ErrNotExist) {
		Utils.FatalExitIfError(os.Mkdir(tempConfigDirPath, os.ModePerm))
	}

	return tempPath
}

func (github gitHubManipulator) GetRemoteConfigHashId() string {
	gitLsRemoteArgs := strings.Fields(fmt.Sprintf("git ls-remote https://github.com/%s/%s HEAD", KeychainPreference.GithubId, KeychainPreference.MacSyncConfigGitRepositoryName))
	gitLsRemoteCmd := exec.Command(gitLsRemoteArgs[0], gitLsRemoteArgs[1:]...)
	stdout, err := gitLsRemoteCmd.CombinedOutput()
	Utils.FatalExitIfError(err)

	return strings.TrimSpace(strings.Split(fmt.Sprintf("%s", stdout), "HEAD")[0])
}

func (github gitHubManipulator) GetMacSyncConfigs() string {
	resp, err := req.C().R().
		SetHeader("Authorization", fmt.Sprintf("token %s", KeychainPreference.GithubToken)).
		SetHeader("Cache-control", "no-cache").
		SetPathParam("userName", KeychainPreference.GithubId).
		SetPathParam("repoName", KeychainPreference.MacSyncConfigGitRepositoryName).
		SetPathParam("branchName", GetGitBranchName()).
		SetPathParam("fileName", "mac-sync-configs.yaml").
		EnableDump().
		Get("https://raw.githubusercontent.com/{userName}/{repoName}/{branchName}/{fileName}")

	Utils.FatalExitIfError(err)
	return resp.String()
}
