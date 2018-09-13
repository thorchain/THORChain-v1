# Scripts to create load/spam on the THORChain networks

## Installation

`make install_spam`

## Creating load

### 1. Create accounts.

The following script creates k = 100 accounts with key name "spam-0" to "spam-999" and transfers 200 RUNE to each of them, signed by the key was name "example". They all get the same password "1234567890".

`thorchainspam account ensure --from deployer --k 100 --amount 200RUNE --chain-id genesis-alpha --spam-prefix spam --spam-password 1234567890 --sign-password my-secret`

### 2. Create spam transactions.

The following script lets the created spam accounts send transactions between each other. The rate limit is the time to wait between each transaction in ms.

`thorchainspam txs send --chain-id genesis-alpha --spam-prefix spam --spam-password 1234567890 --rate-limit 20`