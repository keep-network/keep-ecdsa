const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("./helpers/snapshot")

const KeepRegistry = contract.fromArtifact("KeepRegistry")
const ETHStaking = contract.fromArtifact("ETHStaking")

const {expectEvent, expectRevert} = require("@openzeppelin/test-helpers")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect
const assert = chai.assert

describe("ETHStaking", function () {
  let registry
  let ethStaking

  let operator
  let authorizer
  let beneficiary
  let stakeOwner

  before(async () => {
    operator = accounts[1]
    authorizer = accounts[2]
    beneficiary = accounts[3]
    stakeOwner = accounts[4]

    registry = await KeepRegistry.new()
    ethStaking = await ETHStaking.new(registry.address)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("stake", async () => {
    it("registers stake", async () => {
      await ethStaking.stake(operator, beneficiary, authorizer, {
        from: stakeOwner,
      })

      assert.equal(
        await ethStaking.ownerOf(operator),
        stakeOwner,
        "incorrect stake owner address"
      )

      assert.equal(
        await ethStaking.beneficiaryOf(operator),
        beneficiary,
        "incorrect beneficiary address"
      )

      assert.equal(
        await ethStaking.authorizerOf(operator),
        authorizer,
        "incorrect authorizer address"
      )

      expect(await ethStaking.balanceOf(operator)).to.eq.BN(
        0,
        "incorrect stake balance"
      )
    })

    it("emits event", async () => {
      const receipt = await ethStaking.stake(
        operator,
        beneficiary,
        authorizer,
        {
          from: stakeOwner,
        }
      )

      await expectEvent(receipt, "Staked", {
        owner: stakeOwner,
        operator: operator,
        beneficiary: beneficiary,
        authorizer: authorizer,
      })
    })

    it("registers multiple operators for the same owner", async () => {
      const operator2 = accounts[5]

      await ethStaking.stake(operator, beneficiary, authorizer, {
        from: stakeOwner,
      })

      await ethStaking.stake(operator2, beneficiary, authorizer, {
        from: stakeOwner,
      })

      assert.deepEqual(
        await ethStaking.operatorsOf(stakeOwner),
        [operator, operator2],
        "incorrect operators for owner"
      )
    })

    it("reverts if operator is already in use", async () => {
      await ethStaking.stake(operator, beneficiary, authorizer, {
        from: stakeOwner,
      })

      await expectRevert(
        ethStaking.stake(operator, accounts[5], accounts[5]),
        "Operator already in use"
      )
    })
  })
})
