package src

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/src/utils"
	"github.com/keybase/go-keychain"
)

const (
	CachePath      = "~/Library/Caches/Mac-sync-config"
	PreferencePath = "~/Library/Preferences/Mac-sync-config"
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

var (
	Flag_OverWrite = false
)

func GetRemoteConfigFolderName() string {
	// #TODO: Add env variable here
	return ".mac-sync-configs"
}

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
	Logger.Info("Please enter some information for accessing your Github repository.")
	Logger.Info("This information will be stored in your keychain.")
	Logger.NewLine()

	Logger.Question("Enter your Github ID:")
	ghId := bufio.NewScanner(os.Stdin)
	ghId.Scan()
	config.GithubId = ghId.Text()

	Logger.Question("Enter your Github Access Token:")
	ghToken := bufio.NewScanner(os.Stdin)
	ghToken.Scan()
	config.GithubToken = ghToken.Text()

	Logger.Question("Enter Git repository name for saving the mac-sync-config's configuration files:")
	repoName := bufio.NewScanner(os.Stdin)
	repoName.Scan()
	config.MacSyncConfigGitRepositoryName = repoName.Text()
}

func ReadPreference() Preference {
	preferenceDirPath := RelativePathToAbs(PreferencePath)

	var config Preference

	dat, err := keychain.GetGenericPassword("Mac-sync-config", "jopemachine", "Mac-sync-config", "org.jopemachine")

	// If not exist, create new preference config file
	if len(dat) == 0 {
		err := os.Mkdir(preferenceDirPath, os.ModePerm)
		if err != nil && !errors.Is(err, os.ErrExist) {
			panic(err)
		}

		scanPreference(&config)
		bytesToWrite, err := json.Marshal(config)
		Utils.PanicIfErr(err)

		keyChainItem := keychain.NewItem()
		keyChainItem.SetSecClass(keychain.SecClassGenericPassword)
		keyChainItem.SetService("Mac-sync-config")
		keyChainItem.SetAccount("jopemachine")
		keyChainItem.SetLabel("Mac-sync-config")
		keyChainItem.SetAccessGroup("org.jopemachine")
		keyChainItem.SetData([]byte(bytesToWrite))
		keyChainItem.SetSynchronizable(keychain.SynchronizableNo)
		keyChainItem.SetAccessible(keychain.AccessibleWhenUnlocked)

		err = keychain.AddItem(keyChainItem)

		if err == keychain.ErrorDuplicateItem {
			err = keychain.DeleteItem(keyChainItem)
			Utils.PanicIfErr(err)
			err = keychain.AddItem(keyChainItem)
		}

		Utils.PanicIfErr(err)

		Logger.Success(fmt.Sprintf("mac-sync-config's configuration is saved successfully on the keychain.\n"))
	} else if err != nil {
		Utils.PanicIfErr(err)
	} else {
		err = json.Unmarshal(dat, &config)
		Utils.PanicIfErrWithMsg("Json data malformed. Please delete and remake the keychain", err)
	}

	return config
}

func ReadConfigFileLastChanged() map[string]string {
	configFileLastChangedCachePath := RelativePathToAbs(ConfigFileLastChangedCachePath)

	if _, err := os.Stat(configFileLastChangedCachePath); errors.Is(err, os.ErrNotExist) {
		return make(map[string]string)
	}

	dat, err := ioutil.ReadFile(configFileLastChangedCachePath)
	Utils.PanicIfErr(err)

	var lastChangedMap map[string]string

	err = json.Unmarshal(dat, &lastChangedMap)
	Utils.PanicIfErr(err)

	return lastChangedMap
}

func WriteConfigFileLastChanged(lastChanged map[string]string) {
	cacheDir := RelativePathToAbs(CachePath)
	configFileLastChangedCachePath := RelativePathToAbs(ConfigFileLastChangedCachePath)

	if _, err := os.Stat(cacheDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(cacheDir, os.ModePerm)
		Utils.PanicIfErr(err)
	} else if _, err := os.Stat(configFileLastChangedCachePath); !errors.Is(err, os.ErrNotExist) {
		os.Remove(configFileLastChangedCachePath)
	}

	bytesToWrite, err := json.Marshal(lastChanged)
	Utils.PanicIfErr(err)

	err = ioutil.WriteFile(configFileLastChangedCachePath, bytesToWrite, os.ModePerm)
	Utils.PanicIfErr(err)
}

func ClearCache() {
	configFileLastChangedCachePath := RelativePathToAbs(ConfigFileLastChangedCachePath)

	err := os.Remove(configFileLastChangedCachePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
	Logger.Success("Cache file cleared.")
}

func PrintConfig() {
	resp, err := CreateMacSyncConfigRequest("mac-sync-configs.yaml")
	Utils.PanicIfErr(err)
	Logger.Log(resp.String())
}
