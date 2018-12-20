# The THORChain <--> Ethereum bifrost

## Introduction

This document acts as a guide to implementors of the THORChain/Ethereum Bifrost to allow for Ethereum and ERC-20 tokens to move between THORChain and Ethereum.

### Related Reading
This spec is heavily based on the Cosmos Peggy project spec (https://github.com/cosmos/peggy/blob/master/spec/readme.md) and new Cosmos oracle module (https://github.com/cosmos/cosmos-sdk/blob/master/examples/democoin/x/oracle/README.md). Make sure to read and understand both of those first.

Code from those projects will be relevant too, but the license has not yet been confirmed for Peggy (see https://github.com/cosmos/peggy/issues/44). The Solidity Ethereum smart contract code looks fairly usable, but may require some adaptation and additions. The Peggy Go code is mostly boilerplate and not really useful. The Rust witness code should probably not be used. The Cosmos oracle module looks usable, but may require some adaptation and additions.

## Design Principles
To ensure the Bifrost Protocol is widely extensible to most other blockchains, the following are the principles that are adhered to:

1) The protocol is assymetrically complex to THORChain. 
The modules that govern bridge logic and security are implementated on the THORChain side. All that is required on the blockchain side is an ability to process multi-signatures (bitcoin, monero) or multi-signature emulatations (ethereum). Sponsoring, registering and maintaining a bridge is all completed from THORChain modules. This reduces technical burden and allows bridges to scale to many blockchains.

2) The protocol is opt-in. 
Validators on THORChain opt-in to maintain bridges based on economic incentives only. This minimises risk to validators who may or may not want to support bridges from jurisdictional pressure, as well as ensuring the bridges are scalable across many blockchains. 

3) Bridges have observable security. 
Continuous liquidity pools allow asset pricing on THORChain which enables validators to preserve security thresholds on each bridge. Bridges with weak security are corrected by diluting the value of escrowed assets across more bridges. 

>For this Ethereum spec, we do not treat the process for nominating a new blockchain to be bridged (will require robust on-chain goverance), so we require that the bridge is already instantiated with the starting signature set in the deployed multi-signature. 

## Overview
The Bifrost MVP will add functionality to THORChain that allows the movement of Ether and ERC20 from Ethereum to THORChain and back, where they are represented as ERC20 tokens. It will not support movement of RUNE or THORChain native tokens to Ethereum.

The MVP is split into four components:
 - the Ethereum smart contract for sending tokens into the Bifrost from Ethereum `ethBridge`
 - the current THORChain Full Node and Validator `thorchaind`, with additional bifrost `tcBrifrost` and oracle modules `tcOracle`
 - a signer process, that detects Bifrost requests to send to Ethereum, and signs them appropriately for Ethereum `tcSigner`
 - a relayer process, that detects transactions to relay between THORChain and Ethereum `tcRelayer`

Each validator should run the relayer and signer process along with their THORChain process, which will watch both THORChain (through a local endpoint/socket) and Ethereum to trigger relaying. Validators can elect to run full or remote nodes (such as through infura) and is a optimisation decision only. 

### Prerequisites

1) At least 4 validators running all processes
2) All validators have a hot wallet with ether to pay for gas, and have logged their public addresses

### Instantiating a Bridge

1) A sponsoring validator "sponsor" proposes a new bridge to ethereum, by deploying the factory `ethBridge` contract. 
2) Once deployed, the sponsor adds the `ethBridge` contract address in the `tcBridgeRegistry` module. 
3) Opt-in validators nominate their public ethereum addresses in the `tcBridgeRegistry` module.
4) Once the minimum quorum number `n + 1` is reached, the sponsor adds `n` validators as signatories to the `ethBridge` contract. 

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
