const KeepBonding = artifacts.require("KeepBonding")
const ECDSAKeepFactory = artifacts.require("ECDSAKeepFactory")
const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1 = artifacts.require("BondedECDSAKeepVendorImplV1")

const { deployBondedSortitionPoolFactory } = require('@keep-network/sortition-pools/migrations/scripts/deployContracts')
const BondedSortitionPoolFactory = artifacts.require("BondedSortitionPoolFactory")

let { RandomBeaconAddress, TokenStakingAddress } = require('./external-contracts')

module.exports = async function (deployer) {
    await deployer.deploy(KeepBonding)

    await deployBondedSortitionPoolFactory(artifacts, deployer)

    if (process.env.TEST) {
        TokenStakingStub = artifacts.require("TokenStakingStub")
        TokenStakingAddress = (await TokenStakingStub.new()).address

        RandomBeaconStub = artifacts.require("RandomBeaconStub")
        RandomBeaconAddress = (await RandomBeaconStub.new()).address
    }

    await deployer.deploy(
        ECDSAKeepFactory,
        BondedSortitionPoolFactory.address,
        TokenStakingAddress,
        KeepBonding.address,
        RandomBeaconAddress
    )

    await deployer.deploy(BondedECDSAKeepVendorImplV1)
    await deployer.deploy(BondedECDSAKeepVendor, BondedECDSAKeepVendorImplV1.address)

    const vendor = await BondedECDSAKeepVendorImplV1.at(BondedECDSAKeepVendor.address)
    await vendor.registerFactory(ECDSAKeepFactory.address)
}
