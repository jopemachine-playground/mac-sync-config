package src

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func CopyConfigs(targetFilePath string, dstFilePath string) {
	dstFilePath = (dstFilePath)
	dirPath := filepath.Dir(dstFilePath)

	mkdirCmd := exec.Command("mkdir", "-p", dirPath)
	output, err := mkdirCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	cpCmd := exec.Command("cp", "-fR", targetFilePath, dstFilePath)
	output, err = cpCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)
}

func ReadConfig(filepath string) (ConfigInfo, error) {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		return ConfigInfo{}, err
	}

	dat, err := ioutil.ReadFile(filepath)

	PanicIfErr(err)

	var config ConfigInfo

	err = yaml.Unmarshal(dat, &config)
	PanicIfErr(err)

	return config, nil
}

func CloneMacSyncConfigRepository() string {
	tempPath, err := os.MkdirTemp("", "mac-sync-config-temp-")
	PanicIfErr(err)

	// Should fully clone repository for commit and push
	args := strings.Fields(fmt.Sprintf("git clone https://github.com/%s/%s %s", PreferenceSingleton.GithubId, PreferenceSingleton.MacSyncConfigGitRepositoryName, tempPath))
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	tempConfigDirPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

	if _, err := os.Stat(tempConfigDirPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempConfigDirPath, os.ModePerm)
	}

	return tempPath
}

func PullRemoteConfigs(argFilter string) {
	remoteCommitHashId := FetchRemoteConfigCommitHashId()
	configFileLastChanged := ReadConfigFileLastChanged()

	if configFileLastChanged["remote-commit-hash-id"] == remoteCommitHashId {
		Logger.Info("Config files already up to date.")
		return
	}

	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadConfig(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))

	PanicIfErr(err)

	configPathsToSync := configs.ConfigPathsToSync

	for _, configPathToSync := range configPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

		if argFilter != "" && strings.Contains(filepath.Base(configPathToSync), argFilter) == false {
			continue
		}

		absConfigPathToSync := HandleTildePath(configPathToSync)
		srcFilePath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)

		if _, err := os.Stat(srcFilePath); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"%s\", but the config file not found. Please push the config file before pulling.", configPathToSync, MacSyncConfigsFile))
			continue
		}

		dstPath := HandleTildePath(configPathToSync)

		if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
			err = os.RemoveAll(dstPath)
			PanicIfErr(err)
		}

		CopyConfigs(srcFilePath, dstPath)
		Logger.Success(fmt.Sprintf("\"%s\" updated.", configPathToSync))
	}

	if _, err := os.Stat(tempPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempPath, os.ModePerm)
	}

	if argFilter == "" {
		configFileLastChanged["remote-commit-hash-id"] = remoteCommitHashId
		WriteConfigFileLastChanged(configFileLastChanged)
		Logger.Info("Local config files are updated. Some changes might require to reboot to apply.")
	}
}

func PushConfigFiles() {
	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadConfig(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))
	PanicIfErr(err)

	var commitMsgBuffer bytes.Buffer
	commitMsgBuffer.WriteString("-m")

	for _, configPathToSync := range configs.ConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())
		absConfigPathToSync := HandleTildePath(configPathToSync)
		dstFilePath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)

		// Delete files for update if the files already exist
		if _, err := os.Stat(dstFilePath); !errors.Is(err, os.ErrNotExist) {
			err := os.Remove(dstFilePath)
			PanicIfErr(err)
		}

		commitMsgBuffer.WriteString(fmt.Sprintf("%s\n", HandleWhiteSpaceInPath(configPathToSync)))

		if _, err := os.Stat(absConfigPathToSync); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" not found in the local computer.", configPathToSync))
			continue
		}

		CopyConfigs(absConfigPathToSync, dstFilePath)
		PanicIfErr(err)

		Logger.Success(fmt.Sprintf("\"%s\" updated.", configPathToSync))
	}

	gitAddArgs := strings.Fields(fmt.Sprintf("git add %s", tempPath))
	gitAddCmd := exec.Command(gitAddArgs[0], gitAddArgs[1:]...)
	gitAddCmd.Dir = tempPath
	output, err := gitAddCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	gitCommitCmd := exec.Command("git", "commit", "--author", "github-actions[bot] <41898282+github-actions[bot]@users.noreply.github.com>", "--allow-empty", "-m", "🔧", commitMsgBuffer.String())
	gitCommitCmd.Dir = tempPath

	output, err = gitCommitCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	gitPushArgs := strings.Fields("git push -u origin main --force")
	gitPushCmd := exec.Command(gitPushArgs[0], gitPushArgs[1:]...)
	gitPushCmd.Dir = tempPath
	gitPushCmd.Stdout = os.Stdout
	gitPushCmd.Stderr = os.Stderr
	gitPushCmd.Run()

	Logger.Info("Config files updated successfully.")
	os.RemoveAll(tempPath)
}

func FetchRemoteConfigCommitHashId() string {
	args := strings.Fields(fmt.Sprintf("git ls-remote https://github.com/%s/%s HEAD", PreferenceSingleton.GithubId, PreferenceSingleton.MacSyncConfigGitRepositoryName))
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.CombinedOutput()
	PanicIfErr(err)

	return strings.TrimSpace(strings.Split(fmt.Sprintf("%s", stdout), "HEAD")[0])
}
