package src

import (
	MacSyncConfig "github.com/jopemachine/mac-sync-config/src"
)

func PrintMacSyncConfigs() {
	MacSyncConfig.Logger.Log(MacSyncConfig.Github.GetMacSyncConfigs())
}
