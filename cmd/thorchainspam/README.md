# Scripts to create load/spam on the THORChain networks

## Installation

`make install_spam`

## Creating load

### 1. Create accounts.

Currently, Cosmos does not allow more than 1 tx per block per account – so we need to create enough accounts to create the desired load. The following script creates k = 1000 accounts with key name "spam-0" to "spam-999" and transfers 200 RUNE to each of them, signed by the key was name "example". They all get the same password "1234567890".

`thorchainspam account ensure --from deployer --k 1000 --amount 200RUNE --chain-id genesis-alpha --spam-prefix spam --spam-password 1234567890 --sign-password my-secret`
