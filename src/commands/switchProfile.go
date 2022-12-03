package src

import (
	"fmt"

	MacSyncConfig "github.com/jopemachine/mac-sync-config/src"
)

func isValidName(profileName string) bool {
	// TODO: Add validation logic below
	return true
}

func SwitchProfile(profileName string) {
	localPreference := MacSyncConfig.ReadLocalPreference()
	if profileName == "" {
		MacSyncConfig.Logger.Info(fmt.Sprintf("Current profile is '%s'", localPreference["profile"]))
		return
	}

	if isValidName(profileName) {
		localPreference["profile"] = profileName
		MacSyncConfig.WriteLocalPreference(localPreference)
		MacSyncConfig.Logger.Success(fmt.Sprintf("Now current profile is '%s'.", profileName))
	} else {
		MacSyncConfig.Logger.Error(fmt.Sprintf("'%s' is not valid for profile name.", profileName))
	}
}
