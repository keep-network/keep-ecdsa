# KEEP ECDSA Staking Rewards Claimer: Hardhat CLI Project

This project helps stakers with claiming their KEEP rewards for their work in a single transaction. Normally, a staker needs to claim once per merkle root (i.e. reward period), making it a taxing process in terms of time and in most cases Ether due to tx price volatility.

This tool consists of this README, a copy of ECDSARewardsDistributor's ABI, a `.env` example to be filled out, and the core logic in `hardhat.config.js`. It does not require to deploy any contracts, as it works on top of Keep's Distributor and UniswapV3's implementation of `multicall`.

## Pre-requisites

You need to have the following installed:
- git
- node (tested on v14+)
- yarn (preferably, as it is used through this guide. You can still do everything with just npm + npx)

## "Installing" this tool

1. Clone this repo:
```shell
git clone https://github.com/daramir/keep-ecdsa.git
```
2. Checkout the branch with the tool:
```shell
git checkout feat/ecdsa-rewards-claimer-hardhat-cli
```
3. When in repo root directory, navigate to the sub-directory with the tool:
```shell
cd ./staker-rewards/rewards-claimer-hardhat
```
4. Get all the dependencies with yarn:
```shell
yarn install
```
5. This project uses a fork of `@makerdao/multicall`. You can find the fork version being used as a dependency in [package.json#L10](./package.json#L10). Since that dependency will be cloned, and doesn't come with the `dist` folder, you need to build it as well.
```shell
cd /node_modules/@daramir/multicall
yarn install
yarn build
```
6. Installation is done. Go back to this tool's subdirectory (rewards-claimer-hardhat)


## Usage
```
yarn <TASK> [TASK OPTIONS]

AVAILABLE TASKS:

  accounts              Prints the list of accounts
  balance               Prints an account's balance
  fork                  Starts a JSON-RPC server on top of a Hardhat Network. Will fork network 
                        provided as a param.
  multi-claim           Calls target contract multiple times
  preview-claim         Check the token quantity that would be claimed by the script


To get help for a specific task run: yarn <task> --help
```

### Step by step

1. Copy the file `.env.example` and name the copy `.env`. Fill in the variables with your real values.

  RPC_PROVIDER_URL: Mandatory. Needed to query blockchain state and submit transactions. Has been tested with Infura.

  PRIVATE_KEY_W0: Mandatory. An EOA is needed to sign transactions. This can (and is probably recommended to) be a burner/low-balance wallet. As per ECDSA merkle-distributor rewards design, ANYONE can claim and rewards will ALWAYS go to the beneficiary account. Hence why it is safe to use just any wallet that you're comfortable with extracting it's private key. `.env` is already git-ignored/untracked.
  
  You'll need a balance of around 0.03-0.1 ETH.


2. Preview the expected amount of KEEP to be claimed by running
```shell
yarn preview-claim --op <ecdsa-operator-address>

Output example:
$ hardhat preview-ecdsa-rewards --op 0xffffffffffffffffffffffffffffffffffffffff
Amount of KEEP to claim: 77777.77777770777
Fee data from RPC provider:
maxFeePerGas: xxx 
maxPriorityFeePerGas: y.y
gasPrice: z
Done in 11.98s.
```
3. Send a tx that will process all available* KEEP rewards that have been accrued by the operator in argument. `--priorityfee` and `--maxgasfee` are optional. Run `yarn multi-claim --help` for more info.

```shell
yarn multi-claim --ww 0 --op <ecdsa-operator-address> --priorityfee 1.777 --maxgasfee 99

Output example:
$ hardhat claim-ecdsa-rewards --ww 0 --op <ecdsa-operator-address> --priorityfee 1.70001 --maxgasfee 99
Normalized from address: <address-sending-claim-tx>
Gas estimate for tx is: 722307
Fee data from provider: {
  maxFeePerGas: BigNumber { _hex: '0x25be93214a', _isBigNumber: true },
  maxPriorityFeePerGas: BigNumber { _hex: '0x9502f900', _isBigNumber: true },
  gasPrice: BigNumber { _hex: '0x12d062de25', _isBigNumber: true }
}
Max tx cost estimate: 0.117093984695646942 Ξ
Waiting for tx to be mined ...
```
## Extra functionalities

### Simulate
You can fork mainnet and simulate the claim transaction by:
1. Changing [hardhat.config.js#L312](./hardhat.config.js#L313) from `mainnet` to `localhost`.
2. Running `yarn fork <RPC_PROVIDER_URL>`
3. Going through "Step by step"


## Additional Notes
**available* KEEP rewards**: This program goes by the rewards information available on the local copy of [output-merkle-objects.json](../distributor/output-merkle-objects.json). Feel free to swap it for a newer one, if this branch falls out of date. The latest version should be in the main repo, [https://github.com/keep-network/keep-ecdsa](https://github.com/keep-network/keep-ecdsa/blob/main/staker-rewards/distributor/output-merkle-objects.json).
If `yarn multi-claim` fails to estimate gas, it can be because there are many rewards on the merkle json, that haven't been reflected/provisioned on-chain. In this scenario you can either wait or chop more from [hardhat.config.js#L277](./hardhat.config.js#L277)