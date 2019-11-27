module github.com/keep-network/keep-tecdsa

go 1.12

replace (
	github.com/BurntSushi/toml => github.com/keep-network/toml v0.3.0
	github.com/blockcypher/gobcy => github.com/keep-network/gobcy v1.3.1
	github.com/btcsuite/btcd => github.com/keep-network/btcd v0.0.0-20190427004231-96897255fd17
	github.com/btcsuite/btcutil => github.com/keep-network/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/urfave/cli => github.com/keep-network/cli v1.20.0
)

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/blockcypher/gobcy v0.0.0-00010101000000-000000000000
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/ethereum/go-ethereum v1.9.7
	github.com/gogo/protobuf v1.3.1
	github.com/ipfs/go-log v0.0.1
	github.com/keep-network/go-electrum v0.0.0-20190423065222-2dcd82312dcf
	github.com/keep-network/keep-common v0.1.1-0.20191125111950-b9621e71e096
	github.com/keep-network/keep-core v0.6.1-0.20191125124131-866987d36eb5
	github.com/urfave/cli v0.0.0-00010101000000-000000000000
)
