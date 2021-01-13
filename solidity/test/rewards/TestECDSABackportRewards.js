const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("../helpers/snapshot")
const {initialize, fund, createMembers} = require("./rewardsSetup")

const {expectRevert} = require("@openzeppelin/test-helpers")

const ECDSABackportRewards = contract.fromArtifact("ECDSABackportRewards")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

describe("ECDSABackportRewards", () => {
  let rewardsContract
  let keepToken

  let tokenStaking
  let keepFactory
  let keepMembers

  const owner = accounts[0]

  const numberOfCreatedKeeps = 41
  const tokenDecimalMultiplier = web3.utils.toBN(10).pow(web3.utils.toBN(18))
  const firstKeepCreationTimestamp = 1589408352

  // 1,000,000,000 - total KEEP supply
  //   200,000,000 - 20% of the total supply goes to staker rewards
  //   180,000,000 - 90% of staker rewards goes to the ECDSA staker rewards
  //     1,800,000 - 1% of ECDSA staker rewards goes to May - Sep keeps
  const ECDSABackportKEEPRewards = web3.utils
    .toBN(1800000)
    .mul(tokenDecimalMultiplier)

  before(async () => {
    const setup = await initialize()
    tokenStaking = setup.tokenStaking
    keepToken = setup.keepToken
    keepFactory = setup.keepFactory

    keepMembers = await createMembers(tokenStaking)
    rewardsContract = await ECDSABackportRewards.new(
      keepToken.address,
      keepFactory.address,
      {from: owner}
    )

    await fund(keepToken, rewardsContract, ECDSABackportKEEPRewards)
    await createKeeps()
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("interval allocation", async () => {
    it("should equal the full allocation", async () => {
      const expectedAllocation = 1800000

      await rewardsContract.allocateRewards(0)

      const allocated = (await rewardsContract.getAllocatedRewards(0)).div(
        tokenDecimalMultiplier
      )
      expect(allocated).to.eq.BN(expectedAllocation)
    })
  })

  describe("rewards withdrawal", async () => {
    it("should correctly distribute rewards between beneficiaries", async () => {
      // All 3 signers belong to all 41 keeps for testing purposes.
      // KEEP is added to the signers in every iteration; total 41 times (number of keeps).
      //
      // 1800000 / 3 = 600000 KEEP.
      const expectedBeneficiaryBalance = new BN(600000)

      for (let i = 0; i < numberOfCreatedKeeps; i++) {
        const keepAddress = await keepFactory.getKeepAtIndex(i)
        await rewardsContract.receiveReward(keepAddress)
      }

      await assertKeepBalanceOfBeneficiaries(expectedBeneficiaryBalance)
    })

    it("should fail for non-existing group", async () => {
      await expectRevert(
        rewardsContract.receiveReward(
          "0x1111111111111111111111111111111111111111"
        ),
        "Keep not recognized by factory"
      )
    })
  })

  async function assertKeepBalanceOfBeneficiaries(expectedBalance) {
    // Solidity is not very good when it comes to floating point precision,
    // we are allowing for ~1 KEEP difference margin between expected and
    // actual value.
    const precision = 1

    for (let i = 0; i < keepMembers.length; i++) {
      const actualBalance = (
        await keepToken.balanceOf(keepMembers[i].beneficiary)
      ).div(tokenDecimalMultiplier)

      expect(actualBalance).to.gte.BN(expectedBalance.subn(precision))
      expect(actualBalance).to.lte.BN(expectedBalance.addn(precision))
    }
  }

  async function createKeeps() {
    const members = keepMembers.map((m) => m.operator)
    let keepCreationTimestamp = firstKeepCreationTimestamp
    for (let i = 0; i < numberOfCreatedKeeps; i++) {
      await keepFactory.stubOpenKeep(owner, members, keepCreationTimestamp)
      keepCreationTimestamp += 7200 // adding 2 hours interval between each opened keep
    }
  }
})
