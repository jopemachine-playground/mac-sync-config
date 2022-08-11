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

func CompressConfigs(targetFilePath string, dstFilePath string) {
	cpCmd := exec.Command("cp", "-R", targetFilePath, dstFilePath)
	output, err := cpCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	hashValue := filepath.Base(dstFilePath)
	// c: create archive
	// j: compress by bzip2
	// f: specify file name
	tarArgs := strings.Fields(fmt.Sprintf("tar -cjf %s.tar %s", dstFilePath, hashValue))
	tarCmd := exec.Command(tarArgs[0], tarArgs[1:]...)
	tarCmd.Dir = filepath.Dir(dstFilePath)
	output, err = tarCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)
}

func DecompressConfigs(filepath string) string {
	configsDirPath := strings.Split(filepath, ".tar")[0]

	if err := os.Mkdir(configsDirPath, os.ModePerm); err != nil {
		panic(err)
	}

	// x: decompress archive
	// f: specify file name
	// C: specify target directory
	tarCmd := exec.Command("tar", "-xf", filepath, "-C", configsDirPath)
	output, err := tarCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	return configsDirPath
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

func DownloadRemoteConfigs() {
	remoteCommitHashId := FetchRemoteConfigCommitHashId()
	configFileLastChanged := ReadConfigFileLastChanged()

	if configFileLastChanged["remote-commit-hash-id"] == remoteCommitHashId {
		Logger.Info("Config files already up to dated.")
		return
	}

	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadConfig(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))

	PanicIfErr(err)

	configPathsToSync := configs.ConfigPathsToSync

	for _, configPathToSync := range configPathsToSync {
		hash := GetConfigHash(configPathToSync)

		configDirPath := fmt.Sprintf("%s/%s/%s", tempPath, GetRemoteConfigFolderName(), hash)
		configZipFilePath := fmt.Sprintf("%s.tar", configDirPath)

		if _, err := os.Stat(configZipFilePath); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"%s\", but the config file not found. Upload the config file before download", configPathToSync, MacSyncConfigsFile))
			continue
		}

		DecompressConfigs(configZipFilePath)
		srcPath := fmt.Sprintf("%s/%s", configDirPath, hash)
		dstPath := HandleTildePath(configPathToSync)
		dirPath := filepath.Dir(dstPath)

		mkdirCmd := exec.Command("mkdir", "-p", HandleWhiteSpaceInPath(dirPath))
		output, err := mkdirCmd.CombinedOutput()
		PanicIfErrWithOutput(string(output), err)

		if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
			err = os.RemoveAll(dstPath)
			PanicIfErr(err)
		}
		
		err = os.Rename(srcPath, dstPath)
		PanicIfErr(err)
	}

	if _, err := os.Stat(tempPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempPath, os.ModePerm)
	}

	configFileLastChanged["remote-commit-hash-id"] = remoteCommitHashId
	WriteConfigFileLastChanged(configFileLastChanged)

	Logger.Success("Local config files are updated. Some changes might requires reboot to apply.")
}

func UploadConfigFiles() {
	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadConfig(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))
	PanicIfErr(err)

	var commitMsgBuffer bytes.Buffer
	commitMsgBuffer.WriteString("-m")

	for _, configPathToSync := range configs.ConfigPathsToSync {
		hashId := GetConfigHash(configPathToSync)
		dstFilePath := fmt.Sprintf("%s/%s/%s.tar", tempPath, GetRemoteConfigFolderName(), hashId)
		dstFilePathWithoutExt := strings.Split(dstFilePath, ".tar")[0]

		// Update files if already exist
		if _, err := os.Stat(dstFilePath); !errors.Is(err, os.ErrNotExist) {
			err := os.Remove(dstFilePath)
			PanicIfErr(err)
		}

		commitMsgBuffer.WriteString(fmt.Sprintf("%s\n", HandleWhiteSpaceInPath(configPathToSync)))
		absConfigPathToSync := HandleTildePath(configPathToSync)

		if _, err := os.Stat(absConfigPathToSync); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" file not found in the local", configPathToSync))
			continue
		}

		CompressConfigs(absConfigPathToSync, dstFilePathWithoutExt)
		err := os.RemoveAll(dstFilePathWithoutExt)
		PanicIfErr(err)

		Logger.Success(fmt.Sprintf("\"%s\" file updated.", configPathToSync))
	}

	gitAddArgs := strings.Fields(fmt.Sprintf("git add %s", tempPath))
	gitAddCmd := exec.Command(gitAddArgs[0], gitAddArgs[1:]...)
	gitAddCmd.Dir = tempPath
	output, err := gitAddCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	gitCommitCmd := exec.Command("git", "commit", "-m", "🔧", commitMsgBuffer.String())
	gitCommitCmd.Dir = tempPath

	output, err = gitCommitCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	gitPushArgs := strings.Fields("git push -u origin main --force")
	gitPushCmd := exec.Command(gitPushArgs[0], gitPushArgs[1:]...)
	gitPushCmd.Dir = tempPath
	output, err = gitPushCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	Logger.Info("Config files updated successfully")
	os.RemoveAll(tempPath)
}

func FetchRemoteConfigCommitHashId() string {
	args := strings.Fields(fmt.Sprintf("git ls-remote https://github.com/%s/%s HEAD", PreferenceSingleton.GithubId, PreferenceSingleton.MacSyncConfigGitRepositoryName))
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.CombinedOutput()
	PanicIfErr(err)

	return strings.TrimSpace(strings.Split(fmt.Sprintf("%s", stdout), "HEAD")[0])
}
