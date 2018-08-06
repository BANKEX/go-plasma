# BANKEX Foundation Plasma block producer implementation

# NOT READY FOR PRODUCTION, WIP, FOR DEMO AND TESTING PURPOSES ONLY

## Technical overview

In Plasma implementation the block processor allows a centralized (or federated) party to process transactions at much higher speed due to pratial centralization. Such centralization comes with the price of sophisticated smart-contract in Ethereum network that allows verifier (all other users who observe a chain, but can not produce blocks) to automatically resolve disputes and fraud attempts such as double-spends. Our Plasma implementation follows the UTXO model better known as More Viable Plasma.

## Workflow of the block processor

A well-written block processor should correctly process the following 3 events:
- Event of deposit of some asset to the Plasma: some asset is deposited on the smart-contract in Ethereum network and a new unspent UTXO should be granted for the user with the same asset type, amount and correct ownership.
- Event of exit: user wants to move his asset back to the Ethereum network. There are two options:
  - Users behaves correcrly and the UTXO (an asset record) he wants to move to Ethereum network is not spent. In this case such UTXO should be marked as "spent" and further transactions with this UTXO should be forbidden.
  - User has already spent such an UTXO. In this case a message should be sent to the smart-contract in Ethereum network for purposes of challenge. Upon succesful challenge a process of "exiting" of this UTXO will be blocked by the smart-contract, so no spent assets are allowed to be moved from Plasma to Ethereum network.
- Normal transaction in Plasma: a user sends a serialized transaction with a set of UTXOs he wants to spend, set of UTXOs he wants to create and his signature. Such transaction should guarantee that:
  - "input" UTXOs are not yet spend
  - user has an ownership right on "input" UTXOs
  - total amount of assets in "inputs" is equalt to the total amount in "outputs"
  
  ## Some details
  
  This implementation is written in the `Go` programming language for ease of concurrency and reuse of some functions (such as `Keccak256` hash function) from the Ethereum [reference implementation](https://github.com/ethereum/go-ethereum). An asymmetric cryptography is using the same curve `secp256k1` that is used in Ethereum network with additions of safe separate contexts for use in multiple go-routines. Transaction format can be found [here](https://github.com/BANKEX/PlasmaParentContract) along with the example of the governing smart-contract.
  This implementation used `FoundationDB` key-value database for storage purposes. While this database is ACID and released by the large corporate player, there are not many reports of production use of this database, so proceed with caution!
  
## Authors

Alex Vlasov, [@shamatar](https://github.com/shamatar),  alex.m.vlasov@gmail.com

## License

All the original code in this repository is available under the Apache License 2.0 license. See the [LICENSE](https://github.com/BankEx/go-plasma/blob/master/LICENSE) file for more info.
