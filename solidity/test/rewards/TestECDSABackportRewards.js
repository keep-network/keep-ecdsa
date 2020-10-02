const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("../helpers/snapshot")

const {expectRevert, time} = require("@openzeppelin/test-helpers")

const {mineBlocks} = require("../helpers/mineBlocks")

const truffleAssert = require("truffle-assertions")

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
const TokenStaking = contract.fromArtifact("TokenStaking")
const TokenGrant = contract.fromArtifact("TokenGrant")
const BondedSortitionPool = contract.fromArtifact("BondedSortitionPool")
const BondedSortitionPoolFactory = contract.fromArtifact(
  "BondedSortitionPoolFactory"
)
const RandomBeaconStub = contract.fromArtifact("RandomBeaconStub")
const BondedECDSAKeep = contract.fromArtifact("BondedECDSAKeep")
const ECDSABackportRewards = contract.fromArtifact("ECDSABackportRewards")
const KeepToken = contract.fromArtifact("KeepToken")
const KeepTokenGrant = contract.fromArtifact("TokenGrant")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect
const assert = chai.assert

describe("ECDSABackportRewards", () => {
  let registry
  let rewardsContract
  let keepToken

  let tokenStaking
  let tokenGrant
  let keepFactory
  let bondedSortitionPoolFactory
  let keepBonding
  let randomBeacon

  const owner = accounts[0]

  const keepSize = 16
  const intervalWeights = [100]

  // BondedECDSAKeepFactory deployment date
  const initiationTime = 1589408351

  // 1,000,000,000 - total KEEP supply
  //   200,000,000 - 20% of the total supply goes to staker rewards
  //   180,000,000 - 90% of staker rewards goes to the ECDSA stakers
  //     1,800,000 - 1% of ECDSA staker rewards goes to May - Sep keeps
  const ECDSABackportKEEPRewards = 1800000

  async function createStakedMembers() {
    const minimumStake = await tokenStaking.minimumStake.call()

    console.log("mnimum stake", minimumStake.toString())
    const members = []

    // 16 members in each keep
    for (let i = 0; i < keepSize; i++) {
      const operator = accounts[i]
      const beneficiary = accounts[keepSize + i]
      const authorizer = accounts[i]

      const delegation = Buffer.concat([
        Buffer.from(web3.utils.hexToBytes(beneficiary)),
        Buffer.from(web3.utils.hexToBytes(operator)),
        Buffer.from(web3.utils.hexToBytes(authorizer)),
      ])

      console.log("keepToken.address1 ", keepToken.address)
      await keepToken.approveAndCall(
        tokenStaking.address,
        minimumStake,
        delegation
      )
      members.push(accounts[i])
    }

    return members
  }

  async function initializeNewFactory() {
    keepToken = await KeepToken.new()
    const keepTokenGrant = await KeepTokenGrant.new(keepToken.address)
    registry = await KeepRegistry.new()

    bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
    await TokenStaking.detectNetwork()
    await TokenStaking.link(
      "MinimumStakeSchedule",
      (await MinimumStakeSchedule.new()).address
    )
    await TokenStaking.link("GrantStaking", (await GrantStaking.new()).address)
    await TokenStaking.link("Locks", (await Locks.new()).address)
    await TokenStaking.link("TopUps", (await TopUps.new()).address)

    const stakingEscrow = await TokenStakingEscrow.new(
      keepToken.address,
      keepTokenGrant.address
    )

    const stakeInitializationPeriod = 30 // In seconds

    tokenStaking = await TokenStaking.new(
      keepToken.address,
      keepTokenGrant.address,
      stakingEscrow.address,
      registry.address,
      stakeInitializationPeriod
    )
    tokenGrant = await TokenGrant.new(keepToken.address)

    keepBonding = await KeepBonding.new(
      registry.address,
      tokenStaking.address,
      tokenGrant.address
    )
    randomBeacon = await RandomBeaconStub.new()
    const bondedECDSAKeepMasterContract = await BondedECDSAKeep.new()
    keepFactory = await BondedECDSAKeepFactoryStub.new(
      bondedECDSAKeepMasterContract.address,
      bondedSortitionPoolFactory.address,
      tokenStaking.address,
      keepBonding.address,
      randomBeacon.address
    )

    await registry.approveOperatorContract(keepFactory.address)
  }

  async function createKeeps(timestamps) {
    await keepFactory.openSyntheticKeeps([alice, bob], timestamps)
    for (let i = 0; i < timestamps.length; i++) {
      const keepAddress = await keepFactory.getKeepAtIndex(i)
      const keep = await RewardsKeepStub.at(keepAddress)
      await keep.setTimestamp(timestamps[i])
    }
  }

  before(async () => {
    await BondedSortitionPoolFactory.detectNetwork()
    await BondedSortitionPoolFactory.link(
      "StackLib",
      (await StackLib.new()).address
    )

    await initializeNewFactory()
    await createStakedMembers()
    rewardsContract = await ECDSABackportRewards.new(
      tokenStaking.address,
      keepFactory.address
    )

    await fund(ECDSABackportKEEPRewards)
  })

  async function fund(amount) {
    // await keepToken.mint(owner, amount)
    console.log("keepToken.address2 ", keepToken.address)
    await keepToken.approveAndCall(rewardsContract.address, amount, "0x0", {
      from: owner,
    })
  }

  async function timeJumpToEndOfInterval(intervalNumber) {
    const endOf = await rewardsContract.endOf(intervalNumber)
    const now = await time.latest()

    if (now.lt(endOf)) {
      await time.increaseTo(endOf.addn(1))
    }
  }

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe.only("interval allocation", async () => {
    it("should equal the full allocation", async () => {
      const expectedAllocation = 1800000

      await timeJumpToEndOfInterval(0)
      await rewardsContract.allocateRewards(0)

      const allocated = await rewardsContract.getAllocatedRewards(0)
      //   const allocatedKeep = allocated.div(tokenDecimalMultiplier)

      //   expect(allocatedKeep).to.eq.BN(expectedAllocation)
      expect(allocated).to.eq.BN(expectedAllocation)
    })
  })
})
