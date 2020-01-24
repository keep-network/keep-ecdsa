const KeepBonding = artifacts.require("./KeepBonding.sol");
const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const ECDSAKeepVendor = artifacts.require("./ECDSAKeepVendor.sol");
const KeepRegistry = artifacts.require("./KeepRegistry.sol");

const SortitionPoolFactoryStub = artifacts.require("./SortitionPoolFactoryStub");

let { SortitionPoolFactoryAddress } = require('./externals')

module.exports = async function (deployer) {
    await deployer.deploy(KeepBonding)

    if (process.env.TEST = true) {
        SortitionPoolFactoryAddress = (await deployer.deploy(SortitionPoolFactoryStub)).address
    }

    await deployer.deploy(ECDSAKeepFactory, SortitionPoolFactoryAddress)
    const ecdsaKeepFactory = await ECDSAKeepFactory.deployed()

    const ecdsaKeepVendor = await deployer.deploy(ECDSAKeepVendor)
    await ecdsaKeepVendor.registerFactory(ecdsaKeepFactory.address)

    const keepRegistry = await deployer.deploy(KeepRegistry)
    await keepRegistry.setVendor('ECDSAKeep', ecdsaKeepVendor.address)
}
