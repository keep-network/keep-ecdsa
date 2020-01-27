const KeepBonding = artifacts.require("./KeepBonding.sol");
const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const ECDSAKeepVendor = artifacts.require("./ECDSAKeepVendor.sol");
const KeepRegistry = artifacts.require("./KeepRegistry.sol");

const external = require("./external")

module.exports = async function (deployer) {
    let SortitionPoolFactory

    await deployer.deploy(KeepBonding)

    if (process.env.TEST) {
        const SortitionPoolFactoryStub = artifacts.require("./SortitionPoolFactoryStub");
        SortitionPoolFactory = (await deployer.deploy(SortitionPoolFactoryStub))
    } else {
        SortitionPoolFactory = external.SortitionPoolFactory
        SortitionPoolFactory.setProvider(web3.eth.currentProvider)
        await SortitionPoolFactory.deployed()
    }

    await deployer.deploy(ECDSAKeepFactory, SortitionPoolFactory.address)

    const ecdsaKeepVendor = await deployer.deploy(ECDSAKeepVendor)
    await ecdsaKeepVendor.registerFactory(ECDSAKeepFactory.address)

    const keepRegistry = await deployer.deploy(KeepRegistry)
    await keepRegistry.setVendor('ECDSAKeep', ECDSAKeepVendor.address)
}
