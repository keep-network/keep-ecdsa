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
	github.com/aristanetworks/goarista v0.0.0-20191206003309-5d8d36c240c9 // indirect
	github.com/binance-chain/tss-lib v1.1.1
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/ethereum/go-ethereum v1.9.7
	github.com/gogo/protobuf v1.3.1
	github.com/ipfs/go-log v0.0.1
	github.com/keep-network/keep-common v0.1.1-0.20191203134929-648c427de66e
	github.com/keep-network/keep-core v0.7.0
	github.com/libp2p/go-libp2p-core v0.2.5
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/pkg/errors v0.8.1
	github.com/urfave/cli v0.0.0-00010101000000-000000000000
)
