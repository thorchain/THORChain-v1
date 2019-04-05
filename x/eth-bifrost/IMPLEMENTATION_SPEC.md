# Cosmos <-> Ethereum Bridge Specification
By [Swish Labs](https://www.swishlabs.com/)
With contributions from Aidan Musnitzky, Jazear Brooks, Philip Stanislaus, JP Thor and Jessica Watson Miller

## Introduction

We present a design for a Cosmos-based decentralized bridge protocol, enabling the transfer of assets between blockchains without the use of a trusted third party such as an exchange. 

This document acts as a guide to implementers of the initial Cosmos/Ethereum bridge to allow for Ethereum (and later ERC-20 tokens and Cosmos-based tokens) to move between the Cosmos and Ethereum networks. It describes the initial implementation of a proof-of-concept unidirectional Ethereum -> Cosmos bridge, as well as the further implementation of the complete bidirectional bridge architecture.

This spec is partially based on the Cosmos [Peggy project spec](https://github.com/cosmos/peggy/blob/master/spec/readme.md) and the new Cosmos [oracle module](https://github.com/cosmos/cosmos-sdk/blob/master/examples/democoin/x/oracle/README.md). 

## Design Principles
To ensure this bridge protocol is widely extensible to most other blockchains, we apply the following principles:

1) **The protocol is assymetrically complex to Cosmos.**

The modules that govern bridge logic and security are implementated on the Cosmos side. All that is required on the external blockchain side is an ability to process multi-signatures (Bitcoin, Monero) or multi-signature emulatations (Ethereum). Sponsoring, registering and maintaining a bridge is all completed through Cosmos modules. This reduces the technical burden of implementing a bridge and allows bridges to scale to many blockchains.

2) **The protocol is opt-in.**

Validators on Cosmos opt-in to maintain bridges based only on economic incentives. This minimises the risk to validators who may not want to defend bridges from jurisdictional pressure, as well as ensuring the bridges are scalable across many blockchains by utilizing diverse validator sets. 

3) **Bridges have observable security.**

Cosmos asset pricing modules allow validators to preserve security thresholds on each bridge. Weak security on a given bridge can be rectified by diluting the value of escrowed assets across more bridges. 

4) **Bridge parties are never trusted.**

An important distinction with this bridge protocol is that it is signed by `m of n of k` signatures, which is a subset to the full validator set. This ensures the bridge's liveness, and allows a dynamic validator set - and by extension, a dynamic signing set. 

`m` - minimum number of signatures

`n` - quorum size

`k` - maximum number of validators able to be quorum members

> As starting defaults we choose `7 of 10 of 15`: 7 signatures required per bridge, quorum size of 10, with 15 cycled through the quorum. 

For the MVP we will adhere to 1) and 4), and relax 2) and 3).

## Overview
The unidirectional MVP will add functionality to Cosmos that allows the movement of Ether to Cosmos, where they are represented as cETH tokens. For now it will not support the movement of ATOM or Cosmos native tokens to Ethereum, or of ERC20 tokens to Cosmos.

The MVP is split into five components:
 - the Ethereum smart contract for sending tokens into the bridge from Ethereum `ethBridge`
 - the current Cosmos Full Node and Validator `cosmosd`, with additional bridge `cBridge` and oracle modules `cOracle`
 - a signer process, that detects bridge requests to send to Ethereum, and signs them appropriately for Ethereum `tcSigner`
 - a relayer process, that detects transactions to relay between Cosmos and Ethereum `cRelayer`
 - a datastore containing the registry and configuration for bridges `cBridgeRegistry`

Each validator should run the relayer and signer process along with their Cosmos process, which will watch both Cosmos (through a local endpoint/socket) and Ethereum to trigger relaying. Validators can elect to run full or remote nodes (such as through Infura); this makes no difference for the bridge and is purely an optimization decision on the part of individual validators. 

### Prerequisites

1) At least 15 validators running all processes.

### Registering a Bridged Blockchain

Validators elect to add blockchains to a given Cosmos bridge. Their proposal to add a new blockchain contains details on how to sync and report on chainstate, as well as deterministic information about the blockchain. 

