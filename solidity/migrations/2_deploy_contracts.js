const TECDSAKeepFactory = artifacts.require("./TECDSAKeepFactory.sol");
const TECDSAKeepVendor = artifacts.require("./TECDSAKeepVendor.sol");
const KeepRegistry = artifacts.require("./KeepRegistry.sol");

module.exports = async function (deployer) {
    await deployer.deploy(TECDSAKeepFactory)
    const tecdsaKeepFactory = await TECDSAKeepFactory.deployed()

    const tecdsaKeepVendor = await deployer.deploy(TECDSAKeepVendor)
    await tecdsaKeepVendor.registerFactory(tecdsaKeepFactory.address)

    const keepRegistry = await deployer.deploy(KeepRegistry)
    await keepRegistry.setVendor('TECDSAKeep', tecdsaKeepVendor.address)
}
