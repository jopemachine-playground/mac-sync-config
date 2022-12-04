package src

import (
	MacSyncConfig "github.com/jopemachine/mac-sync-config/src"
	Utils "github.com/jopemachine/mac-sync-config/utils"
	"github.com/keybase/go-keychain"
)

func DeleteKeychainConfig() {
	MacSyncConfig.Logger.Info("Press \"y\" to enter new information or press \"n\" to ignore it.")

	if yes := Utils.MakeYesNoQuestion(); yes {
		Utils.FatalExitIfError(keychain.DeleteGenericPasswordItem("Mac-sync-config", "jopemachine"))
		MacSyncConfig.Logger.Success("Keychain data deleted successfully.")
	}
}
