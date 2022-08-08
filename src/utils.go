package src

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/imroc/req/v3"
	"gopkg.in/yaml.v3"
)

func GetGitUserId() string {
	args := strings.Fields("git config --global user.email")
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.Split(fmt.Sprintf("%s", stdout), "@")[0]
}

func CreateMacSyncConfigRequest(fileName string) (*req.Response, error) {
	return req.C().R().
		SetHeader("Authorization", fmt.Sprintf("token %s", PreferenceSingleton.GithubToken)).
		SetHeader("Cache-control", "no-cache").
		SetPathParam("userName", GetGitUserId()).
		SetPathParam("repoName", "mac-sync-configs").
		SetPathParam("branchName", "main").
		SetPathParam("fileName", fileName).
		EnableDump().
		Get("https://raw.githubusercontent.com/{userName}/{repoName}/{branchName}/{fileName}")
}

func FetchRemoveProgramInfo() map[string]PackageManagerInfo {
	resp, err := CreateMacSyncConfigRequest("programs.yaml")

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

func CompressConfigs(targetFilePath string, dstFilePath string) {
	cpArgs := strings.Fields(fmt.Sprintf("cp -R %s %s", targetFilePath, dstFilePath))
	cpCmd := exec.Command(cpArgs[0], cpArgs[1:]...)
	_, err := cpCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	hashValue := filepath.Base(dstFilePath)
	tarArgs := strings.Fields(fmt.Sprintf("tar -cjf %s.tar %s", dstFilePath, hashValue))
	tarCmd := exec.Command(tarArgs[0], tarArgs[1:]...)
	tarCmd.Dir = filepath.Dir(dstFilePath)
	_, err = tarCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	bzipArgs := strings.Fields(fmt.Sprintf("bzip2 %s.tar", dstFilePath))
	bzipCmd := exec.Command(bzipArgs[0], bzipArgs[1:]...)
	_, err = bzipCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
}

func DecompressConfigs(filepath string) string {
	bunzipArgs := strings.Fields(fmt.Sprintf("bunzip2 %s", filepath))
	bunzipCmd := exec.Command(bunzipArgs[0], bunzipArgs[1:]...)
	_, err := bunzipCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	tarFilePath := strings.Split(filepath, ".bz2")[0]
	configsDirPath := strings.Split(tarFilePath, ".tar")[0]

	if err = os.Mkdir(configsDirPath, 0777); err != nil {
		panic(err)
	}

	tarArgs := strings.Fields(fmt.Sprintf("tar -xvf %s -C %s", tarFilePath, configsDirPath))
	tarCmd := exec.Command(tarArgs[0], tarArgs[1:]...)
	_, err = tarCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	return configsDirPath
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

func IsRootUser() bool {
	return os.Geteuid() == 0
}
