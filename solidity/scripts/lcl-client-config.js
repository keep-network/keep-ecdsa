/*
This script is used to update client configuration file with latest deployed contracts
addresses.

Example:
KEEP_ECDSA_CONFIG_FILE_PATH=~go/src/github.com/keep-network/keep-ecdsa/configs/config.toml \
    CLIENT_APP_ADDRESS="0x2AA420Af8CB62888ACBD8C7fAd6B4DdcDD89BC82" \
    truffle exec scripts/lcl-client-config.js --network local
*/
const fs = require("fs")
const toml = require("toml")
const tomlify = require("tomlify-j0.4")

const BondedECDSAKeepFactory = artifacts.require("BondedECDSAKeepFactory")
const KeepBonding = artifacts.require("KeepBonding")
const {TokenStakingAddress} = require("../migrations/external-contracts")

module.exports = async function () {
  try {
    const configFilePath = process.env.KEEP_ECDSA_CONFIG_FILE_PATH
    const sanctionedApp = process.env.CLIENT_APP_ADDRESS

    let keepFactoryAddress
    let keepBondingAddress
    try {
      const keepFactory = await BondedECDSAKeepFactory.deployed()
      keepFactoryAddress = keepFactory.address

      const keepBonding = await KeepBonding.deployed()
      keepBondingAddress = keepBonding.address
    } catch (err) {
      console.error("failed to get deployed contracts", err)
      process.exit(1)
    }

    try {
      const fileContent = toml.parse(fs.readFileSync(configFilePath, "utf8"))

      fileContent.ethereum.URL = web3.currentProvider.connection._url

      fileContent.ethereum.ContractAddresses.BondedECDSAKeepFactory = keepFactoryAddress
      fileContent.ethereum.ContractAddresses.KeepBonding = keepBondingAddress
      fileContent.ethereum.ContractAddresses.TokenStaking = TokenStakingAddress

      fileContent.SanctionedApplications.Addresses = [sanctionedApp]

      /*
            tomlify.toToml() writes our Seed/Port values as a float.  The added precision renders our config
            file unreadable by the keep-client as it interprets 3919.0 as a string when it expects an int.
            Here we format the default rendering to write the config file with Seed/Port values as needed.
            */
      const formattedConfigFile = tomlify.toToml(fileContent, {
        space: 2,
        replace: (key, value) => {
          return key == "Port" ? value.toFixed(0) : false
        },
      })

      fs.writeFileSync(configFilePath, formattedConfigFile, (err) => {
        if (err) throw err
      })

      console.log(`keep-ecdsa config written to ${configFilePath}`)
    } catch (err) {
      console.error("failed to update client config", err)
      process.exit(1)
    }
  } catch (err) {
    console.error(err)
    process.exit(1)
  }
  process.exit(0)
}
