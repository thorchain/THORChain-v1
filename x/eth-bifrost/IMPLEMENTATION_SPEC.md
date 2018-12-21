# The THORChain <--> Ethereum Bifrost

## Introduction

This document acts as a guide to implementors of the THORChain/Ethereum Bifrost to allow for Ethereum and ERC-20 tokens to move between THORChain and Ethereum.

### Related Reading
This spec is heavily based on the Cosmos Peggy project spec (https://github.com/cosmos/peggy/blob/master/spec/readme.md) and new Cosmos oracle module (https://github.com/cosmos/cosmos-sdk/blob/master/examples/democoin/x/oracle/README.md). Make sure to read and understand both of those first.

## Design Principles
To ensure the Bifrost Protocol is widely extensible to most other blockchains, the following are the principles that are adhered to:

1) The protocol is assymetrically complex to THORChain. 
The modules that govern bridge logic and security are implementated on the THORChain side. All that is required on the blockchain side is an ability to process multi-signatures (bitcoin, monero) or multi-signature emulatations (ethereum). Sponsoring, registering and maintaining a bridge is all completed from THORChain modules. This reduces technical burden and allows bridges to scale to many blockchains.

2) The protocol is opt-in. 
Validators on THORChain opt-in to maintain bridges based on economic incentives only. This minimises risk to validators who may or may not want to support bridges from jurisdictional pressure, as well as ensuring the bridges are scalable across many blockchains. 

3) Bridges have observable security. 
Continuous liquidity pools allow asset pricing on THORChain which enables validators to preserve security thresholds on each bridge. Bridges with weak security are corrected by diluting the value of escrowed assets across more bridges. 

4) Bridge parties are never trusted. 
An important distinction with the Bifrost Protocol is that it is signed by `m of n of k` signatures, which is a subset to the full validator set. This ensures liveness to the bridge, and allows a dynamic validator set, and by extension, a dynamic signing set. 

`m` - minimum number of signatures
`n` - quorum size
`k` - maximum number of validators able to be quorum members

> As starting defaults we choose `7 of 10 of 15`: 7 signatures required per bridge, quorum size of 10, with 15 cycled through the quorum. 

## Overview
The Bifrost MVP will add functionality to THORChain that allows the movement of Ether and ERC20 from Ethereum to THORChain and back, where they are represented as ERC20 tokens. For now it will not support movement of RUNE or THORChain native tokens to Ethereum.

The MVP is split into five components:
 - the Ethereum smart contract for sending tokens into the Bifrost from Ethereum `ethBridge`
 - the current THORChain Full Node and Validator `thorchaind`, with additional bifrost `tcBifrost` and oracle modules `tcOracle`
 - a signer process, that detects Bifrost requests to send to Ethereum, and signs them appropriately for Ethereum `tcSigner`
 - a relayer process, that detects transactions to relay between THORChain and Ethereum `tcRelayer`
 - a datastore containing the registry and configuration for bridges `tcBridgeRegistry`

Each validator should run the relayer and signer process along with their THORChain process, which will watch both THORChain (through a local endpoint/socket) and Ethereum to trigger relaying. Validators can elect to run full or remote nodes (such as through infura) and is a optimisation decision only. 

### Prerequisites

1) At least 15 validators running all processes.

### Registering a Bridged Blockchain

Validators elect to add blockchains to THORChain's Bifrost. They propose a new blockchain which contains details on how to sync and report on chainstate, as well as deterministic information about the blockchain. 

1) A sponsoring validator "sponsor" proposes a new entry to `tcBridgeRegistry` with flag `proposed` which contains:

