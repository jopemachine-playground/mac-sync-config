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
	ProgramCachePath   = strings.Join([]string{CachePath, "local-programs.yaml"}, "/")
	PreferenceFilePath = strings.Join([]string{PreferencePath, "configs.json"}, "/")
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
	GithubToken  string `json:"github_token"`
	UserPassword string `json:"user_password"`
}

var (
	PreferenceSingleton = ReadPreference()
)

func ReadPreference() Preference {
	preferenceDirPath := HandleTildePath(PreferencePath)
	preferenceFilePath := HandleTildePath(PreferenceFilePath)

	var config Preference

	// If not exist, create new preference config file
	if _, err := os.Stat(preferenceFilePath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(preferenceDirPath, 0777)
		if err != nil && !errors.Is(err, os.ErrExist) {
			panic(err)
		}

		Logger.Info("Enter your Github token:")
		ghToken := bufio.NewScanner(os.Stdin)
		ghToken.Scan()
		config.GithubToken = ghToken.Text()

		Logger.Info("Enter your User account's password:")
		password := bufio.NewScanner(os.Stdin)
		password.Scan()
		config.UserPassword = password.Text()

		bytesToWrite, err := json.Marshal(config)
		if err != nil {
			panic(err)
		}

		os.WriteFile(preferenceFilePath, bytesToWrite, 0777)
		Logger.Success(fmt.Sprintf("Preference file saved successfully on the '%s'", preferenceFilePath))
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

func WriteLocalProgramCache(cache map[string]PackageManagerInfo) {
	cacheDir := HandleTildePath(CachePath)
	programCachePath := HandleTildePath(ProgramCachePath)

	if _, err := os.Stat(cacheDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(cacheDir, 0777)
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

	if err := ioutil.WriteFile(programCachePath, bytesToWrite, 0777); err != nil {
		panic(err)
	}
}

func ClearCache() {
	programCachePath := HandleTildePath(ProgramCachePath)
	os.Remove(programCachePath)
	Logger.Success("Cache file cleared")
}
