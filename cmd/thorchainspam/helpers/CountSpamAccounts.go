package helpers

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/thorchain/THORChain/cmd/thorchainspam/constants"
)

func CountSpamAccounts(arr []keys.Info) int {
	count := 0

	for _, info := range arr {
		if strings.HasPrefix(info.GetName(), constants.SpamAccountPrefix) {
			count++
		}
	}

	return count
}