```go
type bridgeProposal struct {
  Name              string              //  Name of Bridged Blockchain
  Ticker .          string              //  Ticker of base token
  Description       string              //  Description of Blockchain
  URI               string              //  Resource URI for blockchain (logo, website, github)
  factoryContract   bytes               //  ABI code for the factory contract
  RelayerInst       string              //  Plaintext of how to link Relayer (including code for ABI)
  RelayerData       relayerData .       //  Object describing relayerData
  ChainData         chainData           //  Object describing chainData
  Cycle             int64               //  Number of blocks to cycle quorum
  Confirmations     int64               //  Number of confirmations to observe for on chain
  QuorumBuffer      int64               //  Number of blocks that a validator can enter quorum with
  minAssets         int64               //  The minimum amount of ether on the public address for gas
  Votes             map[string]int64    //  Votes for each option (Yes, No, Abstain)
  Address           map[uint256]address //  Public addresses mapped after voting
}
```

ChainData is reported as:

```go
type chainData struct {
  blockHeight       int64               //  Blockheight declared
  blockHash         uint256             //  Blockhash at Blockheight
  chainHash         uint256             //  Hash of full chain data
}
```

RelayerData is reported as:
```go
type relayerData struct {
  dataPath        string               //  Filepath to the chaindata '/home/user/gethDataDir/geth/chaindata'
  confRPC         uint256              //  API or RPC to get conf number 'web3.eth.blockNumber-web3.eth.getTransaction(<txhash>).blockNumber'
  chainHash         uint256             //  Hash of full chain data
}
```

>The Relayer is instructed on where to look for the new chaindata on disk by registering the filepath to the chaindata, such as `'/home/user/gethDataDir/geth/chaindata'`. Once linked, the relayer will scan the chaindata for the matching blockheight and hash, using `confRPC` to check for confirmation count. These details can be over-ridden in a local config file if the validator is using a different client. 

2) Via Cosmos on-chain governance, the proposal is accepted and flagged as `approved` once 15 signatures are collected:
* 15 validators must agree with the declared blockheight, blockhash and chaindata hash. 
* Accepting validators register their public address in the data store

>The index accepted for the blockchain now becomes the CLP index. The CLP inherits the proposed details of the blockchain (Name, ticker, description, URI). All 15 validators are now added to `tcBridgeRegistry` as quorum-compliant. 

> If a group of validators disagree with the chaindata, they can propose a corrected proposal. If both proposals are accepted, then two different blockchains are linked - this is fine, since it indicates the external blockchain is forked. 

The quorum list in `tcBridgeRegistry` now contains the full list of validators in the following states:

* `Non-compliant` - reporting outdated chaindata (older than `QuorumBuffer`)
* `Compliant` - reporting chaindata inside of `QuorumBuffer`
* `Quorum` - participating in quorum for a bridge. Requires a bridge index.
* `BridgeIndex` - if more than one bridge per blockchain.

3) Anyone can now send a genesis transaction `genTx` that will create the CLP for the bridge. The `genTx` has the following details:
* The `approved` CLP to `activate`.
* The Rune used in the `genTX` is the Rune side of the CLP.
* The Token side is created with 0 tokens, which effectively locks the CLP.
* The CLP is linked to whatever blockchain that is reported on in `tcBridgeRegistry` under the correct index.

The blockchain is now bridged from the THORChain side. Even without assets the shuffling process begin:

4) Every cycle the proposing validator removes the psuedo-randomly chosen validator from quorum from each bridge, and replaces them with a compliant validator.
* The `tcBifrost` module orchestrates the update to the states of validators `tcBridgeRegistry`.
* The `tcSigner` process then collects signatures for the `ethBridge` smart contract to enact the signature shuffling. 
* The `tcRelayer` process reports the events.

> If the number of bridges exceed `k - m`, then validators will be shuffled through each bridge quorum, before idling in a compliant state. 

> ChainData is continually reported to ensure compliant validators. 

#### Error Handling

1) Insufficient `minAssets` on validator wallets
>  Validators are required to keep the minimum amount of assets on-chain to pay for gas. If below, the become `non-compliant` with a reason

2) Mismatched/non-existant ChainData reported. 
> Validator is removed from quorum on next cycle, and labelled as `non-compliant`. Note - they may still be able to sign transactions. 


### Instantiating a Bridge

Once a bridge is registered, it needs to be functionally linked. 

