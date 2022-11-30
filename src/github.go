package src

import (
	"fmt"

	"github.com/imroc/req/v3"
	Utils "github.com/jopemachine/mac-sync-config/src/utils"
)

func GetMacSyncConfigs() string {
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
