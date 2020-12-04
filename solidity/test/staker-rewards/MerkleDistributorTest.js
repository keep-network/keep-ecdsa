const {contract, web3} = require("@openzeppelin/test-environment")
const {expectRevert, expectEvent} = require("@openzeppelin/test-helpers")
const {createSnapshot, restoreSnapshot} = require("../helpers/snapshot")
const {testValues} = require("./rewardsData.js")

const ECDSARewardsDistributor = contract.fromArtifact("ECDSARewardsDistributor")
const KeepToken = contract.fromArtifact("KeepToken")

const BN = web3.utils.BN
const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

describe("MerkleDistributor", () => {
  let keepToken
  let rewardsDistributor

  before(async () => {
    keepToken = await KeepToken.new()
    rewardsDistributor = await ECDSARewardsDistributor.new(keepToken.address)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("allocating rewards", () => {
    it("should successfuly allocate KEEP tokens", async () => {
      let balance = await keepToken.balanceOf(rewardsDistributor.address)
      expect(balance).to.eq.BN(0)

      await allocateTokens(
        testValues.interval0.merkleRoot,
        testValues.interval0.amountToAllocate
      )

      balance = await keepToken.balanceOf(rewardsDistributor.address)
      expect(balance).to.eq.BN(testValues.interval0.amountToAllocate)
    })

    it("should successfuly emit RewardsAllocated event after rewards were allocated", async () => {
      const merkleRoot = testValues.interval0.merkleRoot
      const value = testValues.interval0.amountToAllocate

      await keepToken.approve(rewardsDistributor.address, value)
      const receipt = await rewardsDistributor.allocate(merkleRoot, value)

      const amount = new BN(value)
      expectEvent(receipt, "RewardsAllocated", {
        merkleRoot,
        amount,
      })
    })

    it("should fail allocating KEEP tokens without prior approval", async () => {
      await expectRevert(
        rewardsDistributor.allocate(
          testValues.interval0.merkleRoot,
          testValues.interval0.amountToAllocate
        ),
        "SafeERC20: low-level call failed"
      )
    })
  })

  describe("claiming rewards", () => {
    const merkle0 = testValues.interval0
    const merkle1 = testValues.interval1
    beforeEach(async () => {
      await allocateTokens(
        testValues.interval0.merkleRoot,
        testValues.interval0.amountToAllocate
      )

      await allocateTokens(
        testValues.interval1.merkleRoot,
        testValues.interval1.amountToAllocate
      )
    })

    it("should successfuly claim rewards and emit an event", async () => {
      const merkleRoot = merkle0.merkleRoot
      const index = merkle0.claims[0].index
      const account = merkle0.claims[0].account
      const amount = merkle0.claims[0].amount
      const proof = merkle0.claims[0].proof

      const claimed = await rewardsDistributor.claim(
        merkleRoot,
        index,
        account,
        amount,
        proof
      )

      expect(await keepToken.balanceOf(account)).to.eq.BN(amount)

      expectEvent(claimed, "RewardsClaimed", {
        merkleRoot,
        index,
        account,
        amount,
      })
    })

    it("should successfuly claim multiple rewards and update contract balance", async () => {
      const initialBalance = await keepToken.balanceOf(
        rewardsDistributor.address
      )

      let claimedAmounts = new BN(0)
      for (let i = 0; i < merkle0.claims.length; i++) {
        const merkleRoot = merkle0.merkleRoot
        const index = merkle0.claims[i].index
        const account = merkle0.claims[i].account
        const amount = merkle0.claims[i].amount
        const proof = merkle0.claims[i].proof

        claimedAmounts = claimedAmounts.addn(parseInt(amount))

        await rewardsDistributor.claim(
          merkleRoot,
          index,
          account,
          amount,
          proof
        )

        expect(await keepToken.balanceOf(account)).to.eq.BN(amount)
      }

      const actualBalance = await keepToken.balanceOf(
        rewardsDistributor.address
      )

      expect(actualBalance).to.eq.BN(
        initialBalance.sub(claimedAmounts),
        "invalid unbonded value"
      )
    })

    it("should revert claiming transaction when proof is not valid", async () => {
      const merkleRoot = merkle0.merkleRoot
      const index = merkle0.claims[0].index
      const account = merkle0.claims[0].account
      const amount = merkle0.claims[0].amount
      const proof = [
        "0x1111111111111111111111111111111111111111111111111111111111111111",
        "0xb335096692ef570690f2d858f2d52c268728d60b12a2a856f2691155ccf36377",
      ]

      await expectRevert(
        rewardsDistributor.claim(merkleRoot, index, account, amount, proof),
        "Invalid proof"
      )
    })

    it("should successfuly claim rewards from multiple merkle roots", async () => {
      const initialBalance = await keepToken.balanceOf(
        rewardsDistributor.address
      )

      let claimedAmounts = new BN(0)

      let merkleRoot = merkle0.merkleRoot
      let index = merkle0.claims[0].index
      let account = merkle0.claims[0].account
      let amount = merkle0.claims[0].amount
      let proof = merkle0.claims[0].proof

      claimedAmounts = claimedAmounts.addn(parseInt(amount))

      await rewardsDistributor.claim(merkleRoot, index, account, amount, proof)

      merkleRoot = merkle1.merkleRoot
      index = merkle1.claims[1].index
      account = merkle1.claims[1].account
      amount = merkle1.claims[1].amount
      proof = merkle1.claims[1].proof

      claimedAmounts = claimedAmounts.addn(parseInt(amount))

      await rewardsDistributor.claim(merkleRoot, index, account, amount, proof)

      const actualBalance = await keepToken.balanceOf(
        rewardsDistributor.address
      )

      expect(actualBalance).to.eq.BN(
        initialBalance.sub(claimedAmounts),
        "invalid unbonded value"
      )
    })

    it("should revert when claiming a reward twice", async () => {
      let merkleRoot = merkle0.merkleRoot
      let index = merkle0.claims[0].index
      let account = merkle0.claims[0].account
      let amount = merkle0.claims[0].amount
      let proof = merkle0.claims[0].proof

      await rewardsDistributor.claim(merkleRoot, index, account, amount, proof)

      await expectRevert(
        rewardsDistributor.claim(merkleRoot, index, account, amount, proof),
        "Reward already claimed"
      )

      merkleRoot = merkle1.merkleRoot
      index = merkle1.claims[1].index
      account = merkle1.claims[1].account
      amount = merkle1.claims[1].amount
      proof = merkle1.claims[1].proof

      await rewardsDistributor.claim(merkleRoot, index, account, amount, proof)

      await expectRevert(
        rewardsDistributor.claim(merkleRoot, index, account, amount, proof),
        "Reward already claimed"
      )
    })

    it("should check if the reward was claimed", async () => {
      const merkleRoot = merkle0.merkleRoot
      const index = merkle0.claims[0].index
      const account = merkle0.claims[0].account
      const amount = merkle0.claims[0].amount
      const proof = merkle0.claims[0].proof

      let isRewardClaimed = await rewardsDistributor.isClaimed(
        merkleRoot,
        index
      )
      expect(isRewardClaimed).to.be.false

      await rewardsDistributor.claim(merkleRoot, index, account, amount, proof)

      isRewardClaimed = await rewardsDistributor.isClaimed(merkleRoot, index)
      expect(isRewardClaimed).to.be.true
    })

    it("should revert when there are no rewards for a given merkle root", async () => {
      const merkleRoot =
        "0x1111111111111111111111111111111111111111111111111111111111111111"
      const index = "0"
      const account = "0x012ed55a0876Ea9e58277197DC14CbA47571CE28"
      const amount = "42"
      const proof = [
        "0x2222222222222222222222222222222222222222222222222222222222222222",
      ]

      await expectRevert(
        rewardsDistributor.claim(merkleRoot, index, account, amount, proof),
        "Rewards must be allocated for a given merkle root"
      )
    })
  })

  async function allocateTokens(merkleRoot, amount) {
    await keepToken.approve(rewardsDistributor.address, amount)
    await rewardsDistributor.allocate(merkleRoot, amount)
  }
})
