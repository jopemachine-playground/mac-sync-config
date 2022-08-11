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

func CompressConfigs(targetFilePath string, dstFilePath string) {
	cpCmd := exec.Command("cp", "-pR", targetFilePath, dstFilePath)
	output, err := cpCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	hashValue := filepath.Base(dstFilePath)
	tarArgs := strings.Fields(fmt.Sprintf("tar -cjf %s.tar %s", dstFilePath, hashValue))
	tarCmd := exec.Command(tarArgs[0], tarArgs[1:]...)
	tarCmd.Dir = filepath.Dir(dstFilePath)
	output, err = tarCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	bzipArgs := strings.Fields(fmt.Sprintf("bzip2 %s.tar", dstFilePath))
	bzipCmd := exec.Command(bzipArgs[0], bzipArgs[1:]...)
	output, err = bzipCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)
}

func DecompressConfigs(filepath string) string {
	bunzipArgs := strings.Fields(fmt.Sprintf("bunzip2 %s", filepath))
	bunzipCmd := exec.Command(bunzipArgs[0], bunzipArgs[1:]...)
	output, err := bunzipCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	tarFilePath := strings.Split(filepath, ".bz2")[0]
	configsDirPath := strings.Split(tarFilePath, ".tar")[0]

	if err = os.Mkdir(configsDirPath, os.ModePerm); err != nil {
		panic(err)
	}

	tarArgs := strings.Fields(fmt.Sprintf("tar -xvf %s -C %s", tarFilePath, configsDirPath))
	tarCmd := exec.Command(tarArgs[0], tarArgs[1:]...)
	output, err = tarCmd.CombinedOutput()
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

func DownloadRemoteConfigs() error {
	remoteCommitHashId := FetchRemoteConfigCommitHashId()
	configFileLastChanged := ReadConfigFileLastChanged()

	if configFileLastChanged["remote-commit-hash-id"] == remoteCommitHashId {
		Logger.Info("Config files already up to dated.")
		return nil
	}

	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadConfig(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))

	PanicIfErr(err)

	configPathsToSync := configs.ConfigPathsToSync

	for _, configPathToSync := range configPathsToSync {
		hash := GetConfigHash(configPathToSync)

		configDirPath := fmt.Sprintf("%s/%s/%s", tempPath, GetRemoteConfigFolderName(), hash)
		configZipFilePath := fmt.Sprintf("%s.tar.bz2", configDirPath)

		if _, err := os.Stat(configZipFilePath); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"%s\", but the config file not found. Upload the config file before download", configPathToSync, MacSyncConfigsFile))
			continue
		}

		DecompressConfigs(configZipFilePath)
		srcPath := fmt.Sprintf("%s/%s", configDirPath, hash)
		dstPath := HandleWhiteSpaceInPath(HandleTildePath(configPathToSync))
		dirPath := filepath.Dir(dstPath)

		mkdirArgs := strings.Fields(fmt.Sprintf("mkdir -p %s", dirPath))
		mkdirCmd := exec.Command(mkdirArgs[0], mkdirArgs[1:]...)
		_, err := mkdirCmd.CombinedOutput()
		PanicIfErr(err)

		os.Rename(srcPath, dstPath)
	}

	if _, err := os.Stat(tempPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempPath, os.ModePerm)
	}

	configFileLastChanged["remote-commit-hash-id"] = remoteCommitHashId
	WriteConfigFileLastChanged(configFileLastChanged)

	Logger.Success("Local config files are updated. Some changes might requires reboot to apply.")
	return nil
}

func UploadConfigFiles() {
	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadConfig(fmt.Sprintf("%s/%s", tempPath, MacSyncConfigsFile))
	PanicIfErr(err)

	for _, configPathToSync := range configs.ConfigPathsToSync {
		hashId := GetConfigHash(configPathToSync)
		dstFilePath := fmt.Sprintf("%s/%s/%s.tar.bz2", tempPath, GetRemoteConfigFolderName(), hashId)
		dstFilePathWithoutExt := strings.Split(dstFilePath, ".tar")[0]

		// Update files if already exist
		if _, err := os.Stat(dstFilePath); !errors.Is(err, os.ErrNotExist) {
			err := os.Remove(dstFilePath)
			PanicIfErr(err)
		}

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

	gitCommitArgs := strings.Fields("git commit -m 🔧 -m updated_by_mac-sync")
	gitCommitCmd := exec.Command(gitCommitArgs[0], gitCommitArgs[1:]...)
	gitCommitCmd.Dir = tempPath

	output, err = gitCommitCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	gitPushArgs := strings.Fields("git push -u origin main --force")
	gitPushCmd := exec.Command(gitPushArgs[0], gitPushArgs[1:]...)
	gitPushCmd.Dir = tempPath
	output, err = gitPushCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	Logger.Info("🔧 Config files updated successfully")
	os.RemoveAll(tempPath)
}

func FetchRemoteConfigCommitHashId() string {
	args := strings.Fields(fmt.Sprintf("git ls-remote https://github.com/%s/%s HEAD", PreferenceSingleton.GithubId, PreferenceSingleton.MacSyncConfigGitRepositoryName))
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.CombinedOutput()
	PanicIfErr(err)

	return strings.TrimSpace(strings.Split(fmt.Sprintf("%s", stdout), "HEAD")[0])
}
