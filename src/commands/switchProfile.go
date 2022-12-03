package src

import (
	MacSyncConfig "github.com/jopemachine/mac-sync-config/src"
)

func SwitchProfile(profileName string) {
	localPreference := MacSyncConfig.ReadLocalPreference()
	localPreferencePath := MacSyncConfig.RelativePathToAbs(MacSyncConfig.LocalPreferencePath)

	localPreference["profile"] = profileName
	MacSyncConfig.WriteJSON(localPreferencePath, localPreference)
}
