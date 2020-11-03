const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {expectRevert, time} = require("@openzeppelin/test-helpers")

const BondedECDSAKeepStub = contract.fromArtifact("BondedECDSAKeepStub")
const ECDSARewards = contract.fromArtifact("ECDSARewardsStub")
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
      tokenStaking.address,
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

      await timeJumpToEndOfIntervalIfApplicable(0)
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
      await rewardsContract.allocateRewards(0)

      const firstIntervalEnd = await rewardsContract.endOf(0)
      await keepFactory.stubBatchOpenFakeKeeps(50, firstIntervalEnd.addn(1))

      await timeJumpToEndOfIntervalIfApplicable(1)

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
      const secondIntervalEnd = await rewardsContract.endOf(1)
      await keepFactory.stubBatchOpenFakeKeeps(40, secondIntervalEnd.addn(1))

      await timeJumpToEndOfIntervalIfApplicable(2)

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

  describe("rewards allocation", async () => {
    it("should not be possible when a keep is not closed", async () => {
      await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      const keepAddress = await keepFactory.getKeepAtIndex(0)

      const isEligible = await rewardsContract.eligibleForReward(keepAddress)
      expect(isEligible).to.be.false

      await timeJumpToEndOfIntervalIfApplicable(0)
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

      await timeJumpToEndOfIntervalIfApplicable(0)
      await expectRevert(
        rewardsContract.receiveReward(keepAddress),
        "Keep is not closed"
      )
    })

    it("should exclude terminated groups", async () => {
      await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart + 1)

      await timeJumpToEndOfIntervalIfApplicable(0)

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
      const expectedBeneficiaryAllocation = new BN(2376)
      await assertAllocatedRewards(expectedBeneficiaryAllocation, 0)

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

    it("should report group terminations in batch", async () => {
      // Open 5 keeps
      for (let i = 0; i < 5; i++) {
        await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart + i)
      }

      await timeJumpToEndOfIntervalIfApplicable(0)

      const keepTerminatedAddresses = []
      // Mark 3 keeps as terminated
      for (let i = 0; i < 3; i++) {
        const keepTerminatedAddress = await keepFactory.getKeepAtIndex(i)
        const keepTerminated = await BondedECDSAKeepStub.at(
          keepTerminatedAddress
        )
        await keepTerminated.publicMarkAsTerminated()
        keepTerminatedAddresses.push(keepTerminatedAddress)
      }

      const keepClosedAddresses = []
      // Mark 2 keeps as closed properly
      for (let i = 3; i < 5; i++) {
        const keepAddress = await keepFactory.getKeepAtIndex(i)
        const keep = await BondedECDSAKeepStub.at(keepAddress)
        await keep.publicMarkAsClosed()
        keepClosedAddresses.push(keepAddress)
      }

      await rewardsContract.receiveRewards(keepClosedAddresses)

      // Full allocation for the first interval is 7,128,000 KEEP.
      // Because just 5 keeps were created, the allocation is:
      // 7,128,000 * 0.5% = 35,640.
      // The reward per keep is 35,640 / 5 = 7,128
      // 178,200,000 - 35,640 = 178,164,360 stays in unallocated rewards
      let actualUnallocated = await rewardsContract.unallocatedRewards()
      actual = actualUnallocated.div(tokenDecimalMultiplier)
      expect(actual).to.eq.BN(178164360)

      // The fact that three keeps were terminated needs to be
      // reported to recalculate the unallocated amount.
      await rewardsContract.reportTerminations(keepTerminatedAddresses)

      // Rewards returned to the unallocated pool upon termination:
      // 7,128 * 3 = 21,384
      actualUnallocated = await rewardsContract.unallocatedRewards()
      actual = actualUnallocated.div(tokenDecimalMultiplier)
      // 178,164,360 + 21,384 = 178,185,744
      expect(actual).to.eq.BN(178185744)
    })

    it("should cap the maximum reward per beneficiary", async () => {
      // There is currently no way to implement a test with the original
      // 3M KEEP cap for beneficiary given block gas limit and test execution
      // time. Hence, we set the limit to 10k KEEP for test purposes.
      const rewardsCap = web3.utils.toBN(10000).mul(tokenDecimalMultiplier)
      await rewardsContract.setBeneficiaryRewardCap(rewardsCap)

      for (let i = 0; i < 10; i++) {
        await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      }

      await timeJumpToEndOfIntervalIfApplicable(0)

      const keepAddresses = []
      // Mark 5 keeps as properly closed
      for (let i = 0; i < 5; i++) {
        const keepAddress = await keepFactory.getKeepAtIndex(i)
        keepAddresses.push(keepAddress)
        const keep = await BondedECDSAKeepStub.at(keepAddress)
        await keep.publicMarkAsClosed()
      }

      await rewardsContract.receiveRewards(keepAddresses)
      await withdrawRewards(0)

      // Full allocation for the first interval would be
      // 178,200,000 * 4% = 7,128,000.
      // Because just 1% of minimum keep quota is met, the allocation is
      // 7,128,000 * 1% = 71,280.
      // Keeps created: 10 => 7,128 KEEP per keep
      // Member receives: 7,128 / 3 = 2,376 (3 signers per keep)
      // Each member was in 5 properly closed keeps: 2,376 * 5 = 11,880
      // 11,880 exceeds the maximum rewards cap per beneficiary, which is 10,000
      // Expected beneficiary balance should be equal to max cap.
      const expectedBeneficiaryBalance = new BN(10000)
      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)

      // Each beneficiary has a surplus of 11,880 - 10,000 = 1,880 that has to be
      // returned back to the unallocated pool rewards: 1,880 * 3 = 5,640
      // Expected unallocated pool: 178,200,000 - 71,280 + 5,640 = 178,134,360
      const expectedUnallocatedRewards = new BN(178134360)
      const actualUnallocatedRewards = (
        await rewardsContract.unallocatedRewards()
      ).div(tokenDecimalMultiplier)

      expect(actualUnallocatedRewards).to.eq.BN(expectedUnallocatedRewards)
    })

    it("should keep per-interval beneficiary reward cap at 3M KEEP", async () => {
      const cap = await rewardsContract.beneficiaryRewardCap()
      expect(cap).to.eq.BN(web3.utils.toBN(3000000).mul(tokenDecimalMultiplier))
    })
  })

  describe("rewards withdrawal", async () => {
    it("should correctly distribute rewards between beneficiaries", async () => {
      for (let i = 0; i < 8; i++) {
        await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      }

      await timeJumpToEndOfIntervalIfApplicable(0)

      let keepAddress = await keepFactory.getKeepAtIndex(0)
      let keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress)
      await withdrawRewards(0)

      // Full allocation for the first interval would be
      // 178,200,000 * 4% = 7,128,000.
      // Because just 0.8% of minimum keep quota is met, the allocation is
      // 7,128,000 * 0.8% = 57,024.
      // keeps created: 8 => 7,128 KEEP per keep
      // member receives: 7,128 / 3 = 2,376 (3 signers per keep)
      const expectedBeneficiaryBalance = new BN(2376)
      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)

      // verify second keep in this interval
      keepAddress = await keepFactory.getKeepAtIndex(1)
      keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress)
      await withdrawRewards(0)

      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance.muln(2))
    })

    it("should revert when rewards have been withdrawn already", async () => {
      for (let i = 0; i < 8; i++) {
        await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      }

      await timeJumpToEndOfIntervalIfApplicable(0)

      // mark first keep as closed
      const keepAddress = await keepFactory.getKeepAtIndex(0)
      const keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress)

      await withdrawRewards(0)

      for (let i = 0; i < operators.length; i++) {
        await expectRevert(
          rewardsContract.withdrawRewards(0, operators[i], {
            from: operators[i],
          }),
          "No rewards to withdraw"
        )
      }
    })

    it("should revert when rewards have been received already", async () => {
      for (let i = 0; i < 8; i++) {
        await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      }

      await timeJumpToEndOfIntervalIfApplicable(0)

      // mark first keep as closed
      const keepAddress = await keepFactory.getKeepAtIndex(0)
      const keep = await BondedECDSAKeepStub.at(keepAddress)
      await keep.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress)

      await withdrawRewards(0)

      await expectRevert(
        rewardsContract.receiveReward(keepAddress),
        "Rewards already claimed"
      )
    })

    it("should correctly receive rewards from multiple keeps", async () => {
      for (let i = 0; i < 10; i++) {
        await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      }

      await timeJumpToEndOfIntervalIfApplicable(0)

      const keepAddress0 = await keepFactory.getKeepAtIndex(0)
      const keepAddress1 = await keepFactory.getKeepAtIndex(1)
      const keepAddresses = [keepAddress0, keepAddress1]
      const keep0 = await BondedECDSAKeepStub.at(keepAddress0)
      const keep1 = await BondedECDSAKeepStub.at(keepAddress1)

      await keep0.publicMarkAsClosed()
      await keep1.publicMarkAsClosed()

      await rewardsContract.receiveRewards(keepAddresses)
      await withdrawRewards(0)

      // Full allocation for the first interval would be
      // 178,200,000 * 4% = 7,128,000.
      // Because just 1% of minimum keep quota is met, the allocation is
      // 7,128,000 * 1% = 71,280.
      // keeps created: 10 => 7,128 KEEP per keep
      // member receives: 7,128 / 3 = 2,376 (3 signers per keep)
      // member was in 2 properly closed keeps: 2,376 * 2 = 4,752
      const expectedBeneficiaryBalance = new BN(4752)

      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)
    })

    it("should correctly receive rewards in the same interval but not same time", async () => {
      for (let i = 0; i < 10; i++) {
        await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      }

      await timeJumpToEndOfIntervalIfApplicable(0)

      const keepAddress0 = await keepFactory.getKeepAtIndex(0)
      const keep0 = await BondedECDSAKeepStub.at(keepAddress0)
      await keep0.publicMarkAsClosed()

      // Full allocation for the first interval would be
      // 178,200,000 * 4% = 7,128,000.
      // Because just 1% of minimum keep quota is met, the allocation is
      // 7,128,000 * 1% = 71,280.
      // keeps created: 10 => 7,128 KEEP per keep
      // member receives: 7,128 / 3 = 2,376 (3 signers per keep)
      await rewardsContract.receiveReward(keepAddress0)
      await withdrawRewards(0)

      let expectedBeneficiaryBalance = new BN(2376)
      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)

      // Second keep will be closed a day later
      await time.increase(time.duration.days(1))

      const keepAddress1 = await keepFactory.getKeepAtIndex(1)
      const keep1 = await BondedECDSAKeepStub.at(keepAddress1)
      await keep1.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress1)
      await withdrawRewards(0)

      // each member was in 2 properly closed keeps: 2,376 * 2 = 4,752
      expectedBeneficiaryBalance = new BN(4752)
      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)
    })

    it("should correctly track withdrawable and withdrawn amount", async () => {
      for (let i = 0; i < 8; i++) {
        await keepFactory.stubOpenKeep(owner, operators, firstIntervalStart)
      }

      await timeJumpToEndOfIntervalIfApplicable(0)

      // mark first keep as closed
      const keepAddress0 = await keepFactory.getKeepAtIndex(0)
      const keep0 = await BondedECDSAKeepStub.at(keepAddress0)
      await keep0.publicMarkAsClosed()

      // mark second keep as closed
      const keepAddress1 = await keepFactory.getKeepAtIndex(1)
      const keep1 = await BondedECDSAKeepStub.at(keepAddress1)
      await keep1.publicMarkAsClosed()

      await rewardsContract.receiveReward(keepAddress0)
      await rewardsContract.receiveReward(keepAddress1)

      // Full allocation for the first interval would be
      // 178,200,000 * 4% = 7,128,000.
      // Because just 0.8% of minimum keep quota is met, the allocation is
      // 7,128,000 * 0.8% = 57,024.
      // keeps created: 8 => 7,128 KEEP per keep
      // member receives: 7,128 / 3 = 2,376 (3 signers per keep)
      const expectedBeneficiaryBalance = new BN(2376)
      // 2,376 * 2 = 4,752
      const expectedBalanceFromTwoKeeps = expectedBeneficiaryBalance.muln(2)

      await assertWithdrawableRewards(expectedBalanceFromTwoKeeps, 0)
      await assertWithdrawnRewards(new BN(0), 0)

      await withdrawRewards(0)

      await assertWithdrawableRewards(new BN(0), 0)
      await assertWithdrawnRewards(expectedBalanceFromTwoKeeps, 0)
    })
  })

  async function assertKeepBalanceOfBeneficiaries(expectedBalance) {
    // Solidity is not very good when it comes to floating point precision,
    // we are allowing for ~1 KEEP difference margin between expected and an
    // actual value.
    const precision = 1

    for (let i = 0; i < beneficiaries.length; i++) {
      const actualBalance = (await keepToken.balanceOf(beneficiaries[i])).div(
        tokenDecimalMultiplier
      )

      expect(actualBalance).to.gte.BN(expectedBalance.subn(precision))
      expect(actualBalance).to.lte.BN(expectedBalance.addn(precision))
    }
  }

  async function assertAllocatedRewards(expectedBalance, interval) {
    const precision = 1

    for (let i = 0; i < operators.length; i++) {
      const actualWithdrawnRewards = await rewardsContract.getAllocatedRewards(
        interval,
        operators[i]
      )
      const actual = actualWithdrawnRewards.div(tokenDecimalMultiplier)

      expect(actual).to.gte.BN(expectedBalance.subn(precision))
      expect(actual).to.lte.BN(expectedBalance.addn(precision))
    }
  }

  async function assertWithdrawnRewards(expectedBalance, interval) {
    const precision = 1

    for (let i = 0; i < operators.length; i++) {
      const actualWithdrawnRewards = await rewardsContract.getWithdrawnRewards(
        interval,
        operators[i]
      )
      const actual = actualWithdrawnRewards.div(tokenDecimalMultiplier)

      expect(actual).to.gte.BN(expectedBalance.subn(precision))
      expect(actual).to.lte.BN(expectedBalance.addn(precision))
    }
  }

  async function assertWithdrawableRewards(expectedBalance, interval) {
    const precision = 1

    for (let i = 0; i < operators.length; i++) {
      const actualWithdrawnRewards = await rewardsContract.getWithdrawableRewards(
        interval,
        operators[i]
      )
      const actual = actualWithdrawnRewards.div(tokenDecimalMultiplier)

      expect(actual).to.gte.BN(expectedBalance.subn(precision))
      expect(actual).to.lte.BN(expectedBalance.addn(precision))
    }
  }

  async function withdrawRewards(interval) {
    for (let i = 0; i < operators.length; i++) {
      await rewardsContract.withdrawRewards(interval, operators[i], {
        from: operators[i],
      })
    }
  }

  async function timeJumpToEndOfIntervalIfApplicable(intervalNumber) {
    const endOf = await rewardsContract.endOf(intervalNumber)
    const now = await time.latest()

    if (now.lt(endOf)) {
      await time.increaseTo(endOf.addn(60))
    }
  }
})
