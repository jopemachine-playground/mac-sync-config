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
	cpArgs := strings.Fields(fmt.Sprintf("cp -pR %s %s", targetFilePath, dstFilePath))
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

	if err = os.Mkdir(configsDirPath, os.ModePerm); err != nil {
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

func ReadConfig(filepath string) (ConfigInfo, error) {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		return ConfigInfo{}, err
	}

	dat, err := ioutil.ReadFile(filepath)

	if err != nil {
		panic(err)
	}

	var config ConfigInfo

	if err := yaml.Unmarshal(dat, &config); err != nil {
		panic(err)
	}

	return config, nil
}

func CloneMacSyncConfigRepository() string {
	tempPath, err := os.MkdirTemp("", "mac-sync-config-temp-")
	if err != nil {
		panic(err)
	}

	// Should fully clone repository for commit and push
	args := strings.Fields(fmt.Sprintf("git clone https://github.com/%s/%s %s", GetGitUserId(), GetMacSyncConfigRepositoryName(), tempPath))
	cmd := exec.Command(args[0], args[1:]...)
	_, err = cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	tempConfigDirPath := fmt.Sprintf("%s/%s", tempPath, "mac-sync-configs")

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
	configs, err := ReadConfig(fmt.Sprintf("%s/mac-sync-configs.yaml", tempPath))

	if err != nil {
		panic(err)
	}

	configPathsToSync := configs.ConfigPathsToSync

	for _, configPathToSync := range configPathsToSync {
		hash := GetConfigHash(configPathToSync)

		configDirPath := fmt.Sprintf("%s/mac-sync-configs/%s", tempPath, hash)
		configZipFilePath := fmt.Sprintf("%s.tar.bz2", configDirPath)

		if _, err := os.Stat(configZipFilePath); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"mac-sync-configs.yaml\", but the config file not found. Upload the config file before download", configPathToSync))
			continue
		}

		DecompressConfigs(configZipFilePath)
		os.Rename(fmt.Sprintf("%s/%s", configDirPath, hash), HandleTildePath(configPathToSync))
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
	configs, err := ReadConfig(fmt.Sprintf("%s/mac-sync-configs.yaml", tempPath))

	if err != nil {
		panic(err)
	}

	for _, configPathToSync := range configs.ConfigPathsToSync {
		hashId := GetConfigHash(configPathToSync)
		dstFilePath := fmt.Sprintf("%s/mac-sync-configs/%s.tar.bz2", tempPath, hashId)
		dstFilePathWithoutExt := strings.Split(dstFilePath, ".tar")[0]

		// Update files if already exist
		if _, err := os.Stat(dstFilePath); !errors.Is(err, os.ErrNotExist) {
			err := os.Remove(dstFilePath)
			if err != nil {
				panic(err)
			}
		}

		if _, err := os.Stat(HandleTildePath(configPathToSync)); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" file not found in the local", configPathToSync))
			continue
		}

		CompressConfigs(HandleTildePath(configPathToSync), dstFilePathWithoutExt)
		if err := os.RemoveAll(dstFilePathWithoutExt); err != nil {
			panic(err)
		}

		Logger.Success(fmt.Sprintf("\"%s\" file updated.", configPathToSync))
	}

	gitAddArgs := strings.Fields(fmt.Sprintf("git add %s", tempPath))
	gitAddCmd := exec.Command(gitAddArgs[0], gitAddArgs[1:]...)
	gitAddCmd.Dir = tempPath
	_, err = gitAddCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	gitCommitArgs := strings.Fields("git commit -m 🔧 -m updated_by_mac-sync")
	gitCommitCmd := exec.Command(gitCommitArgs[0], gitCommitArgs[1:]...)
	gitCommitCmd.Dir = tempPath

	_, err = gitCommitCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	gitPushArgs := strings.Fields("git push -u origin main --force")
	gitPushCmd := exec.Command(gitPushArgs[0], gitPushArgs[1:]...)
	gitPushCmd.Dir = tempPath
	_, err = gitPushCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	Logger.Info("🔧 Config files updated successfully")
	os.RemoveAll(tempPath)
}

func FetchRemoteConfigCommitHashId() string {
	args := strings.Fields(fmt.Sprintf("git ls-remote https://github.com/%s/%s HEAD", GetGitUserId(), GetMacSyncConfigRepositoryName()))
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(strings.Split(fmt.Sprintf("%s", stdout), "HEAD")[0])
}
