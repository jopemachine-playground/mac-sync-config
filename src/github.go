package src

import (
	"fmt"

	"github.com/imroc/req/v3"
)

func CreateMacSyncConfigRequest(fileName string) (*req.Response, error) {
	return req.C().R().
		SetHeader("Authorization", fmt.Sprintf("token %s", PreferenceSingleton.GithubToken)).
		SetHeader("Cache-control", "no-cache").
		SetPathParam("userName", PreferenceSingleton.GithubId).
		SetPathParam("repoName", PreferenceSingleton.MacSyncConfigGitRepositoryName).
		SetPathParam("branchName", "main").
		SetPathParam("fileName", fileName).
		EnableDump().
		Get("https://raw.githubusercontent.com/{userName}/{repoName}/{branchName}/{fileName}")
}
