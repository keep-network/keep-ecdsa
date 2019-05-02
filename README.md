# keep-tecdsa

## Submit transaction

### Electrum Server

Transaction can be published to the blockchain via Electrum Server.

#### Configuration

To configure connection details of Electrum Server update [config.toml](configs/config.toml) file. 

`configs/config.toml` is default path to the config file. To provide custom configuration CLI supports `--config` flag.

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
./keep-tecdsa publish <rawTx>
``` 
##### Example
###### Input
```sh
./keep-tecdsa publish 02000000000101506fda83a9788dab896b90bfc122c700afccd30459e10a0ff270e951202612481600000017160014997f3e8bcf47183fbbe0c8464175047bf391ea83feffffff0279d513000000000016001432b027edb95eee83b40003762a2ff25ae47d560d40420f000000000016001469abce7925fce369303e247c3a465447f6519b780247304402205be426e3e0c2e243eac4808cca306e3f821f7e431e2f0fc0c534851d5551779402202e98a4c7899d0eb2f473af77af6b434a5a47a0d898ed5b41ca2a7692601e36eb012102cb34ff4b355a7f02104cb912dbf4e35a14733030450ac311f6ef498d150a6ad2eb171700
```
###### Output
```sh
Connected to Electrum Server.
Server version: ElectrumX 1.11.0 [Protocol 1.4]
2019/05/01 15:47:01 publish failed [transaction broadcast failed [errNo: 1, errMsg: the transaction was rejected by network rules.

transaction already in block chain
[02000000000101506fda83a9788dab896b90bfc122c700afccd30459e10a0ff270e951202612481600000017160014997f3e8bcf47183fbbe0c8464175047bf391ea83feffffff0279d513000000000016001432b027edb95eee83b40003762a2ff25ae47d560d40420f000000000016001469abce7925fce369303e247c3a465447f6519b780247304402205be426e3e0c2e243eac4808cca306e3f821f7e431e2f0fc0c534851d5551779402202e98a4c7899d0eb2f473af77af6b434a5a47a0d898ed5b41ca2a7692601e36eb012102cb34ff4b355a7f02104cb912dbf4e35a14733030450ac311f6ef498d150a6ad2eb171700]]]
```
