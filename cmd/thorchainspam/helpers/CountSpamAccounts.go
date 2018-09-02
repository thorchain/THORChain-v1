package helpers

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

// Counts the number of spam accounts
func CountSpamAccounts(arr []keys.Info, spamPrefix string) int {
	count := 0

	for _, info := range arr {
		if strings.HasPrefix(info.GetName(), spamPrefix) {
			count++
		}
	}

	return count
}
