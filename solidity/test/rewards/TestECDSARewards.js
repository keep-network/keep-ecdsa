const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {expectRevert, time} = require("@openzeppelin/test-helpers")

const BondedECDSAKeepStub = contract.fromArtifact("BondedECDSAKeepStub")
const ECDSARewards = contract.fromArtifact("ECDSARewards")
const {createSnapshot, restoreSnapshot} = require("../helpers/snapshot")
const {initialize, fund, createMembers} = require("./rewardsSetup")

const BN = web3.utils.BN
const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

describe("ECDSARewards", () => {
  let keepToken
  let tokenStaking
  let keepFactory
  let rewardsContract

  const owner = accounts[0]
  let operators
  let beneficiaries

  const tokenDecimalMultiplier = web3.utils.toBN(10).pow(web3.utils.toBN(18))
  const firstIntervalStart = 1600041600 // Sep 14 2020

  // 1,000,000,000 - total KEEP supply
  //   200,000,000 - 20% of the total supply goes to staker rewards
  //   180,000,000 - 90% of staker rewards goes to the ECDSA staker rewards
  //   178,200,000 - 99% of ECDSA staker rewards goes to keeps opened after Sep 14 2020
  const totalRewardsAllocation = web3.utils
    .toBN(178200000)
    .mul(tokenDecimalMultiplier)

  before(async () => {
    const contracts = await initialize()
    tokenStaking = contracts.tokenStaking
    keepToken = contracts.keepToken
    keepFactory = contracts.keepFactory

    stakers = await createMembers(tokenStaking)
    operators = stakers.map((s) => s.operator)
    beneficiaries = stakers.map((s) => s.beneficiary)

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
    it("should equal expected allocation in first interval", async () => {
      await keepFactory.stubBatchOpenFakeKeeps(100, firstIntervalStart)

      await timeJumpToEndOfInterval(0)
      await rewardsContract.allocateRewards(0)
      const allocated = await rewardsContract.getAllocatedRewards(0)
      const allocatedKeep = allocated.div(tokenDecimalMultiplier)

      // Full allocation for the first interval would be
      // 178,200,000 * 4% = 7,128,000.
      // Because just 10% of minimum keep quota is met, the allocation is
      // 7,128,000 * 10% = 712,800.
      expect(allocatedKeep).to.eq.BN(712800)
    })

    it("should equal expected allocation in second interval", async () => {
      await keepFactory.stubBatchOpenFakeKeeps(100, firstIntervalStart)
      const firstIntervalEnd = await timeJumpToEndOfInterval(0)
      await keepFactory.stubBatchOpenFakeKeeps(50, firstIntervalEnd)
      await timeJumpToEndOfInterval(1)

      await rewardsContract.allocateRewards(1)
      const allocated = await rewardsContract.getAllocatedRewards(1)
      const allocatedKeep = allocated.div(tokenDecimalMultiplier)

      // 712,800 allocated in the first interval,
      // Full allocation for second interval would be
      // (178,200,000 - 712,800) * 8% = 14,198,976.
      // Because just 5% of minimum keep quota is met, the allocation is
      // 14,198,976 * 5% = 709,948.
      expect(allocatedKeep).to.eq.BN(709948)
    })

    it("should equal expected allocation in third interval", async () => {
      const secondIntervalEnd = await timeJumpToEndOfInterval(1)
      await keepFactory.stubBatchOpenFakeKeeps(40, secondIntervalEnd)
      await timeJumpToEndOfInterval(2)

      await rewardsContract.allocateRewards(2)
      const allocated = await rewardsContract.getAllocatedRewards(2)
      const allocatedKeep = allocated.div(tokenDecimalMultiplier)

      // No rewards allocated in first and second interval.
      // Full allocation for third interval would be
      // 178,200,000 * 10% = 17,820,000.
      // Because just 4% of minimum keep quota is met, the allocation is
      // 17,820,000 * 4% = 712,800.
      expect(allocatedKeep).to.eq.BN(712800)
    })
  })

  describe("rewards distribution", async () => {
    it("should not be possible when a keep is not closed", async () => {
      await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
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
      await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
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
      await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart + 1)

      await timeJumpToEndOfInterval(0)

      const keepTerminatedAddress = await keepFactory.getKeepAtIndex(0)
      const keepTerminated = await BondedECDSAKeepStub.at(keepTerminatedAddress)
      await keepTerminated.publicMarkAsTerminated()

      const keepAddress = await keepFactory.getKeepAtIndex(1)
      const keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress)

      // Full allocation for the first interval would be 7,128,000 KEEP.
      // Because just 2 keeps were created, the allocation is:
      // 7,128,000 * 0.2% = 14,256.
      // The reward per keep is 14,256 / 2 = 7128.
      // First keep was terminated, only the second keep was closed properly.
      // Member receives: 7128 / 3 = 2376 (3 signers per keep)
      const expectedBeneficiaryBalance = new BN(2376)
      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)

      // The 178,200,000 - 14,256 = 178,185,744 stays in unallocated
      // rewards and the fact one keep was terminated needs to be reported to
      // recalculate the unallocated amount
      let unallocated = await rewardsContract.unallocatedRewards()
      let unallocatedInKeep = unallocated.div(tokenDecimalMultiplier)
      expect(unallocatedInKeep).to.eq.BN(178185744)

      await rewardsContract.reportTermination(keepTerminatedAddress)

      unallocated = await rewardsContract.unallocatedRewards()
      unallocatedInKeep = unallocated.div(tokenDecimalMultiplier)
      expect(unallocatedInKeep).to.eq.BN(178192872) // 178,185,744 + 7,128
    })

    it("should correctly distribute rewards between beneficiaries", async () => {
      for (let i = 0; i < 8; i++) {
        await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      }

      await timeJumpToEndOfInterval(0)

      let keepAddress = await keepFactory.getKeepAtIndex(0)
      let keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress)

      // reward for the first interval: 7,128,000 KEEP
      // keeps created: 8 => 891,000 KEEP per keep
      // member receives: 891,000 / 3 = 297,000 (3 signers per keep)
      const expectedBeneficiaryBalance = new BN(297000)
      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)

      // verify second keep in this interval
      keepAddress = await keepFactory.getKeepAtIndex(1)
      keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress)

      // 297,000 * 2 = 594,000
      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance.muln(2))
    })
  })

  async function assertKeepBalanceOfBeneficiaries(expectedBalance) {
    // Solidity is not very good when it comes to floating point precision,
    // we are allowing for ~1 KEEP difference margin between expected and
    // actual value.
    const precision = 1

    for (let i = 0; i < beneficiaries; i++) {
      const actualBalance = (await keepToken.balanceOf(beneficiaries[i])).div(
        tokenDecimalMultiplier
      )

      expect(actualBalance).to.gte.BN(expectedBalance.subn(precision))
      expect(actualBalance).to.lte.BN(expectedBalance.addn(precision))
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
