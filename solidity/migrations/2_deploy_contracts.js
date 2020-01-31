const KeepBonding = artifacts.require("./KeepBonding.sol");
const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const BondedECDSAKeepVendor = artifacts.require("./BondedECDSAKeepVendor.sol");
const BondedECDSAKeepVendorImplV1 = artifacts.require("./BondedECDSAKeepVendorImplV1.sol");

const deploySortitionPoolFactory = require('@keep-network/sortition-pools/migrations/scripts/deployContracts')
const SortitionPoolFactory = artifacts.require("SortitionPoolFactory");

module.exports = async function (deployer) {
    await deployer.deploy(KeepBonding)

    await deploySortitionPoolFactory(artifacts, deployer)

    await deployer.deploy(ECDSAKeepFactory, SortitionPoolFactory.address, KeepBonding.address)

    await deployer.deploy(BondedECDSAKeepVendorImplV1)
    await deployer.deploy(BondedECDSAKeepVendor, BondedECDSAKeepVendorImplV1.address)

    const vendor = await BondedECDSAKeepVendorImplV1.at(BondedECDSAKeepVendor.address)
    await vendor.registerFactory(ECDSAKeepFactory.address)
}
