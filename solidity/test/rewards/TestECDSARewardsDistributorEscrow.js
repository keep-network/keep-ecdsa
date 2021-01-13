const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("../helpers/snapshot")
const {expectRevert} = require("@openzeppelin/test-helpers")
const {expect} = require("chai")

const KeepToken = contract.fromArtifact("KeepToken")
const TokenStakingStub = contract.fromArtifact("TokenStakingStub")
const PhasedEscrow = contract.fromArtifact("PhasedEscrow")
const ECDSARewardsEscrowBeneficiary = contract.fromArtifact(
  "ECDSARewardsEscrowBeneficiary"
)
const ECDSARewardsDistributor = contract.fromArtifact("ECDSARewardsDistributor")
const ECDSARewardsDistributorEscrow = contract.fromArtifact(
  "ECDSARewardsDistributorEscrow"
)

describe("ECDSARewardsDistributorEscrow", () => {
  const owner = accounts[1]
  const thirdParty = accounts[2]

  const tokenDecimalMultiplier = web3.utils.toBN(10).pow(web3.utils.toBN(18))
  const totalRewards = web3.utils.toBN(178200000).mul(tokenDecimalMultiplier)

  let token
  let rewardsDistributor
  let escrow

  before(async () => {
    token = await KeepToken.new({from: owner})
    tokenStaking = await TokenStakingStub.new({from: owner})
    rewardsDistributor = await ECDSARewardsDistributor.new(
      token.address,
      tokenStaking.address,
      {
        from: owner,
      }
    )
    escrow = await ECDSARewardsDistributorEscrow.new(
      token.address,
      rewardsDistributor.address,
      {from: owner}
    )

    await rewardsDistributor.transferOwnership(escrow.address, {from: owner})
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("funding", async () => {
    it("can be done from phased escrow", async () => {
      const fundingEscrow = await PhasedEscrow.new(token.address, {from: owner})
      await token.approveAndCall(fundingEscrow.address, totalRewards, "0x0", {
        from: owner,
      })

      const beneficiary = await ECDSARewardsEscrowBeneficiary.new(
        token.address,
        escrow.address,
        {from: owner}
      )
      await beneficiary.transferOwnership(fundingEscrow.address, {from: owner})
      await fundingEscrow.setBeneficiary(beneficiary.address, {from: owner})

      await fundingEscrow.withdraw(totalRewards, {from: owner})
      expect(await token.balanceOf(escrow.address)).to.eq.BN(totalRewards)
    })
  })

  describe("allocateInterval", async () => {
    const merkleRoot =
      "0x65b315f4565a40f738cbaaef7dbab4ddefa14620407507d0f2d5cdbd1d8063f6"
    const amount = web3.utils.toBN(999998997)

    beforeEach(async () => {
      // The initial state set up with approveAndCall and confirmed with the
      // assertion below is the escrow state after getting funded from
      // another PhasedEscrow, as demonstrated in "funding" describe.
      // This reflects the flow of funds on mainnet for the updated ECDSA
      // staker rewards deployment.
      await token.approveAndCall(escrow.address, totalRewards, "0x0", {
        from: owner,
      })
      expect(await token.balanceOf(escrow.address)).to.eq.BN(totalRewards)
    })

    it("can not be called by non-owner", async () => {
      await expectRevert(
        escrow.allocateInterval(merkleRoot, amount, {from: thirdParty}),
        "Ownable: caller is not the owner"
      )
    })

    it("can be called by owner", async () => {
      await escrow.allocateInterval(merkleRoot, amount, {from: owner})
      // ok, no reverts
    })

    it("allocates reward distribution", async () => {
      await escrow.allocateInterval(merkleRoot, amount, {from: owner})

      const eventList = await rewardsDistributor.getPastEvents(
        "RewardsAllocated",
        {
          fromBlock: 0,
          toBlock: "latest",
        }
      )

      expect(eventList.length).to.equal(1, "incorrect number of emitted events")
      const event = eventList[0].returnValues
      expect(event.merkleRoot).to.equal(merkleRoot, "unexpected merkle root")
      expect(event.amount).to.eq.BN(amount)
    })

    it("allocates multiple reward distributions", async () => {
      const merkleRoot2 =
        "0xa7418520411d369b511eabb10ffb214c72b521ca0f6bd021fa83d9c47e65227e"
      const amount2 = web3.utils.toBN(1337)

      await escrow.allocateInterval(merkleRoot, amount, {from: owner})
      await escrow.allocateInterval(merkleRoot2, amount2, {from: owner})

      const eventList = await rewardsDistributor.getPastEvents(
        "RewardsAllocated",
        {
          fromBlock: 0,
          toBlock: "latest",
        }
      )

      expect(eventList.length).to.equal(2, "incorrect number of emitted events")
      const event1 = eventList[0].returnValues
      expect(event1.merkleRoot).to.equal(merkleRoot, "unexpected merkle root")
      expect(event1.amount).to.eq.BN(amount)
      const event2 = eventList[1].returnValues
      expect(event2.merkleRoot).to.equal(merkleRoot2, "unexpected merkle root")
      expect(event2.amount).to.eq.BN(amount2)

      expect(await token.balanceOf(rewardsDistributor.address)).to.eq.BN(
        amount.add(amount2)
      )
    })
  })
})
