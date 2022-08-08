package src

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

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

	args := strings.Fields(fmt.Sprintf("git clone https://github.com/%s/mac-sync-configs %s", GetUserName(), tempPath))
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	tempConfigDirPath := fmt.Sprintf("%s/%s", tempPath, "configs")

	if _, err := os.Stat(tempConfigDirPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempConfigDirPath, 0777)
	}

	Logger.Log(string(stdout))
	return tempPath
}

func DownloadConfigs() error {
	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadConfig(fmt.Sprintf("%s/configs.yaml", tempPath))

	if err != nil {
		panic(err)
	}

	configPathsToSync := configs.ConfigPathsToSync

	for _, configPathToSync := range configPathsToSync {
		hash := GetConfigHash(configPathToSync)

		configDirPath := fmt.Sprintf("%s/configs/%s", tempPath, hash)
		configZipFilePath := fmt.Sprintf("%s.tar.bz2", configDirPath)

		if _, err := os.Stat(configZipFilePath); errors.Is(err, os.ErrNotExist) {
			Logger.Warning(fmt.Sprintf("\"%s\" is specified on your \"configs.yaml\", but the config file not found. Upload the config file before download", configPathToSync))
			continue
		}

		DecompressConfigs(configZipFilePath)
		os.Rename(fmt.Sprintf("%s/%s", configDirPath, hash), HandleTildePath(configPathToSync))
	}

	if _, err := os.Stat(tempPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempPath, 0777)
	}

	Logger.Success("Configs up to dated")
	return nil
}

func UploadConfigs() {
	tempPath := CloneMacSyncConfigRepository()
	configs, err := ReadConfig(fmt.Sprintf("%s/configs.yaml", tempPath))

	if err != nil {
		panic(err)
	}

	for _, configPathToSync := range configs.ConfigPathsToSync {
		hashId := GetConfigHash(configPathToSync)
		dstFilePath := fmt.Sprintf("%s/configs/%s.tar.bz2", tempPath, hashId)
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

		Logger.Success(fmt.Sprintf("\"%s\" file added.", configPathToSync))
	}

	gitAddArgs := strings.Fields(fmt.Sprintf("git add %s", tempPath))
	gitAddCmd := exec.Command(gitAddArgs[0], gitAddArgs[1:]...)
	gitAddCmd.Dir = tempPath
	_, err = gitAddCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	gitCommitArgs := strings.Fields("git commit -m \"Config_files_updated\"")
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

	Logger.Success("🔧 Config files updated")

	os.RemoveAll(tempPath)
}
