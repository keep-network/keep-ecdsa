const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("../helpers/snapshot")
const {initialize, fund, createMembers} = require("./rewardsSetup")

const {expectRevert, time} = require("@openzeppelin/test-helpers")

const BondedECDSAKeepStub = contract.fromArtifact("BondedECDSAKeepStub")
const ECDSARewards = contract.fromArtifact("ECDSARewards")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

describe.only("ECDSARewards", () => {
  let rewardsContract
  let keepToken

  let tokenStaking
  let keepFactory
  let keepMembers

  const expectedIntervalAllocations = [
    7128000.0, 13685760.0, 15738624.0, 16997713.92, 18697485.31, 15892862.52,
    13508933.14, 11482593.17, 9760204.19, 8296173.56, 7051747.53, 5993985.4,
    5094887.59, 4330654.45, 3681056.28, 3128897.84, 2659563.16, 2260628.69,
    1921534.39, 1633304.23, 1388308.59, 1180062.31, 1003052.96, 852595.02
  ]

  const owner = accounts[0]

  // Solidity is not very good when it comes to floating point precision,
  // we are allowing for ~1 KEEP difference margin between expected and
  // actual value.
  const precision = 1
  const tokenDecimalMultiplier = web3.utils.toBN(10).pow(web3.utils.toBN(18))
  const firstKeepCreationTimestamp = 1600041600 // Sep 14 2020

  // 1,000,000,000 - total KEEP supply
  //   200,000,000 - 20% of the total supply goes to staker rewards
  //   180,000,000 - 90% of staker rewards goes to the ECDSA staker rewards
  //   178,200,000 - 99% of ECDSA staker rewards goes to keeps opened after Sep 14 2020
  const totalRewardsAllocation = web3.utils.toBN(178200000).mul(tokenDecimalMultiplier)

  before(async () => {
    const setup = await initialize()
    tokenStaking = setup.tokenStaking
    keepToken = setup.keepToken
    keepFactory = setup.keepFactory

    keepMembers = await createMembers(tokenStaking)
    rewardsContract = await ECDSARewards.new(
      keepToken.address,
      keepFactory.address,
      {from: owner}
    )

    await fund(keepToken, rewardsContract, totalRewardsAllocation)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("interval allocation", async () => {
    it("should equal to expected allocations when 5 keeps were created per interval", async () => {
      await verifyIntervalAllocations(5)
    })

    it("should equal to expected allocations when 1 keep was created per interval", async () => {
      await verifyIntervalAllocations(1)
    })
  })

  async function verifyIntervalAllocations(keepToCreatePerInterval) {
    let keepCreationTimestamp = firstKeepCreationTimestamp

    for (let i = 0; i < 24; i++) {
      await createKeeps(keepToCreatePerInterval, keepCreationTimestamp)

      keepCreationTimestamp = await timeJumpToEndOfInterval(i)

      await rewardsContract.allocateRewards(i)

      const actualBalance = (await rewardsContract.getAllocatedRewards(i)).div(
        tokenDecimalMultiplier
      )

      expect(actualBalance).to.gte.BN(expectedIntervalAllocations - precision)
      expect(actualBalance).to.lte.BN(expectedIntervalAllocations + precision)
    }
  }

  describe("rewards distribution", async () => {
    it("should not be possible when a keep is not closed", async () => {
      await createKeeps(1, firstKeepCreationTimestamp)

      const keepAddress = await keepFactory.getKeepAtIndex(0)

      const isEligible = await rewardsContract.eligibleForReward(keepAddress)
      expect(isEligible).to.be.false

      await timeJumpToEndOfInterval(0)
      await expectRevert(
        rewardsContract.receiveReward(keepAddress),
        "Keep is not closed"
      )
    })

    it("should not be possible when a keep is terminated", async () => {
      await createKeeps(1, firstKeepCreationTimestamp)

      const keepAddress = await keepFactory.getKeepAtIndex(0)
      const keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsTerminated()

      const isEligible = await rewardsContract.eligibleForReward(keepAddress)
      expect(isEligible).to.be.false

      await timeJumpToEndOfInterval(0)
      await expectRevert(
        rewardsContract.receiveReward(keepAddress),
        "Keep is not closed"
      )
    })

    it("should not count terminated groups when distributing rewards", async () => {
      await createKeeps(2, firstKeepCreationTimestamp)
      // reward for the first interval: 7128000 KEEP
      // keeps created: 2, but the min is 4 => 7128000 / 4 = 1782000 KEEP per keep
      // member receives: 1782000 / 3 = 594000 (3 signers per keep)
      const expectedBeneficiaryBalance = new BN(594000)

      await timeJumpToEndOfInterval(0)

      const keepTerminatedAddress = await keepFactory.getKeepAtIndex(0)
      const keepTerminated = await BondedECDSAKeepStub.at(keepTerminatedAddress)
      await keepTerminated.publicMarkAsTerminated()

      const keepAddress = await keepFactory.getKeepAtIndex(1)
      const keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await expectRevert(
        rewardsContract.receiveReward(keepTerminatedAddress),
        "Keep is not closed"
      )

      await rewardsContract.receiveReward(keepAddress)
      // Only the second keep was properly closed.
      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)

      // the remaining 5346000 stays in unallocated rewards but the fact
      // one keep was terminated needs to be reported to recalculate the
      // unallocated amount
      let unallocated = await rewardsContract.unallocatedRewards()
      let unallocatedInKeep = unallocated.div(tokenDecimalMultiplier)
      expect(unallocatedInKeep).to.eq.BN(174636000) // 178200000 - (2 * 1782000)

      await rewardsContract.reportTermination(keepTerminatedAddress)

      unallocated = await rewardsContract.unallocatedRewards()
      unallocatedInKeep = unallocated.div(tokenDecimalMultiplier)
      expect(unallocatedInKeep).to.eq.BN(176418000) // 178200000 - 1782000
    })

    it("should correctly distribute rewards between beneficiaries", async () => {
      await createKeeps(8, firstKeepCreationTimestamp)
      // reward for the first interval: 7128000 KEEP
      // keeps created: 8 => 891000 KEEP per keep
      // member receives: 891000 / 3 = 297000 (3 signers per keep)
      const expectedBeneficiaryBalance = new BN(297000)

      await timeJumpToEndOfInterval(0)

      let keepAddress = await keepFactory.getKeepAtIndex(0)
      let keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress)

      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)

      // verify second keep in this interval
      keepAddress = await keepFactory.getKeepAtIndex(1)
      keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress)

      // 297000 * 2 = 594000
      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance.muln(2))
    })
  })

  async function assertKeepBalanceOfBeneficiaries(expectedBalance) {
    for (let i = 0; i < keepMembers.length; i++) {
      const actualBalance = (
        await keepToken.balanceOf(keepMembers[i].beneficiary)
      ).div(tokenDecimalMultiplier)

      expect(actualBalance).to.gte.BN(expectedBalance.subn(precision))
      expect(actualBalance).to.lte.BN(expectedBalance.addn(precision))
    }
  }

  async function createKeeps(numberOfKeepsToOpen, keepCreationTimestamp) {
    let timestamp = new BN(keepCreationTimestamp)
    const members = keepMembers.map((m) => m.operator)
    for (let i = 0; i < numberOfKeepsToOpen; i++) {
      await keepFactory.stubOpenKeep(keepFactory.address, members, timestamp)
      timestamp = timestamp.addn(7200) // adding 2 hours interval between each opened keep
    }
  }

  async function timeJumpToEndOfInterval(intervalNumber) {
    const endOf = await rewardsContract.endOf(intervalNumber)
    const now = await time.latest()

    if (now.lt(endOf)) {
      await time.increaseTo(endOf.addn(60))
    }

    return await time.latest()
  }
})
