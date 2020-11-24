const artifactsPath = "@keep-network/keep-ecdsa/artifacts"

const clc = require("cli-color");
const truffleContract = require("@truffle/contract")
const BondedECDSAKeepFactoryJson = require(`${artifactsPath}/BondedECDSAKeepFactory.json`)
const BondedECDSAKeepJson = require(`${artifactsPath}/BondedECDSAKeep.json`)

module.exports = async function () {
    try {
        const BondedECDSAKeepFactory = truffleContract(BondedECDSAKeepFactoryJson)
        const BondedECDSAKeep = truffleContract(BondedECDSAKeepJson)

        BondedECDSAKeepFactory.setProvider(web3.currentProvider)
        BondedECDSAKeep.setProvider(web3.currentProvider)

        const factory = await BondedECDSAKeepFactory.deployed()

        const keepCount = await factory.getKeepCount()

        console.log(clc.green(`created keeps count: ${keepCount}`))

        process.exit()
    } catch (error) {
        console.log(error)
    }
}