1) A sponsoring validator "sponsor" proposes a new bridge to ethereum, by deploying the factory `ethBridge` contract. 
* The sponsor proposes the `ethBridge` contract address in the `tcBridgeRegistry` module.
* If validators agree the `ethBridge` contract meets the factory standard (check `factoryContract` bytecode), then they vote to integrate. 
* The sponsor adds all quorum validators to the `ethBridge` contract.

> Another

## User Experience Flow

1. Alice sends a transaction with 10 Ether and a destination address to the Ethereum smart contract. The destination address is Alice's on THORChain.

2. Each relayer detects an Ethereum smart contract event from that transaction, and after the finality threshold (say, 100 blocks) attests to the fact by submitting a THORChain transaction to the oracle module

3. Once `m of n` transactions are received by the oracle, it triggers the Bifrost module kicking off the transfer.

4. The Bifrost module receives the message and credits the destination address from step 1 with 10 tEth, updating the effective supply tracked by the CLP.

5. Alice receives the 10 tEth on THORChain.

6. Alice now sends 4 tEth to Bob on THORChain.

7. Bob now sends those 4 tEth to Ethereum by submitting a transaction to the Bifrost module, which includes his ethereum address and the exit fee `f`.

8. The signers each watch for and detect Bob's transaction. When they see it, they generate a signature and submit a transaction signing their approval to the Bifrost module.

9. Once any relayer sees that `m of n` transactions are received by the Bifrost module, it posts Bob's original message and all signed messages to the Ethereum smart contract. (Note: For these steps 8 and 9, we may want to consider implementing/adapting the oracle module into some kind of 'reverse Oracle' like module that deals with aggregation and pruning of votes outside of the Bifrost module)

10. The ethereum smart contract credits Bob with `4 - f` Ether, and `f / n` tEther to each of `n` participating validators.

Alice started with 10 Ether on Ethereum. She sent all of them to THORChain. There she sent 4 tEth to Bob. Bob then redeemed those 4 tEth back to Ethereum (minus the exit fee).

The end result is that Alice holds 6 tEth and 0 Ether and Bob holds `4 - f` Ether and 0 tEth, and `n` validators have `f / n` tEth.

### Updating a bridge

Rotating a validator:

1) Every `c` blocks on THORChain the `tcBifrost` module removes a validator, and adds a new one from the `tcBridgeRegistry` module. 
2) All validators sign the process on the `ethBridge` contract



# Components

## Ethereum bridge smart contract

The smart contracts verify updates coming from THORChain using the known keys of the signing apps. The smart contracts track updates to the set of signer components, and their associated signatures. The smart contracts supports 6 functions:

* `lock` ETH or ERC20 tokens for use in THORChain
* `unlock` previously-locked (encumbered) ETH or ERC20 tokens
* `update` signing app set signatures
* `register` denomination

### Lock

```
function lock(address _token, uint256 _value, uint256 _tcaddr) public payable returns(bool){

require(_value > 0, "Must be attempting to send a non zero amount of tokens");

        if(_token == address(0)){
            require(msg.value == _value, "Sender must send ETH as payment");
        } else {
            IERC20 token = IERC20(_token);                                        
            require(token.transferFrom(msg.sender, address(this), _value), "Sender must have approved the TKN transfer");
        }

        emit locked(_token, msg.sender, _tcaddr, _value);
    }
}
```

### Unlock, Update

