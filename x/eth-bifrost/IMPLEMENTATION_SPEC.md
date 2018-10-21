# The THORChain <--> Ethereum bifrost

## Introduction

This document acts as a guide to implementors of the THORChain/Ethereum Bifrost to allow for Ethereum and ERC-20 tokens to move between THORChain and Ethereum.

## Related Reading
This spec is heavily based on the Cosmos Peggy project spec (https://github.com/cosmos/peggy/blob/master/spec/readme.md) and new Cosmos oracle module (https://github.com/cosmos/cosmos-sdk/blob/master/examples/democoin/x/oracle/README.md). Make sure to read and understand both of those first.

Code from those projects will be relevant too, but the license has not yet been confirmed for Peggy (see https://github.com/cosmos/peggy/issues/44). The Solidity Ethereum smart contract code looks fairly usable, but may require some adaptation and additions. The Peggy Go code is mostly boilerplate and not really useful. The Rust witness code should probably not be used. The Cosmos oracle module looks usable, but may require some adaptation and additions.

The spec is written under the assumption that the LICENCE for Peggy will be Apache 2.0, like the Cosmos SDK, or that all code can be easily used. The project has not been completed nor updated in the last few months. I suggest contacting the Peggy team and possibly even considering asking for a grant to do the work of getting it completed.

## Relation to main Bifrost paper
The main Bifrost paper describes a more complex and possibly more secure system with a custom `m of n` validator signatures required, support for a changing validator set, random reshuffling and VRFs, SPV proofs and gossip-layer communication. The spec is constrained to a very simple MVP that can be used for a Testnet and get real users running and playing with it asap, and does not include any of those features or security/performance improvements.

## Overview
The Bifrost MVP will add functionality to THORChain that allows the movement of Ether and ERC20 from Ethereum to THORChain and back, where they are represented as ERC20 tokens. It will not support movement of RUNE or THORChain native tokens to Ethereum.

The MVP is split into three components:
 - the Ethereum smart contract for sending tokens into the Bifrost from Ethereum
 - the current THORChain Full Node and Validator, with additional bifrost and oracle modules
 - a signer process, that detects Bifrost requests to send to Ethereum, and signs them appropriately for Ethereum
 - a relayer process, that detects transactions to relay between THORChain and Ethereum

Each node should run the relayer and signer process along with their THORChain process, which will watch both THORChain (through a local endpoint/socket) and Ethereum to trigger relaying. For the initial MVP, we can just use Infura for watching for and submitting transactions on Ethereum, but a short term priority thereafter will be to set up our own Ethereum full nodes.

## Process Flow

Here's an example of the flow of an asset transfer of Ether or an ERC20 token from the Ethereum blockchain to THORChain:

1. Alice sends a transaction with 10 Ether and a destination address to the Ethereum smart contract. The destination address is Alice's on THORChain.

2. Each relayer detects an Ethereum smart contract event from that transaction, and after the finality threshold (say, 100 blocks) attests to the fact by submitting a THORChain transaction to the oracle module

3. Once `m of n` transactions are received by the oracle, it triggers the Bifrost module kicking off the transfer.

4. The Bifrost module receives the message and credits the destination address from step 1 with 10 tEth, updating the effective supply tracked by the CLP.

5. Alice receives the 10 tEth on THORChain.

6. Alice now sends 4 tEth to Bob on THORChain.

7. Bob now sends those 4 tEth to Ethereum by submitting a transaction to the Bifrost module.

8. The signers each watch for and detect Bob's transaction. When they see it, they generate a signature and submit a transaction signing their approval to the Bifrost module.

9. Once any relayer sees that `m of n` transactions are received by the Bifrost module, it posts Bob's original message and all signed messages to the Ethereum smart contract. (Note: For these steps 8 and 9, we may want to consider implementing/adapting the oracle module into some kind of 'reverse Oracle' like module that deals with aggregation and pruning of votes outside of the Bifrost module)

10. The ethereum smart contract credits Bob with 4 Ether.

Alice started with 10 Ether on Ethereum. She sent all of them to THORChain. There she sent 4 tEth to Bob. Bob then redeemed those 4 tEth back to Ethereum.

The end result is that Alice holds 6 tEth and 0 Ether and Bob holds 4 Ether and 0 tEth.

# Components

## Ethereum smart contract component

The smart contracts verify updates coming from THORChain
using the known keys of the signing apps. The smart contracts
track updates to the set of signer components, and their associated
signatures. The smart contracts supports 6 functions:

* `lock` ETH or ERC20 tokens for use in THORChain
* `unlock` previously-locked (encumbered) ETH or ERC20 tokens
* `update` signing app set signatures
* `mint` ERC20 tokens for encumbered denominations
* `burn` ERC20 tokens for encumbered denominations
* `register` denomination

## Oracle and Bifrost modules

The Oracle module, adapted from Cosmos's existing module is responsible for accepting transactions from multiple validators and waiting until an `m on n` thresold to then trigger an action in another module

The bifrost module is responsible for accepting actions from the Oracle module which result in creation of new tEth or tERC20 on THORChain, as well as accepting transactions from users who want to send tEth/tERC20 back out the Bifrost into Ethereum. It is responsible for managing transfers, mints/burns, and clp changes as part of these processes.

## Signer component

The signing apps sign transactions using secp256k1 such that the Ethereum smart contracts can verify them. The signing apps also have an ethereum address, because they have an identity in the Ethereum contract. They watch for new Ethereum-bound transactions using the ABCI app's query functionality, and submit their signatures in a transaction back to that ABCI.

## Relayer component

The relayer process is responsible for communication of state changes between Tendermint and Ethereum. It is stateless, and has at-least-once delivery semantics from one chain to another. If multiple relayers submit the same transaction to Ethereum and collide, they should get a nice error that informs them that the job is already done. If one relayer submits the same transaction to THORChain more than once, it should get a nice error informing it that the transaction has already been received. Then, every update delivered to either chain is idempotent.

Generally anyone that wants their Bifrost to be successful has an incentive to run the relayer process.

# Implementation steps
 - Quick audit of solidity codebase for Peggy to determine what can be used and what needs to be fixed
 - Fork of Peggy solidity codebase to be adapted into Bifrost Ethereum contract with additional needed changes
 - Testnet (likely Ropsten) setup and deployment of Bifrost Ethereum contacts, and of some test ERC20 contracts, test ERC20 RUNE contract to be used for testing
 - Quick audit of cosmos Oracle module to determine what can be used and what needs to be fixed
 - Completion of Oracle module functionality
 - Development of Bifrost module that accepts actions from the oracle module and mints tEth/tERC20 on THORChain and controls whitelisting
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

In the short term, improvements like running our own ethereum full nodes, and distributing our architecture such that relayers, signers, ethereum full nodes, thorchain full nodes and thorchain validator nodes to not need to run on the same virtual machine will be important reliability/performance and security improvements to make.

In the medium term, improvements to deal with potential malicious validator behaviour and unhappy paths in the sysem, like exploring fraud proofs for some scenarios and considering governance solutions should be explored.

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
