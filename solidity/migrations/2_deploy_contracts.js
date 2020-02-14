const Registry = artifacts.require('Registry')
const KeepBonding = artifacts.require("KeepBonding")
const ECDSAKeepFactory = artifacts.require("ECDSAKeepFactory")
const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1 = artifacts.require("BondedECDSAKeepVendorImplV1")

const { deployBondedSortitionPoolFactory } = require('@keep-network/sortition-pools/migrations/scripts/deployContracts')
const BondedSortitionPoolFactory = artifacts.require("BondedSortitionPoolFactory")

let { RandomBeaconAddress, TokenStakingAddress } = require('./external-contracts')

module.exports = async function (deployer) {
    await deployBondedSortitionPoolFactory(artifacts, deployer)

    let registry
    // TODO: Update with PR206 changes once is merged
    if (process.env.TEST) {
        await deployer.deploy(Registry)
    }

    registry = await Registry.deployed()

    if (process.env.TEST) {
        TokenStakingStub = artifacts.require("TokenStakingStub")
        TokenStakingAddress = (await TokenStakingStub.new()).address

        RandomBeaconStub = artifacts.require("RandomBeaconStub")
        RandomBeaconAddress = (await RandomBeaconStub.new()).address
    }

    await deployer.deploy(KeepBonding, registry.address, TokenStakingAddress)

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
    await vendor.initialize(registry.address)

    await registry.approveOperatorContract(ECDSAKeepFactory.address)
    console.log(`approved operator contract [${ECDSAKeepFactory.address}] in registry`)

    // Set service contract owner as operator contract upgrader by default
    const operatorContractUpgrader = await vendor.owner()
    await registry.setOperatorContractUpgrader(vendor.address, operatorContractUpgrader);
    await vendor.registerFactory(ECDSAKeepFactory.address)
}