Adaption of the [Gnosis Multsig](https://github.com/gnosis/MultiSigWallet/blob/master/contracts/MultiSigWallet.sol) which checks the number of signatories first before making a transfer. 


## Oracle and Bifrost modules

The Oracle module, adapted from Cosmos's existing module, is responsible for accepting transactions from multiple validators and waiting until an `m on n` threshold to then trigger an action in another module.

The Bifrost module is responsible for accepting actions from the Oracle module which result in creation of new tEth or tERC20 on THORChain, as well as accepting transactions from users who want to send tEth/tERC20 back out the Bifrost into Ethereum. It is responsible for managing transfers, mints/burns, and CLP changes as part of these processes.

## Signer component

The signing apps sign transactions using secp256k1 such that the Ethereum smart contracts can verify them. The signing apps also have an ethereum address, because they have an identity in the Ethereum contract. They watch for new Ethereum-bound transactions using the ABCI app's query functionality, and submit their signatures in a transaction back to that ABCI.

## Relayer component

The relayer process is responsible for communication of state changes between Tendermint and Ethereum. It is stateless, and has at-least-once delivery semantics from one chain to another. If multiple relayers submit the same transaction to Ethereum and collide, they should get an error that informs them that the job is already done. If one relayer submits the same transaction to THORChain more than once, it should get an error informing it that the transaction has already been received. Then, every update delivered to either chain is idempotent.


# Implementation steps
 - Quick audit of solidity codebase for Peggy to determine what can be used and what needs to be fixed
 - Fork of Peggy solidity codebase to be adapted into Bifrost Ethereum contract with additional needed changes (pending license approval)
 - Testnet (likely Ropsten) setup and deployment of Bifrost Ethereum contacts, and of some test ERC20 contracts, test ERC20 RUNE contract to be used for testing
 - Quick audit of Cosmos Oracle module to determine what can be used and what needs to be fixed
 - Completion of Oracle module functionality
 - Development of Bifrost module that accepts actions from the Oracle module and mints tEth/tERC20 on THORChain and controls whitelisting
 - Development of relayer service that watches Ethereum via Infura and submits transactions to the Oracle module
 - Tests and QA of transfers from Ether/ERC20 into THORChain
 - Enhance Bifrost module to be able to accept and burn tokens from users for sending back over the Bifrost to Ethereum
 - Development of signer service that watches Bifrost module for these transactions, generates Ethereum signatures and submits them back to the Bifrost module for future relaying (or, to the reverse-oracle module if that path is explored)
 - Enhance relayer to detect `m of n` signatures (or, to detect a `m of n` reverse-oracle state change if that path is explored) and relay them to the Ethereum smart contract
 - Tests and QA of transfers of tEth/tERC20 on THORChain back to Ethereum
 - Special rules for the official RUNE ERC20 contract to work differently to generic ERC20 contracts implemented
 - Tests and QA of RUNE-specific transfers into THORChain

# Future additions
Once this initial MVP is complete, there are some additions that can be made to improve the system.

In the short term, improvements like running ethereum full nodes, and distributing our architecture such that relayers, signers, ethereum full nodes, thorchain full nodes and thorchain validator nodes to not need to run on the same virtual machine will be important reliability/performance and security improvements to make.

In the medium term, improvements to deal with potential malicious validator behaviour and unhappy paths in the system, like exploring fraud proofs for some scenarios and considering governance solutions should be explored.

More long term, evolving the Bifrost into the more robust mechansims described in the main Bifrost whitepaper should become a priority (https://github.com/thorchain/Resources/blob/master/Whitepapers/Bifrost-Protocol/whitepaper-en.md). Some valuable additions will be the SPV and gossip network features, and an exploration of using VRFs.

# Appendix

## Example Ethereum Smart Contract External Entry Points (taken from Peggy)

### lock(bytes to, uint64 value, address token) external payable returns (bool)

Locks Ethereum user's ethers/ERC20s in the contract and loggs an event. Called by the users.

* `token` being `0x0` means ethereum; in this case `msg.value` must be same with `value`
* `event Lock(bytes to, uint64 value, address token)` is logged, seen by the relayers

### unlock(address[2] addressArg, uint64 value, bytes chain, uint16[] signers, uint8[] v, bytes32[] r, bytes32[] s) external returns (bool)

Unlocks Ethereum tokens according to the information from the pegzone. Called by the relayers.

* `addressArg[0]` == `to`, `addressArg[1]` == `token`
* transfer tokens to `to`
* hash value for `ecrecover` is calculated as:
```
byte(1) + to.Bytes() + value.PutUint64() + chain.length.PutUint256() + chain
```
