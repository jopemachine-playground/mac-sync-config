package src

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/utils"
	"gopkg.in/yaml.v3"
)

const (
	PREFERENCE_DIR_PATH   = "~/Library/Preferences/Mac-sync-config"
	MAC_SYNC_CONFIGS_FILE = "mac-sync-configs.yaml"
)

var (
	LocalPreferencePath = strings.Join([]string{PREFERENCE_DIR_PATH, "local-preference.json"}, "/")
)

type MacSyncConfigs struct {
	ConfigPathsToSync []string `yaml:"sync"`
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
		Utils.FatalExitIfError(os.RemoveAll(localPreferencePath))
	}

	WriteJSON(localPreferencePath, localPreference)
}

func ReadMacSyncConfigFile(filepath string) MacSyncConfigs {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		Utils.FatalExitIfError(err)
		return MacSyncConfigs{}
	}

	dat, err := ioutil.ReadFile(filepath)
	Utils.FatalExitIfError(err)

	var config MacSyncConfigs

	Utils.FatalExitIfError(yaml.Unmarshal(dat, &config))

	return config
}
