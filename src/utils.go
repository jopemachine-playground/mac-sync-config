package src

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/imroc/req/v3"
	"gopkg.in/yaml.v3"
)

func GetRemoteConfigFolderName() string {
	return ".mac-sync-configs"
}

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

func FetchRemoteProgramInfo() map[string]PackageManagerInfo {
	resp, err := CreateMacSyncConfigRequest(MacSyncProgramsFile)

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
func Remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

func GetConfigHash(text string) string {
	algorithm := sha256.New()
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}

func HandleTildePath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}

	return path
}

func HandleWhiteSpaceInPath(path string) string {
	return strings.ReplaceAll(path, " ", "\\ ")
}

func IsRootUser() bool {
	return os.Geteuid() == 0
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIfErrWithOutput(output string, err error) {
	if err != nil {
		panic(output)
	}
}
