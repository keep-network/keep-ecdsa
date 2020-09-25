const KeepRegistry = artifacts.require("KeepRegistry")

const KeepBonding = artifacts.require("KeepBonding")
const BondedECDSAKeep = artifacts.require("BondedECDSAKeep")
const BondedECDSAKeepFactory = artifacts.require("BondedECDSAKeepFactory")
const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1 = artifacts.require(
  "BondedECDSAKeepVendorImplV1"
)

const FullyBackedBonding = artifacts.require("FullyBackedBonding")
const FullyBackedBondedECDSAKeep = artifacts.require(
  "FullyBackedBondedECDSAKeep"
)
const FullyBackedBondedECDSAKeepFactory = artifacts.require(
  "FullyBackedBondedECDSAKeepFactory"
)

const EthBonding = artifacts.require("EthBonding")

const {
  deployBondedSortitionPoolFactory,
  deployFullyBackedSortitionPoolFactory,
} = require("@keep-network/sortition-pools/migrations/scripts/deployContracts")
const BondedSortitionPoolFactory = artifacts.require(
  "BondedSortitionPoolFactory"
)
const FullyBackedSortitionPoolFactory = artifacts.require(
  "FullyBackedSortitionPoolFactory"
)

let {
  RandomBeaconAddress,
  TokenStakingAddress,
  TokenGrantAddress,
  RegistryAddress,
} = require("./external-contracts")

module.exports = async function (deployer) {
  await deployBondedSortitionPoolFactory(artifacts, deployer)
  await deployFullyBackedSortitionPoolFactory(artifacts, deployer)

  if (process.env.TEST) {
    TokenStakingStub = artifacts.require("TokenStakingStub")
    TokenStakingAddress = (await TokenStakingStub.new()).address

    TokenGrantStub = artifacts.require("TokenGrantStub")
    TokenGrantAddress = (await TokenGrantStub.new()).address

    RandomBeaconStub = artifacts.require("RandomBeaconStub")
    RandomBeaconAddress = (await RandomBeaconStub.new()).address

    RegistryAddress = (await deployer.deploy(KeepRegistry)).address
  }

  // KEEP and ETH backed bonding
  await deployer.deploy(
    KeepBonding,
    RegistryAddress,
    TokenStakingAddress,
    TokenGrantAddress
  )

  await deployer.deploy(BondedECDSAKeep)

  await deployer.deploy(
    BondedECDSAKeepFactory,
    BondedECDSAKeep.address,
    BondedSortitionPoolFactory.address,
    TokenStakingAddress,
    KeepBonding.address,
    RandomBeaconAddress
  )

  const bondedECDSAKeepVendorImplV1 = await deployer.deploy(
    BondedECDSAKeepVendorImplV1
  )

  const implInitializeCallData = bondedECDSAKeepVendorImplV1.contract.methods
    .initialize(RegistryAddress, BondedECDSAKeepFactory.address)
    .encodeABI()

  await deployer.deploy(
    BondedECDSAKeepVendor,
    BondedECDSAKeepVendorImplV1.address,
    implInitializeCallData
  )

  // ETH-only Staking
  await deployer.deploy(FullyBackedBonding, RegistryAddress)

  await deployer.deploy(FullyBackedBondedECDSAKeep)

  await deployer.deploy(
    FullyBackedBondedECDSAKeepFactory,
    FullyBackedBondedECDSAKeep.address,
    FullyBackedSortitionPoolFactory.address,
    FullyBackedBonding.address,
    RandomBeaconAddress
  )

  // ETH-only backed bonding
  await deployer.deploy(EthBonding, RegistryAddress)
}
