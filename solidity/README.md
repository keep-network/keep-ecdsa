
# Solidity

## Configure Development Environment

### NPM dependencies

The project uses [GitHub Package Registry](https://github.com/orgs/keep-network/packages)
for keep-network dependencies. It requires `npm login --registry=https://npm.pkg.github.com
to be executed to authenticate with GitHub account in order to access private
packages. You can login with GitHub access token by providing your username and
instead of password use the token.

Install the project dependencies:

```sh
npm install
```

### keep-core contracts

This project depends on contracts migrated by `keep-core` project and expects its
migration artifacts to be provided in `artifacts` directory.
The contracts can be fetched from published NPM package or local source (for development)
after running migrations in `keep-core` project.

To fetch required contracts addresses from local `keep-core` project use the following
commands:

- in `keep-core/solidity-v1` directory execute: `npm link`,
- in `keep-ecdsa/solidity` directory execute: `npm link @keep-network/keep-core`.

Remember that migration artifacts have to be available in `artifacts` directory. Truffle
by default places them in `build/contracts` so you need to copy them or create
a symlink.

### Staking and bonding

Keeps creation depends on operator's KEEP token staking and available bonding
value. To initialize the operator:

1. Initialize token staking in keep-core:
    ```sh
    # Run from `keep-core/contracts/solidity` directory
    truffle exec ./scripts/demo.js --network local
    ```

2. Initialize operator for Bonded ECDSA Keep Factory and TBTC System contract address.
    ```sh
    # Run from `keep-ecdsa/solidity` directory
    CLIENT_APP_ADDRESS="<TBTC_SYSTEM_ADDRESS>"\
        truffle exec scripts/lcl-initialize.js`
    ```

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


### Testing

#### Unit

Unit tests use Truffle's test framework, and redeploy contracts for a clean environment every test. An example:

```sh
truffle test test/BondedECDSAKeepTest.js
```

#### Scenarios

Tests in `test/integration/` are for testing different scenarios in the Go client. They do **not** redeploy contracts, instead using the already deployed instances from `truffle migrate`.

```sh
truffle exec test/integration/keep_signing.js
```
