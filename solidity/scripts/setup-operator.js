// This script sets up operator for running the client. It updates the client
// configuration file with the address of ECDSAKeepFactory contract.
// It reads operator's account address from the path specified in client's
// configuration file and deposits the account with 10 ETH bonding value.
// The script requires following properties:
//   - `KEEP_ETHEREUM_PASSWORD` environment variable set to a password to the 
//     ethereum account,
//   - CONFIG_FILE_PATH - process argument set to a path to the client's config file.
// 
// To execute this script run:
//   KEEP_ETHEREUM_PASSWORD=password truffle exec scripts/setup-client.js ../configs/config.toml

const fs = require('fs')
const toml = require('toml')
const tomlify = require('tomlify-j0.4')
const concat = require('concat-stream')

const KeepBonding = artifacts.require('./KeepBonding.sol')
const ECDSAKeepFactory = artifacts.require('./ECDSAKeepFactory.sol')

const { depositBondingValue } = require("./helpers/bonding")

module.exports = async function () {
    const ethAccountPassword = process.env.KEEP_ETHEREUM_PASSWORD
    const configFilePath = process.argv[4]

    try {
        const ecdsaKeepFactory = await ECDSAKeepFactory.deployed()
        await updateKeepClientConfig(configFilePath, ecdsaKeepFactory.address)

        const operatorAccount = readOperatorAccount(configFilePath, ethAccountPassword)

        const keepBonding = await KeepBonding.deployed()

        const keepBondingContract = new web3.eth.Contract(keepBonding.abi, keepBonding.address)

        const purse = (await web3.eth.getAccounts())[0]
        await depositBondingValue(keepBondingContract, purse, operatorAccount, '10')
    } catch (err) {
        console.error(err)
        process.exit(1)
    }

    process.exit(0)
}

async function updateKeepClientConfig(configFilePath, ecdsaKeepFactoryAddress) {
    console.log(`update client config file`)
    console.debug(`set ECDSAKeepFactory address to: [${ecdsaKeepFactoryAddress}]`)

    fs.createReadStream(configFilePath, 'utf8').pipe(concat(function (data) {
        let parsedConfigFile = toml.parse(data)

        parsedConfigFile.ethereum.ContractAddresses.ECDSAKeepFactory = ecdsaKeepFactoryAddress

        let formattedConfigFile = tomlify.toToml(parsedConfigFile)

        fs.writeFile(configFilePath, formattedConfigFile, (error) => {
            if (error) throw error
        })
    }))
    console.log(`client config written to: [${configFilePath}]`)
}

function readOperatorAccount(configFilePath, ethAccountPassword) {
    const parsedConfigFile = toml.parse(fs.readFileSync(configFilePath).toString())

    const keyfile = parsedConfigFile.ethereum.account.KeyFile

    console.debug(`read operator account from file: [${keyfile}]`)

    const keystoreJSON = JSON.parse(fs.readFileSync(keyfile))

    const account = web3.eth.accounts.decrypt(keystoreJSON, ethAccountPassword)

    console.log(`operator account address: [${account.address}]`)

    return account.address
}
