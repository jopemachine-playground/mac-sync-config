package src

import (
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

	var selectedFilePaths = []string{}
	var updatedFilePaths = []string{}

	for _, configPathToSync := range configs.ConfigPathsToSync {
		configRootPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())
		absConfigPathToSync := HandleTildePath(configPathToSync)
		dstFilePath := fmt.Sprintf("%s%s", configRootPath, absConfigPathToSync)

		// Delete files for update if the files already exist
		if _, err := os.Stat(dstFilePath); !errors.Is(err, os.ErrNotExist) {
			err := os.RemoveAll(dstFilePath)
			PanicIfErr(err)
		}

		if _, err := os.Stat(absConfigPathToSync); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" not found in the local computer.", configPathToSync))
			continue
		}

		CopyConfigs(absConfigPathToSync, dstFilePath)
		PanicIfErr(err)
		updatedFilePaths = append(updatedFilePaths, dstFilePath)
	}

	if Flag_OverWrite {
		GitAddCwd(tempPath)
		selectedFilePaths = updatedFilePaths
	} else {
		for _, updatedFilePath := range updatedFilePaths {
			if haveDiff := IsUpdated(tempPath, updatedFilePath); haveDiff {
				ShowDiff(tempPath, updatedFilePath)

				Logger.NewLine()
				if yes := EnterYesNoQuestion("Update? (Y/N)"); yes {
					GitAddFile(tempPath, updatedFilePath)
					selectedFilePaths = append(selectedFilePaths, updatedFilePath)
				}
			}

			Logger.NewLine()
		}

		for _, selectedFilePath := range selectedFilePaths {
			Logger.Success(fmt.Sprintf("\"%s\" updated.", selectedFilePath))
		}
	}

	GitCommit(tempPath)
	GitPush(tempPath)

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
