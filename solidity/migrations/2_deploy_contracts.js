const Registry = artifacts.require('./Registry.sol')
const KeepBonding = artifacts.require("./KeepBonding.sol");
const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const BondedECDSAKeepVendor = artifacts.require("./BondedECDSAKeepVendor.sol");
const BondedECDSAKeepVendorImplV1 = artifacts.require("./BondedECDSAKeepVendorImplV1.sol");

const { deployBondedSortitionPoolFactory } = require('@keep-network/sortition-pools/migrations/scripts/deployContracts')
const BondedSortitionPoolFactory = artifacts.require("BondedSortitionPoolFactory");

module.exports = async function (deployer) {
    await deployer.deploy(Registry)
    await deployer.deploy(KeepBonding)

    await deployBondedSortitionPoolFactory(artifacts, deployer)

    await deployer.deploy(ECDSAKeepFactory, BondedSortitionPoolFactory.address, KeepBonding.address)

    await deployer.deploy(BondedECDSAKeepVendorImplV1)
    await deployer.deploy(BondedECDSAKeepVendor, BondedECDSAKeepVendorImplV1.address)

    const vendor = await BondedECDSAKeepVendorImplV1.at(BondedECDSAKeepVendor.address)
    await vendor.initialize(Registry.address)
    await vendor.registerFactory(ECDSAKeepFactory.address)
}
