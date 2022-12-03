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

func (github gitHubManipulator) CloneConfigsRepository() string {
	tempPath, err := os.MkdirTemp("", "mac-sync-config-temp-")
	Utils.PanicIfErr(err)

	// Should fully clone repository for commit and push
	gitCloneArgs := strings.Fields(fmt.Sprintf("git clone https://github.com/%s/%s %s", KeychainPreference.GithubId, KeychainPreference.MacSyncConfigGitRepositoryName, tempPath))
	gitCloneCmd := exec.Command(gitCloneArgs[0], gitCloneArgs[1:]...)
	output, err := gitCloneCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)

	tempConfigDirPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

	if _, err := os.Stat(tempConfigDirPath); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(tempConfigDirPath, os.ModePerm)
		Utils.PanicIfErr(err)
	}

	return tempPath
}

func (github gitHubManipulator) GetRemoteConfigHashId() string {
	gitLsRemoteArgs := strings.Fields(fmt.Sprintf("git ls-remote https://github.com/%s/%s HEAD", KeychainPreference.GithubId, KeychainPreference.MacSyncConfigGitRepositoryName))
	gitLsRemoteCmd := exec.Command(gitLsRemoteArgs[0], gitLsRemoteArgs[1:]...)
	stdout, err := gitLsRemoteCmd.CombinedOutput()
	Utils.PanicIfErr(err)

	return strings.TrimSpace(strings.Split(fmt.Sprintf("%s", stdout), "HEAD")[0])
}

func (github gitHubManipulator) GetMacSyncConfigs() string {
	// TODO: Add branch name as env variable
	// TODO: Add file name as env variable
	resp, err := req.C().R().
		SetHeader("Authorization", fmt.Sprintf("token %s", KeychainPreference.GithubToken)).
		SetHeader("Cache-control", "no-cache").
		SetPathParam("userName", KeychainPreference.GithubId).
		SetPathParam("repoName", KeychainPreference.MacSyncConfigGitRepositoryName).
		SetPathParam("branchName", "main").
		SetPathParam("fileName", "mac-sync-configs.yaml").
		EnableDump().
		Get("https://raw.githubusercontent.com/{userName}/{repoName}/{branchName}/{fileName}")

	Utils.PanicIfErr(err)
	return resp.String()
}
