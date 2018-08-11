# THORChain debug tool

Simple tool for simple debugging.

We try to accept both hex and base64 formats and provide a useful response.

Note we often encode bytes as hex in the logs, but as base64 in the JSON.

## Pubkeys

The following give the same result:

```sh
thorchaindebug pubkey TZTQnfqOsi89SeoXVnIw+tnFJnr4X8qVC0U8AsEmFk4=
thorchaindebug pubkey 4D94D09DFA8EB22F3D49EA17567230FAD9C5267AF85FCA950B453C02C126164E
```

## Txs

Pass in a hex/base64 tx and get back the full JSON

```sh
thorchaindebug tx <hex or base64 transaction>
```

## Hack

This is a command with boilerplate for using Go as a scripting language to hack
on an existing THORChain state.

Currently we have an example for the state of gaia-6001 after it
[crashed](https://github.com/cosmos/cosmos-sdk/blob/master/cmd/gaia/testnets/STATUS.md#june-13-2018-230-est---published-postmortem-of-gaia-6001-failure).
If you run `thorchaindebug hack $HOME/.thorchaind` on that
state, it will do a binary search on the state history to find when the state
invariant was violated.
