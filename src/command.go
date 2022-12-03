package src

func PrintMacSyncConfigs() {
	Logger.Log(GetMacSyncConfigs())
}

func SwitchProfile(profileName string) {
	localPreference := ReadLocalPreference()
	localPreferencePath := RelativePathToAbs(LocalPreferencePath)

	localPreference["profile"] = profileName
	WriteJSON(localPreferencePath, localPreference)
}
