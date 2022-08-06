package src

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func GetUserName() string {
	args := strings.Fields("git config --global user.email")
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.Split(fmt.Sprintf("%s", stdout), "@")[0]
}

func GetGithubToken() string {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	return os.Getenv("gh_token")
}

func createMapSyncConfigRequest(fileName string) (*req.Response, error) {
	return req.C().R().
		SetHeader("Authorization", fmt.Sprintf("token %s", GetGithubToken())).
		SetPathParam("userName", GetUserName()).
		SetPathParam("repoName", "mac-sync-configs").
		SetPathParam("branchName", "main").
		SetPathParam("fileName", fileName).
		EnableDump().
		Get("https://raw.githubusercontent.com/{userName}/{repoName}/{branchName}/{fileName}")
}

func GetDependencies() map[string]PackageManagerInfo {
	resp, err := createMapSyncConfigRequest("dependency.yaml")

	if err != nil {
		panic(err)
	}

	if resp.IsSuccess() {
		var result map[string]PackageManagerInfo
		if err := yaml.Unmarshal(resp.Bytes(), &result); err != nil {
			panic(err)
		}

		return result
	}

	panic(resp.Dump())
}

func GetConfigs() ConfigInfo {
	resp, err := createMapSyncConfigRequest("configs.yaml")

	if err != nil {
		panic(err)
	}

	if resp.IsSuccess() {
		var result ConfigInfo
		if err := yaml.Unmarshal(resp.Bytes(), &result); err != nil {
			panic(err)
		}

		return result
	}

	panic(resp.Dump())
}
