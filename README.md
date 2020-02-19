# keep-tecdsa

---

## Contracts

See [solidity](./solidity/) directory.

---

## Developer's environment setup

To set up developers environment on MacOS execute:

```
./scripts/lcl-macos-setup.sh
```

---

## Quick installation

To quickly install and start a single client use the installation script.

### Prerequisites
To run the script some manual preparation is needed:

- [set up local ethereum chain](https://github.com/keep-network/keep-core/blob/master/docs/development/local-keep-network.adoc#setting-up-local-ethereum-client),
- [config file for the single client](#Configuration) (default name: `config.toml`),
- [npm authorized to access private packages in GitHub's Package Registry](./solidity/README.md#NPM-dependencies).

Please note that the client config file doesn't have to be pre-configured with contracts
addresses as they will be populated during installation.

### Install script

The `install.sh` script will:

- fetch external contracts addresses,
- migrate contracts,
- build client.

The script will ask you for the password to previously created ethereum accounts.



To start the installation execute:
```
./scripts/install.sh
```

### Initialize script

The `initialize.sh` script should be called after external customer application
contract (i.e. `TBTCSystem`) using keep-ecdsa is known. The script will:

- set address to the customer application,
- initialize contracts,
- update client contracts configuration.

The script will ask for the client config file path.

It also requires an external client application address which is an address of an 
external contract that will be requesting keeps creation. For local smoke test
execution this address should be the same as the account you will use in the smoke
test to request keep opening.

To start the initialization execute:
```
./scripts/initialize.sh
```

### Start client

To start the client execute:
```
./scripts/start.sh
```

---

## Go client

### Prerequisites

Dependencies are managed by [Modules](https://github.com/golang/go/wiki/Modules) feature. 
To work in Go 1.11 it may require setting `GO111MODULE=on` environment variable.
```sh
export GO111MODULE=on
```

### Build

To build execute a command:
```sh
# Regenerate Solidity bindings
go generate ./...

go build .
```

### Test

To test execute a command:
```sh
go test ./...
```

### Configuration

`configs/config.toml` is default path to the config file. To provide custom 
configuration CLI supports `--config` flag.
Sample configuration can be found in [config.toml.SAMPLE](configs/config.toml.SAMPLE).

### Run

To start a `keep-tecdsa` client execute:
```sh
LOG_LEVEL="debug" KEEP_ETHEREUM_PASSWORD="password" ./keep-tecdsa start
```

To start a client from source code execute:
```sh
LOG_LEVEL="debug" KEEP_ETHEREUM_PASSWORD="password" go run . start
```

### Docker

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

## Smoke test

**Prerequisites**
- system configured and running as per [quick installation instruction](#Quick-installation),
- sortition pool's registration period (_10 blocks_) passed since operators'
  registration.

To run a smoke test execute:
```sh
cd solidity/
truffle exec integration/smoke_test.js --network local
```

---
