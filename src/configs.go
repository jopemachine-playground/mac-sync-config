package src

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/utils"
	"github.com/keybase/go-keychain"
	"gopkg.in/yaml.v3"
)

const (
	PREFERENCE_DIR_PATH   = "~/Library/Preferences/Mac-sync-config"
	MAC_SYNC_CONFIGS_FILE = "mac-sync-configs.yaml"
)

var (
	LocalPreferencePath = strings.Join([]string{PREFERENCE_DIR_PATH, "local-preference.json"}, "/")
	KeychainPreference  = GetKeychainPreference()
)

var (
	Flags struct {
		Overwrite bool
	}
)

func GetRemoteConfigFolderName() string {
	if folderName := os.Getenv("MAC_SYNC_CONFIG_FOLDER"); folderName != "" {
		return folderName
	}

	return ".mac-sync-configs"
}

func GetGitBranchName() string {
	if branchEnv := os.Getenv("MAC_SYNC_CONFIG_BRANCH"); branchEnv != "" {
		return branchEnv
	}
	return "main"
}

type MacSyncConfigs struct {
	ConfigPathsToSync []string `yaml:"sync"`
}

type KeychainPreferenceType struct {
	GithubId                       string `json:"github_id"`
	GithubToken                    string `json:"github_token"`
	MacSyncConfigGitRepositoryName string `json:"mac_sync_config_git_repository_name"`
}

func scanKeyChainPreference(config *KeychainPreferenceType) {
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

func GetKeychainPreference() KeychainPreferenceType {
	var config KeychainPreferenceType

	dat, err := keychain.GetGenericPassword("Mac-sync-config", "jopemachine", "Mac-sync-config", "org.jopemachine")

	// TODO: Improve below error handling logic.
	// If not exist, create new preference config file
	if len(dat) == 0 {
		scanKeyChainPreference(&config)
		bytesToWrite, err := json.Marshal(config)
		Utils.FatalExitIfError(err)

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
			Utils.FatalExitIfError(err)
			err = keychain.AddItem(keyChainItem)
		}

		Utils.FatalExitIfError(err)

		Logger.Success(fmt.Sprintf("mac-sync-config's configuration is saved successfully on the keychain.\n"))
	} else if err != nil {
		log.Fatal(err)
	} else {
		if json.Unmarshal(dat, &config); err != nil {
			Logger.Error("JSON data seems to be malformed or outdated.\nPress \"y\" to enter new information or press \"n\" to ignore it.")
			yes := Utils.MakeYesNoQuestion()
			if yes {
				keychain.DeleteGenericPasswordItem("Mac-sync-config", "jopemachine")
				Logger.Success("Keychain data deleted successfully.")
			}
		}
	}

	return config
}

func ReadLocalPreference() map[string]string {
	return ReadJSON(LocalPreferencePath)
}

func WriteLocalPreference(localPreference map[string]string) {
	localPreferenceDir := RelativePathToAbs(PREFERENCE_DIR_PATH)
	localPreferencePath := RelativePathToAbs(LocalPreferencePath)

	if _, err := os.Stat(localPreferenceDir); errors.Is(err, os.ErrNotExist) {
		Utils.FatalExitIfError(os.MkdirAll(localPreferenceDir, os.ModePerm))
	} else if _, err := os.Stat(localPreferencePath); !errors.Is(err, os.ErrNotExist) {
		os.Remove(localPreferencePath)
	}

	WriteJSON(localPreferencePath, localPreference)
}

func ReadMacSyncConfigFile(filepath string) (MacSyncConfigs, error) {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		return MacSyncConfigs{}, err
	}

	dat, err := ioutil.ReadFile(filepath)
	Utils.FatalExitIfError(err)

	var config MacSyncConfigs

	Utils.FatalExitIfError(yaml.Unmarshal(dat, &config))

	return config, nil
}
