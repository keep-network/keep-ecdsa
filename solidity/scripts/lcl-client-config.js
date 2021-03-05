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

module.exports = async function () {
  try {
    const configFilePath = process.env.KEEP_ECDSA_CONFIG_FILE_PATH
    const sanctionedApp = process.env.CLIENT_APP_ADDRESS

    let keepFactoryAddress
    try {
      const keepFactory = await BondedECDSAKeepFactory.deployed()
      keepFactoryAddress = keepFactory.address
    } catch (err) {
      console.error("failed to get deployed contracts", err)
      process.exit(1)
    }

    try {
      const fileContent = toml.parse(fs.readFileSync(configFilePath, "utf8"))

      fileContent.ethereum.URL = web3.currentProvider.connection._url

      fileContent.ethereum.ContractAddresses.BondedECDSAKeepFactory = keepFactoryAddress

      fileContent.ethereum.ContractAddresses.TBTCSystem = sanctionedApp

      /*
            tomlify.toToml() writes our Seed/Port values as a float.  The added precision renders our config
            file unreadable by the keep-client as it interprets 3919.0 as a string when it expects an int.
            Here we format the default rendering to write the config file with Seed/Port values as needed.
            */
      const formattedConfigFile = tomlify.toToml(fileContent, {
        space: 2,
        replace: (key, value) => {
          let result
          try {
            result =
              // We expect the config file to contain arrays, in such case key for
              // each entry is its' index number. We verify if the key is a string
              // so we can run the following match check.
              typeof key === "string" &&
              // Find keys that match exactly `Port`, `MiningCheckInterval`,
              // `MaxGasPrice` or end with `MetricsTick` or `Limit`.
              key.match(
                /(^Port|^MiningCheckInterval|^MaxGasPrice|MetricsTick|Limit)$/
              )
                ? value.toFixed(0) // convert float to integer
                : false // do nothing
          } catch (err) {
            console.error(
              `tomlify replace failed for key ${key} and value ${value} with error: [${err}]`
            )
            process.exit(1)
          }

          return result
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
