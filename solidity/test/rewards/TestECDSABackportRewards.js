const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("../helpers/snapshot")

const {expectRevert} = require("@openzeppelin/test-helpers")

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
const BondedECDSAKeep = contract.fromArtifact("BondedECDSAKeep")
const ECDSABackportRewards = contract.fromArtifact("ECDSABackportRewards")

const KeepTokenGrant = contract.fromArtifact("TokenGrant")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

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
  let keepMembers

  const owner = accounts[0]

  const keepSize = 16
  const numberOfCreatedKeeps = 41
  const keepCreationTimestamp = 1589408353

  // 1,000,000,000 - total KEEP supply
  //   200,000,000 - 20% of the total supply goes to staker rewards
  //   180,000,000 - 90% of staker rewards goes to the ECDSA stakers
  //     1,800,000 - 1% of ECDSA staker rewards goes to May - Sep keeps
  const ECDSABackportKEEPRewards = 1800000

  before(async () => {
    await BondedSortitionPoolFactory.detectNetwork()
    await BondedSortitionPoolFactory.link(
      "StackLib",
      (await StackLib.new({from: owner})).address
    )

    await initializeNewFactory()
    keepMembers = await createMembers()
    rewardsContract = await ECDSABackportRewards.new(
      keepToken.address,
      keepFactory.address,
      {from: owner}
    )

    await fund(ECDSABackportKEEPRewards)
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

      const allocated = await rewardsContract.getAllocatedRewards(0)
      expect(allocated).to.eq.BN(expectedAllocation)
    })
  })

  describe("rewards withdrawal", async () => {
    it("should correctly distribute rewards between beneficiaries", async () => {
      // 1800000 / 41 / 16 = 2743.9024 KEEP
      // First 15 beneficiaries receive 2743 KEEP
      // Decimals (0.9024) are rolled over to the last keep signer.
      // The last keep signer receives 0.9024 * 15 = 13.536 more KEEP than other signers.
      // The last keep signer receives 2743.9024 + 13.536 = 2757.4384 KEEP => 2757
      //
      // All 16 signers belong to all 41 keeps for testing purposes.
      // KEEP is added to the signers in every iteration; total 41 times (number of keeps)
      const expectedSingleReward = 2743
      const expectedSingleRewardForLastSigner = 2757

      for (let i = 0; i < numberOfCreatedKeeps; i++) {
        const keepAddress = await keepFactory.getKeepAtIndex(i)
        await rewardsContract.receiveReward(keepAddress)

        await assertKeepBalanceOfBeneficiaries(
          i,
          expectedSingleReward,
          expectedSingleRewardForLastSigner
        )
      }
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

  async function assertKeepBalanceOfBeneficiaries(
    keepNumber,
    expectedSingleReward,
    expectedSingleRewardForLastSigner
  ) {
    // Check the balance of all beneficiaries but the last one.
    for (let i = 0; i < keepSize - 1; i++) {
      const actualBalance = await keepToken.balanceOf(
        keepMembers[i].beneficiary
      )
      const expectedBalance =
        expectedSingleReward + keepNumber * expectedSingleReward

      expect(actualBalance).to.eq.BN(expectedBalance)
    }

    const actualLastSignerBalance = await keepToken.balanceOf(
      keepMembers[keepSize - 1].beneficiary
    )
    const expectedLastSignerBalance =
      expectedSingleRewardForLastSigner +
      keepNumber * expectedSingleRewardForLastSigner

    // Check the balance of the last beneficiary.
    expect(actualLastSignerBalance).to.eq.BN(expectedLastSignerBalance)
  }

  async function initializeNewFactory() {
    keepToken = await KeepToken.new({from: owner})
    const keepTokenGrant = await KeepTokenGrant.new(keepToken.address)
    registry = await KeepRegistry.new({from: owner})

    bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new({
      from: owner,
    })
    await TokenStaking.detectNetwork()
    await TokenStaking.link(
      "MinimumStakeSchedule",
      (await MinimumStakeSchedule.new({from: owner})).address
    )
    await TokenStaking.link(
      "GrantStaking",
      (await GrantStaking.new({from: owner})).address
    )
    await TokenStaking.link("Locks", (await Locks.new({from: owner})).address)
    await TokenStaking.link("TopUps", (await TopUps.new({from: owner})).address)

    const stakingEscrow = await TokenStakingEscrow.new(
      keepToken.address,
      keepTokenGrant.address,
      {from: owner}
    )

    const stakeInitializationPeriod = 30 // In seconds

    tokenStaking = await TokenStaking.new(
      keepToken.address,
      keepTokenGrant.address,
      stakingEscrow.address,
      registry.address,
      stakeInitializationPeriod,
      {from: owner}
    )
    tokenGrant = await TokenGrant.new(keepToken.address, {from: owner})

    keepBonding = await KeepBonding.new(
      registry.address,
      tokenStaking.address,
      tokenGrant.address,
      {from: owner}
    )
    randomBeacon = await RandomBeaconStub.new({from: owner})
    const bondedECDSAKeepMasterContract = await BondedECDSAKeep.new({
      from: owner,
    })
    keepFactory = await BondedECDSAKeepFactoryStub.new(
      bondedECDSAKeepMasterContract.address,
      bondedSortitionPoolFactory.address,
      tokenStaking.address,
      keepBonding.address,
      randomBeacon.address,
      {from: owner}
    )

    await registry.approveOperatorContract(keepFactory.address, {from: owner})
  }

  async function createMembers() {
    const membersArr = []

    // 16 members in each keep
    for (let i = 0; i < keepSize; i++) {
      const operator = accounts[i]
      const beneficiary = accounts[keepSize + i]
      await tokenStaking.setBeneficiary(operator, beneficiary)
      const member = {
        operator: operator,
        beneficiary: beneficiary,
      }

      membersArr.push(member)
    }

    return membersArr
  }

  async function createKeeps() {
    const members = keepMembers.map((m) => m.operator)
    for (let i = 0; i < numberOfCreatedKeeps; i++) {
      await keepFactory.stubOpenKeep(
        owner,
        members,
        tokenStaking.address,
        keepFactory.address,
        keepCreationTimestamp
      )
    }
  }

  async function fund(amount) {
    await keepToken.approveAndCall(rewardsContract.address, amount, "0x0", {
      from: owner,
    })
    await rewardsContract.markAsFunded({from: owner})
  }
})
