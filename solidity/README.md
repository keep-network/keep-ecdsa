
# Solidity

## Configure Development Environment

### Truffle

[Truffle] is a development framework for Ethereum.

To install it run:
```sh
npm install -g truffle
```

Configuration file `truffle-config.js` requires to contain a blockchain
connection details. See next section for information on running a test blockchain.

### Ganache

To start testing and developing you need to have a test blockchain set up. You 
can use [Ganache] for this.

 To install Ganache on MacOS run:
```shell
brew cask install ganache
```

Open Ganache app and configure a server to be exposed with hostname `127.0.0.1` 
on port `8545`.

### Deploy contracts

To deploy contracts ensure Ganache is running and Truffle configured. If all is set
run:
```
truffle migrate --reset
```

Command will output details of deployed contracts, find `contract address` value
for each contract and copy-paste it to [config.toml](../configs/config.toml) file.


[Truffle]: https://www.truffleframework.com/truffle
[Ganache]: https://truffleframework.com/ganache
