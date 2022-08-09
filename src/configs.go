package src

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const CachePath = "~/Library/Caches/Mac-sync"
const PreferencePath = "~/Library/Preferences/Mac-sync"

var (
	ProgramCachePath               = strings.Join([]string{CachePath, "local-programs.yaml"}, "/")
	ConfigFileLastChangedCachePath = strings.Join([]string{CachePath, "last-changed.json"}, "/")
	PreferenceFilePath             = strings.Join([]string{PreferencePath, "preference.json"}, "/")
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
	GithubToken                    string `json:"github_token"`
	UserPassword                   string `json:"user_password"`
	MacSyncConfigGitRepositoryName string `json:"mac_sync_config_git_repository_name"`
}

var (
	PreferenceSingleton = ReadPreference()
)

func scanPreference(config *Preference) {
	Logger.Question("Enter your Github token:")
	ghToken := bufio.NewScanner(os.Stdin)
	ghToken.Scan()
	config.GithubToken = ghToken.Text()

	Logger.Question("Enter your User account's password:")
	password := bufio.NewScanner(os.Stdin)
	password.Scan()
	config.UserPassword = password.Text()

	Logger.Question("Enter a Git repository name for storing mac-sync's configuration files:")
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
		if err != nil {
			panic(err)
		}

		os.WriteFile(preferenceFilePath, bytesToWrite, os.ModePerm)
		Logger.Success(fmt.Sprintf("Preference file is saved successfully on the '%s'", preferenceFilePath))
	} else {
		dat, err := ioutil.ReadFile(preferenceFilePath)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(dat, &config); err != nil {
			panic(err)
		}
	}

	return config
}

func ReadLocalProgramCache() map[string]PackageManagerInfo {
	programCachePath := HandleTildePath(ProgramCachePath)

	if _, err := os.Stat(programCachePath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	dat, err := ioutil.ReadFile(programCachePath)
	if err != nil {
		panic(err)
	}

	var config map[string]PackageManagerInfo

	if err := yaml.Unmarshal(dat, &config); err != nil {
		panic(err)
	}

	return config
}

func ReadConfigFileLastChanged() map[string]string {
	configFileLastChangedCachePath := HandleTildePath(ConfigFileLastChangedCachePath)

	if _, err := os.Stat(configFileLastChangedCachePath); errors.Is(err, os.ErrNotExist) {
		return make(map[string]string)
	}

	dat, err := ioutil.ReadFile(configFileLastChangedCachePath)
	if err != nil {
		panic(err)
	}

	var lastChangedMap map[string]string

	if err := json.Unmarshal(dat, &lastChangedMap); err != nil {
		panic(err)
	}

	return lastChangedMap
}

func WriteConfigFileLastChanged(lastChanged map[string]string) {
	cacheDir := HandleTildePath(CachePath)
	configFileLastChangedCachePath := HandleTildePath(ConfigFileLastChangedCachePath)

	if _, err := os.Stat(cacheDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(cacheDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	} else if _, err := os.Stat(configFileLastChangedCachePath); !errors.Is(err, os.ErrNotExist) {
		os.Remove(configFileLastChangedCachePath)
	}

	bytesToWrite, err := json.Marshal(lastChanged)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(configFileLastChangedCachePath, bytesToWrite, os.ModePerm); err != nil {
		panic(err)
	}
}

func WriteLocalProgramCache(cache map[string]PackageManagerInfo) {
	cacheDir := HandleTildePath(CachePath)
	programCachePath := HandleTildePath(ProgramCachePath)

	if _, err := os.Stat(cacheDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(cacheDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	} else if _, err := os.Stat(programCachePath); !errors.Is(err, os.ErrNotExist) {
		os.Remove(programCachePath)
	}

	bytesToWrite, err := yaml.Marshal(cache)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(programCachePath, bytesToWrite, os.ModePerm); err != nil {
		panic(err)
	}
}

func ClearCache() {
	programCachePath := HandleTildePath(ProgramCachePath)
	configFileLastChangedCachePath := HandleTildePath(ConfigFileLastChangedCachePath)
	os.Remove(programCachePath)
	os.Remove(configFileLastChangedCachePath)
	Logger.Success("Cache file cleared")
}
