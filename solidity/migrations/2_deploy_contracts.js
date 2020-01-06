const KeepBonding = artifacts.require("./KeepBonding.sol");
const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const ECDSAKeepVendor = artifacts.require("./ECDSAKeepVendor.sol");
const KeepRegistry = artifacts.require("./KeepRegistry.sol");

module.exports = async function (deployer) {
    await deployer.deploy(KeepBonding)
    const keepBonding = await KeepBonding.deployed()

    await deployer.deploy(ECDSAKeepFactory, keepBonding.address)
    const ecdsaKeepFactory = await ECDSAKeepFactory.deployed()

    const ecdsaKeepVendor = await deployer.deploy(ECDSAKeepVendor)
    await ecdsaKeepVendor.registerFactory(ecdsaKeepFactory.address)

    const keepRegistry = await deployer.deploy(KeepRegistry)
    await keepRegistry.setVendor('ECDSAKeep', ecdsaKeepVendor.address)
}
