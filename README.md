# keep-tecdsa

## Prerequisites

Dependencies are managed by [Modules](https://github.com/golang/go/wiki/Modules) feature. 
To work in Go 1.11 it may require setting `GO111MODULE=on` environment variable.
```sh
export GO111MODULE=on
```

## Build

To build execute a command:
```sh
# Regenerate Solidity bindings
go generate ./...

go build .
```

## Test

To test execute a command:
```sh
go test ./...
```

---

## Docker

To build a Docker image execute a command:
```sh
docker build -t keep-tecdsa .
```

To run execute a command:
```sh
docker run -it keep-tecdsa keep-tecdsa sign <hash>
```
Where `<hash>` is a message to sign.

---

## Contracts

See [solidity](./solidity/) directory.

---

## Configuration

`configs/config.toml` is default path to the config file. To provide custom 
configuration CLI supports `--config` flag.
Sample configuration can be found in [config.toml.SAMPLE](configs/config.toml.SAMPLE).

---

## Smoke test

**Prerequisites**
- contracts deployed: `truffle migrate --reset`
- `ECDSAKeepFactory` contract address provided in `config.toml`
- ethereum account `KeyFile` path provided in `config.toml` and password to the
  key file provided as `KEEP_ETHEREUM_PASSWORD` environment variable
- [addresses of application](https://github.com/keep-network/keep-tecdsa/blob/69768219221e6adef7dbcaf0035237846f3523a5/solidity/integration/smoke_test.js#L27) provided in `config.toml` in `SanctionedApplications` list
- off-chain [client](#Client) running

To run a smoke test execute:
```sh
cd solidity/
truffle exec integration/smoke_test.js
```

---

## Client

To start a `keep-tecdsa` client execute:
```sh
LOG_LEVEL="debug" KEEP_ETHEREUM_PASSWORD="password" ./keep-tecdsa start
```

### Development

To start a client from source code execute:
```sh
LOG_LEVEL="debug" KEEP_ETHEREUM_PASSWORD="password" go run . start
```

---

## Sign hash

To run execute a command:
```sh
./keep-tecdsa sign <hash>
```
With `<hash>` as a string to sign.

Sample output:
```sh
➜  keep-tecdsa git:(signer) ./keep-tecdsa sign YOLO
--- Generated Public Key:
X: 2295845dbe5b5af2b516afa990e9113073793b6f861b66aa36e453e3a0e976f1
Y: d6d1923fa28c29d9fc2eb274cb54efc16875fab6d2d741e56a8afc7783e3f03b
--- Signature:
R: 6479fff99d7aa3f22d9b489f164a6e904abdb74d6cc44fa5b274903accee366a
S: 3ae8dcb534aa12c84214e7f448c5f60dbc048c64e60977d2b0a81b76cece76c8
```

---

## Publish transaction

Following ways of transaction publishing are supported:
- via Electrum network
- via Block Cypher API

To choose a service pass `--broadcast-api` flag to the execution command.

### Electrum

Transaction is submitted to Electrum Server to be published to a network.

#### Configuration

To configure connection details of Electrum Server update `config.toml`
 file. 

A list of Electrum Servers working on Bitcoin testnet can be found [here](https://1209k.com/bitcoin-eye/ele.php?chain=tbtc).

**NOTE**: Currently only SSL connection type is supported.

Sample config file content:

```toml
[electrum]
    ServerHost  = "testnet.hsmiths.com"
    ServerPort  = "53012"
```

#### Execution

To publish a raw transaction from the CLI run:

```sh
./keep-tecdsa --broadcast-api electrum publish <rawTx>
``` 
Sample output:
```sh
./keep-tecdsa --broadcast-api electrum publish 02000000000101506fda83a9788dab896b90bfc122c700afccd30459e10a0ff270e951202612481600000017160014997f3e8bcf47183fbbe0c8464175047bf391ea83feffffff0279d513000000000016001432b027edb95eee83b40003762a2ff25ae47d560d40420f000000000016001469abce7925fce369303e247c3a465447f6519b780247304402205be426e3e0c2e243eac4808cca306e3f821f7e431e2f0fc0c534851d5551779402202e98a4c7899d0eb2f473af77af6b434a5a47a0d898ed5b41ca2a7692601e36eb012102cb34ff4b355a7f02104cb912dbf4e35a14733030450ac311f6ef498d150a6ad2eb171700

Connected to Electrum Server.
Server version: ElectrumX 1.11.0 [Protocol 1.4]
2019/05/01 15:47:01 publish failed [transaction broadcast failed [errNo: 1, errMsg: the transaction was rejected by network rules.

transaction already in block chain
[02000000000101506fda83a9788dab896b90bfc122c700afccd30459e10a0ff270e951202612481600000017160014997f3e8bcf47183fbbe0c8464175047bf391ea83feffffff0279d513000000000016001432b027edb95eee83b40003762a2ff25ae47d560d40420f000000000016001469abce7925fce369303e247c3a465447f6519b780247304402205be426e3e0c2e243eac4808cca306e3f821f7e431e2f0fc0c534851d5551779402202e98a4c7899d0eb2f473af77af6b434a5a47a0d898ed5b41ca2a7692601e36eb012102cb34ff4b355a7f02104cb912dbf4e35a14733030450ac311f6ef498d150a6ad2eb171700]]]
```

### Block Cypher

Transaction is submitted to [Block Cypher's API](https://www.blockcypher.com/dev/bitcoin/) 
to be published on the chain.

#### Configuration

To configure Block Cypher communication details update `config.toml`
 file. 

 Sample config file content:
 ```toml
[blockcypher]
    Token   = "block-cypher-user-token"
    Coin    = "btc"
    Chain   = "test3"
 ```
To get a developer's `Token` register on Block Cypher [site](https://accounts.blockcypher.com).

#### Execution

To publish a raw transaction from the CLI run:

```sh
./keep-tecdsa --broadcast-api blockcypher publish <rawTx>
``` 
Sample output:
```sh
./keep-tecdsa --broadcast-api blockcypher publish 02000000000101506fda83a9788dab896b90bfc122c700afccd30459e10a0ff270e951202612481600000017160014997f3e8bcf47183fbbe0c8464175047bf391ea83feffffff0279d513000000000016001432b027edb95eee83b40003762a2ff25ae47d560d40420f000000000016001469abce7925fce369303e247c3a465447f6519b780247304402205be426e3e0c2e243eac4808cca306e3f821f7e431e2f0fc0c534851d5551779402202e98a4c7899d0eb2f473af77af6b434a5a47a0d898ed5b41ca2a7692601e36eb012102cb34ff4b355a7f02104cb912dbf4e35a14733030450ac311f6ef498d150a6ad2eb171700

2019/05/02 14:26:30 publish failed [transaction broadcast failed [HTTP 400 Bad Request, Message(s): Error validating transaction: Transaction with hash 30f42c9517fb26c93e172647fcf0ab50fa7dfe1f1e0eb9fe9a832ef81f990c47 already exists..]]
```
