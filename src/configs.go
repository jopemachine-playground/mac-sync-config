package src

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	CachePath      = "~/Library/Caches/Mac-sync"
	PreferencePath = "~/Library/Preferences/Mac-sync"
)

const (
	MacSyncConfigsFile = "mac-sync-configs.yaml"
)

var (
	ConfigFileLastChangedCachePath = strings.Join([]string{CachePath, "last-changed.json"}, "/")
	PreferenceFilePath             = strings.Join([]string{PreferencePath, "preference.json"}, "/")
)

var (
	PreferenceSingleton = ReadPreference()
)

type PackageManagerInfo struct {
	InstallCommand   string   `yaml:"install"`
	UninstallCommand string   `yaml:"uninstall"`
	Programs         []string `yaml:"programs"`
}

type ConfigInfo struct {
	ConfigPathsToSync []string `yaml:"sync"`
}

type Preference struct {
	GithubId                       string `json:"github_id"`
	GithubToken                    string `json:"github_token"`
	MacSyncConfigGitRepositoryName string `json:"mac_sync_config_git_repository_name"`
}

func scanPreference(config *Preference) {
	Logger.Question("Enter your Github id:")
	ghId := bufio.NewScanner(os.Stdin)
	ghId.Scan()
	config.GithubId = ghId.Text()

	Logger.Question("Enter your Github access token:")
	ghToken := bufio.NewScanner(os.Stdin)
	ghToken.Scan()
	config.GithubToken = ghToken.Text()

	Logger.Question("Enter a Git repository name for saving mac-sync's configuration files:")
	repoName := bufio.NewScanner(os.Stdin)
	repoName.Scan()
	config.MacSyncConfigGitRepositoryName = repoName.Text()
}

func ReadPreference() Preference {
	preferenceDirPath := HandleTildePath(PreferencePath)
	preferenceFilePath := HandleTildePath(PreferenceFilePath)

	var config Preference

	// If not exist, create new preference config file
	if _, err := os.Stat(preferenceFilePath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(preferenceDirPath, os.ModePerm)
		if err != nil && !errors.Is(err, os.ErrExist) {
			panic(err)
		}

		scanPreference(&config)

		bytesToWrite, err := json.Marshal(config)
		PanicIfErr(err)

		os.WriteFile(preferenceFilePath, bytesToWrite, os.ModePerm)
		Logger.Success(fmt.Sprintf("mac-sync's preference file is saved successfully on the '%s'.", preferenceFilePath))
	} else {
		dat, err := ioutil.ReadFile(preferenceFilePath)
		PanicIfErr(err)

		err = json.Unmarshal(dat, &config)
		PanicIfErr(err)
	}

	return config
}

func ReadConfigFileLastChanged() map[string]string {
	configFileLastChangedCachePath := HandleTildePath(ConfigFileLastChangedCachePath)

	if _, err := os.Stat(configFileLastChangedCachePath); errors.Is(err, os.ErrNotExist) {
		return make(map[string]string)
	}

	dat, err := ioutil.ReadFile(configFileLastChangedCachePath)
	PanicIfErr(err)

	var lastChangedMap map[string]string

	err = json.Unmarshal(dat, &lastChangedMap)
	PanicIfErr(err)

	return lastChangedMap
}

func WriteConfigFileLastChanged(lastChanged map[string]string) {
	cacheDir := HandleTildePath(CachePath)
	configFileLastChangedCachePath := HandleTildePath(ConfigFileLastChangedCachePath)

	if _, err := os.Stat(cacheDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(cacheDir, os.ModePerm)
		PanicIfErr(err)
	} else if _, err := os.Stat(configFileLastChangedCachePath); !errors.Is(err, os.ErrNotExist) {
		os.Remove(configFileLastChangedCachePath)
	}

	bytesToWrite, err := json.Marshal(lastChanged)
	PanicIfErr(err)

	err = ioutil.WriteFile(configFileLastChangedCachePath, bytesToWrite, os.ModePerm)
	PanicIfErr(err)
}

func ClearCache() {
	configFileLastChangedCachePath := HandleTildePath(ConfigFileLastChangedCachePath)

	err := os.Remove(configFileLastChangedCachePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
	Logger.Success("Cache file cleared.")
}
