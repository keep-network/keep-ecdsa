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
