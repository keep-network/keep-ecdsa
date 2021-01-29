const { accounts, contract } = require("@openzeppelin/test-environment")

const KeepToken = contract.fromArtifact("KeepToken")
const StackLib = contract.fromArtifact("StackLib")
const KeepRegistry = contract.fromArtifact("KeepRegistry")
const BondedECDSAKeepFactoryStub = contract.fromArtifact(
  "BondedECDSAKeepFactoryStub"
)
const KeepBonding = contract.fromArtifact("KeepBonding")
const MinimumStakeSchedule = contract.fromArtifact("MinimumStakeSchedule")
const GrantStaking = contract.fromArtifact("GrantStaking")
const Locks = contract.fromArtifact("Locks")
const TopUps = contract.fromArtifact("TopUps")
const TokenStakingEscrow = contract.fromArtifact("TokenStakingEscrow")
const TokenStaking = contract.fromArtifact("TokenStakingStub")
const TokenGrant = contract.fromArtifact("TokenGrant")
const BondedSortitionPoolFactory = contract.fromArtifact(
  "BondedSortitionPoolFactory"
)
const RandomBeaconStub = contract.fromArtifact("RandomBeaconStub")
const BondedECDSAKeepStub = contract.fromArtifact("BondedECDSAKeepStub")
const KeepTokenGrant = contract.fromArtifact("TokenGrant")

async function initialize() {
  const owner = accounts[0]

  await BondedSortitionPoolFactory.detectNetwork()
  await BondedSortitionPoolFactory.link(
    "StackLib",
    (await StackLib.new({ from: owner })).address
  )

  keepToken = await KeepToken.new({ from: owner })
  const keepTokenGrant = await KeepTokenGrant.new(keepToken.address)
  const registry = await KeepRegistry.new({ from: owner })

  const bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new({
    from: owner,
  })
  await TokenStaking.detectNetwork()
  await TokenStaking.link(
    "MinimumStakeSchedule",
    (await MinimumStakeSchedule.new({ from: owner })).address
  )
  await TokenStaking.link(
    "GrantStaking",
    (await GrantStaking.new({ from: owner })).address
  )
  await TokenStaking.link("Locks", (await Locks.new({ from: owner })).address)
  await TokenStaking.link("TopUps", (await TopUps.new({ from: owner })).address)

  const stakingEscrow = await TokenStakingEscrow.new(
    keepToken.address,
    keepTokenGrant.address,
    { from: owner }
  )

  const stakeInitializationPeriod = 30 // In seconds

  tokenStaking = await TokenStaking.new(
    keepToken.address,
    keepTokenGrant.address,
    stakingEscrow.address,
    registry.address,
    stakeInitializationPeriod,
    { from: owner }
  )
  const tokenGrant = await TokenGrant.new(keepToken.address, { from: owner })

  const keepBonding = await KeepBonding.new(
    registry.address,
    tokenStaking.address,
    tokenGrant.address,
    { from: owner }
  )
  const randomBeacon = await RandomBeaconStub.new({ from: owner })
  const bondedECDSAKeepMasterContract = await BondedECDSAKeepStub.new({
    from: owner,
  })
  keepFactory = await BondedECDSAKeepFactoryStub.new(
    bondedECDSAKeepMasterContract.address,
    bondedSortitionPoolFactory.address,
    tokenStaking.address,
    keepBonding.address,
    randomBeacon.address,
    { from: owner }
  )

  await registry.approveOperatorContract(keepFactory.address, { from: owner })

  return {
    tokenStaking: tokenStaking,
    keepToken: keepToken,
    keepFactory: keepFactory,
  }
}

async function fund(keepToken, rewardsContract, amount) {
  const owner = accounts[0]

  await keepToken.approveAndCall(rewardsContract.address, amount, "0x0", {
    from: owner,
  })
  await rewardsContract.markAsFunded({ from: owner })
}

async function createMembers(tokenStaking) {
  const members = []
  const keepSize = 3

  // 3 members in each keep
  for (let i = 0; i < keepSize; i++) {
    const operator = accounts[i]
    const beneficiary = accounts[keepSize + i]
    await tokenStaking.setBeneficiary(operator, beneficiary)
    const member = {
      operator: operator,
      beneficiary: beneficiary,
    }

    members.push(member)
  }

  return members
}

module.exports.initialize = initialize
module.exports.fund = fund
module.exports.createMembers = createMembers
