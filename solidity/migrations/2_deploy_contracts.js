const KeepRegistry = artifacts.require("KeepRegistry")

const KeepBonding = artifacts.require("KeepBonding")
const BondedECDSAKeep = artifacts.require("BondedECDSAKeep")
const BondedECDSAKeepFactory = artifacts.require("BondedECDSAKeepFactory")
const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1 = artifacts.require(
  "BondedECDSAKeepVendorImplV1"
)

const FullyBackedBonding = artifacts.require("FullyBackedBonding")
const FullyBackedECDSAKeep = artifacts.require("FullyBackedECDSAKeep")
const FullyBackedECDSAKeepFactory = artifacts.require(
  "FullyBackedECDSAKeepFactory"
)

const SortitionPoolsDeployer = require("@keep-network/sortition-pools/migrations/scripts/deployContracts")
const BondedSortitionPoolFactory = artifacts.require(
  "BondedSortitionPoolFactory"
)
const FullyBackedSortitionPoolFactory = artifacts.require(
  "FullyBackedSortitionPoolFactory"
)

const LPRewardsTBTCETH = artifacts.require("LPRewardsTBTCETH")
const LPRewardsKEEPETH = artifacts.require("LPRewardsKEEPETH")
const LPRewardsKEEPTBTC = artifacts.require("LPRewardsKEEPTBTC")
const TestToken = artifacts.require("./test/TestToken")
const ECDSARewards = artifacts.require("ECDSARewards")
const ECDSARewardsDistributor = artifacts.require("ECDSARewardsDistributor")

let initializationPeriod = 43200 // 12 hours in seconds

let {
  RandomBeaconAddress,
  TokenStakingAddress,
  TokenGrantAddress,
  RegistryAddress,
  KeepTokenAddress,
} = require("./external-contracts")

module.exports = async function (deployer, network) {
  // Set the stake initialization period to 1 second for local development and testnet.
  if (network === "local" || network === "ropsten" || network === "keep_dev") {
    initializationPeriod = 1
  }

  const sortitionPoolsDeployer = new SortitionPoolsDeployer(deployer, artifacts)
  await sortitionPoolsDeployer.deployBondedSortitionPoolFactory()
  await sortitionPoolsDeployer.deployFullyBackedSortitionPoolFactory()

  if (process.env.TEST) {
    TokenStakingStub = artifacts.require("TokenStakingStub")
    TokenStakingAddress = (await TokenStakingStub.new()).address

    TokenGrantStub = artifacts.require("TokenGrantStub")
    TokenGrantAddress = (await TokenGrantStub.new()).address

    RandomBeaconStub = artifacts.require("RandomBeaconStub")
    RandomBeaconAddress = (await RandomBeaconStub.new()).address

    RegistryAddress = (await deployer.deploy(KeepRegistry)).address
  }

  // KEEP staking and ETH bonding
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

  // ETH bonding only
  await deployer.deploy(
    FullyBackedBonding,
    RegistryAddress,
    initializationPeriod
  )

  await deployer.deploy(FullyBackedECDSAKeep)

  await deployer.deploy(
    FullyBackedECDSAKeepFactory,
    FullyBackedECDSAKeep.address,
    FullyBackedSortitionPoolFactory.address,
    FullyBackedBonding.address,
    RandomBeaconAddress
  )

  // Liquidity Rewards
  const WrappedTokenKEEPETH = await deployer.deploy(TestToken)
  await deployer.deploy(
    LPRewardsKEEPETH,
    KeepTokenAddress,
    WrappedTokenKEEPETH.address
  )

  const WrappedTokenTBTCETH = await deployer.deploy(TestToken)
  await deployer.deploy(
    LPRewardsTBTCETH,
    KeepTokenAddress,
    WrappedTokenTBTCETH.address
  )

  const WrappedTokenKEEPTBTC = await deployer.deploy(TestToken)
  await deployer.deploy(
    LPRewardsKEEPTBTC,
    KeepTokenAddress,
    WrappedTokenKEEPTBTC.address
  )

  // ECDSA Rewards
  await deployer.deploy(
    ECDSARewards,
    KeepTokenAddress,
    BondedECDSAKeepFactory.address,
    TokenStakingAddress
  )

  await deployer.deploy(
    ECDSARewardsDistributor,
    KeepTokenAddress,
    TokenStakingAddress
  )
}