1) A sponsoring validator "sponsor" proposes a new entry to `cBridgeRegistry` with flag `proposed` which contains:

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
  confRPC         string               //  API or RPC to get conf number 'web3.eth.blockNumber-web3.eth.getTransaction(<txhash>).blockNumber'
  chainHash         uint256             //  Hash of full chain data
}
```

>The Relayer is instructed on where to look for the new chaindata on disk by registering the filepath to the chaindata, such as `'/home/user/gethDataDir/geth/chaindata'`. Once linked, the relayer will scan the chaindata for the matching blockheight and hash, using `confRPC` to check for confirmation count. These details can be over-ridden in a local config file if the validator is using a different client. 

2) Via Cosmos on-chain governance, the proposal is accepted and flagged as `approved` once 15 signatures are collected:
* 15 validators must agree with the declared blockheight, blockhash and chaindata hash. 
* Accepting validators register their public address in the data store.

>The index accepted for the blockchain now becomes the CLP index. The CLP inherits the proposed details of the blockchain (Name, ticker, description, URI). All 15 validators are now added to `tcBridgeRegistry` as quorum-compliant. 

> If a group of validators disagree with the chaindata, they can propose a corrected proposal. If both proposals are accepted, then two different blockchains are linked - this is fine, since it indicates the external blockchain is forked. 

The quorum list in `cBridgeRegistry` now contains the full list of validators in the following states:

* `Non-compliant` - reporting outdated chaindata (older than `QuorumBuffer`)
* `Compliant` - reporting chaindata inside of `QuorumBuffer`
* `Quorum` - participating in quorum for a bridge. Requires a bridge index.
* `BridgeIndex` - if more than one bridge per blockchain.

3) Anyone can now send a genesis transaction `genTx` that will create the CLP for the bridge. The `genTx` has the following details:
* The `approved` CLP to `activate`.
* The Rune used in the `genTX` is the Rune side of the CLP.
* The Token side is created with 0 tokens, which effectively locks the CLP.
* The CLP is linked to whatever blockchain that is reported on in `tcBridgeRegistry` under the correct index.

The blockchain is now bridged from the THORChain side. Even without assets the shuffling process can begin:

4) Every cycle the proposing validator removes the psuedo-randomly chosen validator from quorum from each bridge, and replaces them with a compliant validator.
* The `tcBifrost` module orchestrates the update to the states of validators `tcBridgeRegistry`.

> If the number of bridges exceed `k - m`, then validators will be shuffled through each bridge quorum, before idling in a compliant state. 

> ChainData is continually reported to ensure compliant validators. 

#### Error Handling

1) Insufficient `minAssets` on validator wallets
>  Validators are required to keep the minimum amount of assets on-chain to pay for gas. If below, they become `non-compliant` with a reason

2) Mismatched/non-existent ChainData reported. 
> Validator is removed from quorum on next cycle, and labelled as `non-compliant`. Note - they may still be able to sign transactions. 


### Instantiating a Bridge

Once a bridge is registered, it needs to be functionally linked. 

1) A sponsoring validator "sponsor" proposes a new bridge to Ethereum, by deploying the factory `ethBridge` contract. 
* The sponsor proposes the `ethBridge` contract address in the `cBridgeRegistry` module.
* If validators agree the `ethBridge` contract meets the factory standard (check `factoryContract` bytecode), then they vote to integrate. 
* The sponsor adds all quorum validators to the `ethBridge` contract.

Once the bridge is linked, the cycle will now include the signer and relayer processes:

* The `cBridge` module orchestrates the update to the states of validators `cBridgeRegistry`.
* The `cSigner` process then collects signatures for the `ethBridge` smart contract to enact the signature shuffling. 
* The `cRelayer` process reports the events.

The bridge is ready. 


### Incoming Assets

1. Alice sends a transaction to the bridge contract:
* Includes the token type, value amount and a destination address to the Ethereum smart contract. The destination address is Alice's on THORChain.

2. Each relayer detects an Ethereum event from that transaction:
* After the confirmation threshold attests to the fact by submitting a THORChain transaction to the `tcOracle` module.
* Once `n` transactions are received by the oracle, it triggers the Bifrost module kicking off the transfer.

> Instead of `m`, `n` signatures are chosen since if the bridge is valid, 100% of the subset should be able to attest to the transaction. Additionally, it will be very disruptive to roll an incoming asset transfer back. 

4. The Bifrost module receives the message:
 * Credits the destination address with the asset value
 * Updates the effective supply tracked by the CLP.

> Alice receives the assets on her THORChain address.

#### Error Handling
The following are the errors to handle on the smart contract:
* Non-tracked token -> smart contract to reject
* Non-valid destination address -> smart contract to reject
* NaN value transfer -> reject

On the THORChain side:
* `m` signatures aren't collected from offline validators -> wait until the bridge is cycled.
* Incorrect value amount -> should not collect `n` signatures.
* Incorrect address credited -> should not collect `n` signatures.

> For a deep chain re-org, on-chain governance will be required to socialise losses (or underwrite them from a community fund).

### Future work on the bidirectional bridge protocol

?allows the movement of Ether and ERC20 from Ethereum to THORChain and back, where they are represented as ERC20 tokens. For now it will not support movement of RUNE or THORChain native tokens to Ethereum.

### Outgoing Assets

1. Bob now sends assets to Ethereum by submitting a transaction to the Bifrost module, 
* Includes his destination ethereum address with value `x`
* Exit fee is set as `f`.

> Note: this transaction is not finalised until 67% of validators commit it as a valid transaction. This enables fraud-proofs to be published in the next step. 

2. The bridge signers each watch for and detect Bob's transaction:
* Generate a signature and submit a transaction signing their approval to the Bifrost module.
* Once any relayer sees that `m of n` transactions are received by the Bifrost module, it posts Bob's original message and all signed messages to the Ethereum smart contract. 

3. The funds are dispersed:
* The ethereum smart contract credits Bob with `x - f` Ether
* The Bifrost Module credits `f / n` tEther to each of `n` participating validators.

>Bob holds `x - f` Ether and 0 tEth, and `n` validators have `f / n` tEth.

#### Error Handling
On the THORChain side:
* A validator may attempt to spend an invalid transaction -> slash
* Incorrect ethereum address -> error check first
* Incorrect value amount -> sanity check first

On smart contract side:
* Not all validators may have enough gas -> wait until bridge cycle
* Not all validators may be online -> wait until bridge cycle


# Components

## Ethereum bridge smart contract

The smart contracts verify updates coming from THORChain using the known keys of the signing apps. The smart contracts track updates to the set of signer components, and their associated signatures. The smart contracts supports 6 functions:

* `lock` ETH or ERC20 tokens for use in THORChain
* `unlock` previously-locked (encumbered) ETH or ERC20 tokens
* `update` signing app set signatures

### Lock

```
    /// @dev User to send incoming assets.
    /// @param _token Token to transfer (address(0) is ether).
    /// @param _value Value to transfer.
    /// @param _tcaddr Address of recipient.
    
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

