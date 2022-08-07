package src

import (
	"bytes"
	"encoding/gob"
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

func FetchDependencies() map[string]PackageManagerInfo {
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

func FetchConfigs() ConfigInfo {
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

// TODO: Replace below function with stdlib's one when it is merged
// Ref: https://stackoverflow.com/questions/10485743/contains-method-for-a-slice
func StringContains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

// Removes slice element at index(s) and returns new slice
func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
