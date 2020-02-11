const KeepBonding = artifacts.require("./KeepBonding.sol");
const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const BondedECDSAKeepVendor = artifacts.require("./BondedECDSAKeepVendor.sol");
const BondedECDSAKeepVendorImplV1 = artifacts.require("./BondedECDSAKeepVendorImplV1.sol");

const { deployBondedSortitionPoolFactory } = require('@keep-network/sortition-pools/migrations/scripts/deployContracts')
const BondedSortitionPoolFactory = artifacts.require("BondedSortitionPoolFactory");

// TokenStaking artifact is expected to be copied over from previous keep-core
// migrations.
let TokenStaking;

const { RandomBeaconAddress } = require('./externals')

module.exports = async function (deployer) {
    await deployer.deploy(KeepBonding)

    await deployBondedSortitionPoolFactory(artifacts, deployer)

    if (process.env.TEST) {
        TokenStakingStub = artifacts.require("TokenStakingStub")
        TokenStaking = await TokenStakingStub.new()
    } else {
        TokenStaking = artifacts.require("TokenStaking")
    }

    await deployer.deploy(
        ECDSAKeepFactory,
        BondedSortitionPoolFactory.address,
        TokenStaking.address,
        KeepBonding.address,
        RandomBeaconAddress
    )

    await deployer.deploy(BondedECDSAKeepVendorImplV1)
    await deployer.deploy(BondedECDSAKeepVendor, BondedECDSAKeepVendorImplV1.address)

    const vendor = await BondedECDSAKeepVendorImplV1.at(BondedECDSAKeepVendor.address)
    await vendor.registerFactory(ECDSAKeepFactory.address)
}