Submitting a transaction:

```

    /// @dev Allows an owner to submit and confirm a transaction.
    /// @param destination Transaction target address.
    /// @param value Transaction ether value.
    /// @param data Transaction data payload.
    /// @return Returns transaction ID.
    function submitTransaction(address destination, uint value, bytes data)
        public
        returns (uint transactionId)
    {
        transactionId = addTransaction(destination, value, data);
        confirmTransaction(transactionId);
    }

    /// @dev Allows an owner to confirm a transaction.
    /// @param transactionId Transaction ID.
    function confirmTransaction(uint transactionId)
        public
        ownerExists(msg.sender)
        transactionExists(transactionId)
        notConfirmed(transactionId, msg.sender)
    {
        confirmations[transactionId][msg.sender] = true;
        Confirmation(msg.sender, transactionId);
        executeTransaction(transactionId);
    }

```

Updating signatories:

```
    /// @dev Allows to replace an owner with a new owner. Transaction has to be sent by wallet.
    /// @param owner Address of owner to be replaced.
    /// @param newOwner Address of new owner.
    function replaceOwner(address owner, address newOwner)
        public
        onlyWallet
        ownerExists(owner)
        ownerDoesNotExist(newOwner)
    {
        for (uint i=0; i<owners.length; i++)
            if (owners[i] == owner) {
                owners[i] = newOwner;
                break;
            }
        isOwner[owner] = false;
        isOwner[newOwner] = true;
        OwnerRemoval(owner);
        OwnerAddition(newOwner);
    }
```

## Cosmos Modules

### Oracle Module

The Oracle module, adapted from Cosmos's existing module, is responsible for accepting transactions from multiple validators and waiting until an `m on n` threshold to then trigger an action in another module.

