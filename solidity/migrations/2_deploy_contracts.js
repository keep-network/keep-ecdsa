const KeepBonding = artifacts.require("./KeepBonding.sol");
const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const ECDSAKeepVendor = artifacts.require("./ECDSAKeepVendor.sol");
const KeepRegistry = artifacts.require("./KeepRegistry.sol");

const deploySortitionPoolFactory = require('@keep-network/sortition-pools/migrations/scripts/deployContracts')
const SortitionPoolFactory = artifacts.require("SortitionPoolFactory");

module.exports = async function (deployer) {
    await deployer.deploy(KeepBonding)

    await deploySortitionPoolFactory(artifacts, deployer)

    await deployer.deploy(ECDSAKeepFactory, SortitionPoolFactory.address)
    const ecdsaKeepFactory = await ECDSAKeepFactory.deployed()

    const ecdsaKeepVendor = await deployer.deploy(ECDSAKeepVendor)
    await ecdsaKeepVendor.registerFactory(ecdsaKeepFactory.address)

    const keepRegistry = await deployer.deploy(KeepRegistry)
    await keepRegistry.setVendor('ECDSAKeep', ecdsaKeepVendor.address)
}
