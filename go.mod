module github.com/keep-network/keep-ecdsa

go 1.13

replace (
	github.com/BurntSushi/toml => github.com/keep-network/toml v0.3.0
	github.com/blockcypher/gobcy => github.com/keep-network/gobcy v1.3.1
	github.com/btcsuite/btcd => github.com/keep-network/btcd v0.0.0-20190427004231-96897255fd17
	github.com/btcsuite/btcutil => github.com/keep-network/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/urfave/cli => github.com/keep-network/cli v1.20.0
)

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/binance-chain/tss-lib v1.3.1
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/ethereum/go-ethereum v1.9.10
	github.com/gogo/protobuf v1.3.1
	github.com/google/gofuzz v1.1.0
	github.com/ipfs/go-log v0.0.1
	github.com/keep-network/keep-common v1.1.1-0.20200703125023-d9872a19ebd1
	github.com/keep-network/keep-core v1.2.4-rc.0.20200730171509-23390fca2250
	github.com/pkg/errors v0.9.1
	github.com/urfave/cli v1.22.1
	go.uber.org/automaxprocs v1.3.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/tools v0.0.0-20200522201501-cb1345f3a375 // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)
