
# Solidity

## Configure Development Environment

Install the project dependencies:

```sh
npm install
```

The project uses [GitHub Package Registry](https://github.com/orgs/keep-network/packages)
for dependencies. It requires `npm login` to be executed to authenticate with 
GitHub account. You can login with GitHub access token by providing your username
and instead of password use the token.

## Usage

Currently contracts can be installed as npm dependency, in the future we may
consider supporting [EthPM](http://www.ethpm.com/).

<!-- 
TODO: Configure EthPM, publish contracts and use them where needed.
https://www.trufflesuite.com/docs/truffle/reference/configuration#ethpm-configuration
-->

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
```sh
brew cask install ganache
```

Open Ganache app and configure a server to be exposed with hostname `127.0.0.1` 
on port `8545`.

### Deploy contracts

To deploy contracts ensure Ganache is running and Truffle configured. If all is set
run:

```sh
truffle migrate --reset
```

Command will output details of deployed contracts, find `contract address` value
for each contract and copy-paste it to [config.toml](../configs/config.toml) file.


[Truffle]: https://www.truffleframework.com/truffle
[Ganache]: https://truffleframework.com/ganache

### Setup Client

After scripts were migrated it is required to update client configuration. This
is covered by running following script:

```sh
KEEP_ETHEREUM_PASSWORD=password truffle exec scripts/setup-operator.js ../configs/config.toml
```

### Testing

#### Unit

Unit tests use Truffle's test framework, and redeploy contracts for a clean environment every test. An example:

```sh
truffle test test/ECDSAKeepTest.js
```

#### Scenarios

Tests in `test/integration/` are for testing different scenarios in the Go client. They do **not** redeploy contracts, instead using the already deployed instances from `truffle migrate`.

```sh
truffle exec test/integration/keep_signing.js
```