### Bridge Module
The Bridge module is responsible for accepting actions from the Oracle module which result in creation of new tEth or tERC20 on THORChain, as well as accepting transactions from users who want to send tEth/tERC20 back out the Bifrost into Ethereum. It is responsible for managing transfers, mints/burns, and CLP changes as part of these processes.

#### Fee Sub-module
The fee component utilises the Relayer process in order to determine the loss of fungibility in the bridge, and uses outgoing transactions to restore it, as well as incentivising validators to be part of quorum. Users are shown the expected fees to exit, as well as choosing a miner fee. Miner fee pays for the gas of the outgoing transaction. 

**Exit Fee**
1) The total number of assets in the bridge are monitored `a`
2) The total number of assets in the CLP are monitored `a'`
3) The deficiency is `d = a' - a`
4) The miner fee `mfee` is calculated as `g * n`, where `g = gas estimated`
5) The exit fee is thus `2 * (d + mfee)`, charged to the user. 

**Validator Reward**
The exit fee pays the deficiency (if any), the mining fee, as well as the bridge reward, such that each validator party to the bridge receives `(d + mfee) / n`. 

> The multi-signature contract should be adjusted to pay out to `n` validators in ether the `mfee / n` with the outgoing transaction to cover the mining fee. If this is difficult, the mining fee can be paid alongside the actual reward in tEther, to prevent a continual unnecessary bleed of assets out of THORChain. 

#### Rebalancing Module

The rebalancing module continually tracks the total value of assets based on CLP pricing, and ensures they are evenly distributed along enough bridges. This prevents a bridge having more assets than is safe. 

1) The total value of `m` stakes in quorum is tracked `quorumStake`
2) The value of the bridge is monitored `v`, 
3) If `quorumStake < (2 * v)`, then the bridge is insecure, with the difference `s`
4) The difference `s` is then atomically sent to a bridge that does not exceed security in a designated cycle

#### Fraud-proof Module

The fraud-proof module process fraud-proofs and enables slashing of malicious validators. A user who wishes to exit a bridge performs the following transaction on Cosmos, called an `exitRequestTx`:

`exit[<coin>, <amount>, <bridge>, <destinationAddr>, <fee>]`

Once published on-chain, bridge nodes will be in possession of the `exitRequestTx` (instantly final) and any one of them will initiate the spend transaction by gossiping their signed spend transaction, based on the user’s request. The spend transaction will include a hash of the `exitRequestTx`. The rest of the nodes will also gossip their spend transaction and collect incoming signatures. Any gossiped `spendTx` that are received by Bridge nodes with the following conditions, will result in the offending node being booted from Quorum, and the `spendTx` marked as invalid:

* Missing an accompanying `exitRequestTx` hash
* SpendTx differing from exitRequestTx in any manner (addr, amount, fee)

This will prevent an incorrect (or fraudulent) spendTX from being propagated. If any of the nodes observe another node publishing on the bridgeChain a `spendTx`, the following is true:

* Any node can prove a fraud exists (`onChainSpendtx` differs to `exitRequestTx`)
* Any node can prove the fraudulent actor (the signed tx matches the previously declared Bitcoin Public Key).

> Note: the assets are still safe as the multi-sig bridge wallet requires a supermajority to sign on-chain.

The bridge nodes will then sign and gossip a `FraudProofTX` with the `onChainSpendtx` and the `exitRequestTx` to the Quorum. Each node in the Quorum that receives a `FraudProofTX` will add their signature if they agree, until 67% of the signatures are collected. At this point any node can publish the final signed FraudProofTx on-chain to THORChain. The block producer for that block (which may not have visibility of the BridgeChain) will be able to validate the `FraudProofTX` as long 67% of the quorum have signed (which they do have visibility of). The block will be fully-validated by 67% of THORChain validators, which includes a transaction to slash the offending Node and distribute to all non-offending Validators.

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

# Related reading

THORChain's Bifrost Whitepaper: https://github.com/thorchain/Resources/blob/master/Whitepapers/Bifrost-Protocol/whitepaper-en.md

THORChain's Bifrost Networking Spec: https://github.com/thorchain/Resources/blob/master/Whitepapers/Bifrost-Protocol/networking-whitepaper-en.md

# Additional sections

3) **Bridges have observable security.**

Asset pricing modules allow validators to preserve security thresholds on each bridge. Weak security on a given bridge can be rectified by diluting the value of escrowed assets across more bridges. 
