/*
This script is used to update client configuration file with latest deployed contracts
addresses.

Example:
KEEP_ECDSA_CONFIG_FILE_PATH=~go/src/github.com/keep-network/keep-ecdsa/configs/config.toml \
    CLIENT_APP_ADDRESS="0x2AA420Af8CB62888ACBD8C7fAd6B4DdcDD89BC82" \
    truffle exec scripts/lcl-client-config.js --network local
*/
const fs = require('fs')
const toml = require('toml')
const tomlify = require('tomlify-j0.4')

const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory')

module.exports = async function () {
    try {
        const configFilePath = process.env.KEEP_ECDSA_CONFIG_FILE_PATH
        const sanctionedApp = process.env.CLIENT_APP_ADDRESS

        let keepFactoryAddress
        try {
            const keepFactory = await ECDSAKeepFactory.deployed()
            keepFactoryAddress = keepFactory.address
        } catch (err) {
            console.error('failed to get deployed contracts', err)
            process.exit(1)
        }

        try {
            const fileContent = toml.parse(fs.readFileSync(configFilePath, 'utf8'))

            fileContent.ethereum.URL = web3.currentProvider.connection._url

            fileContent.ethereum.ContractAddresses.ECDSAKeepFactory = keepFactoryAddress

            fileContent.SanctionedApplications.Addresses = [sanctionedApp]

            fs.writeFileSync(configFilePath, tomlify.toToml(fileContent), (err) => {
                if (err) throw err
            })

            console.log(`keep-ecdsa config written to ${configFilePath}`)
        } catch (err) {
            console.error('failed to update client config', err)
            process.exit(1)
        }

    } catch (err) {
        console.error(err)
        process.exit(1)
    }
    process.exit(0)
}
