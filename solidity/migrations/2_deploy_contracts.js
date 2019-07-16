const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const ECDSAKeepVendor = artifacts.require("./ECDSAKeepVendor.sol");
const KeepRegistry = artifacts.require("./KeepRegistry.sol");

module.exports = async function (deployer) {
    await deployer.deploy(ECDSAKeepFactory)
    let ecdsaKeepFactory = await ECDSAKeepFactory.deployed()

    await deployer.deploy(ECDSAKeepVendor)
        .then((instance) => {
            instance.registerFactory(ecdsaKeepFactory.address)
        })
    let ecdsaKeepVendor = await ECDSAKeepVendor.deployed()

    await deployer.deploy(KeepRegistry)
        .then((instance) => {
            instance.setKeepTypeVendor('ECDSAKeep', ecdsaKeepVendor.address)
        })
}
