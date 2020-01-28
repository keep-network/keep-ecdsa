const KeepBonding = artifacts.require("./KeepBonding.sol");
const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const ECDSAKeepVendor = artifacts.require("./ECDSAKeepVendor.sol");
const ECDSAKeepVendorImplV1 = artifacts.require("./ECDSAKeepVendorImplV1.sol");

const deploySortitionPoolFactory = require('@keep-network/sortition-pools/migrations/scripts/deployContracts')
const SortitionPoolFactory = artifacts.require("SortitionPoolFactory");

module.exports = async function (deployer) {
    await deployer.deploy(KeepBonding)

    await deploySortitionPoolFactory(artifacts, deployer)

    await deployer.deploy(ECDSAKeepFactory, SortitionPoolFactory.address)
    const ecdsaKeepFactory = await ECDSAKeepFactory.deployed()

    const ecdsaKeepVendor = await deployer.deploy(ECDSAKeepVendorImplV1);    
    await deployer.deploy(ECDSAKeepVendor, ecdsaKeepVendor.address)
    await ecdsaKeepVendor.registerFactory(ecdsaKeepFactory.address)
}
