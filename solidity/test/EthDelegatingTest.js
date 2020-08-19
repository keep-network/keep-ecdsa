const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("./helpers/snapshot")

const KeepRegistry = contract.fromArtifact("KeepRegistry")
const EthDelegating = contract.fromArtifact("EthDelegating")

const {expectEvent, expectRevert} = require("@openzeppelin/test-helpers")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect
const assert = chai.assert

describe("EthDelegating", function () {
  let registry
  let ethDelegating

  let operator
  let authorizer
  let beneficiary
  let owner

  before(async () => {
    operator = accounts[1]
    authorizer = accounts[2]
    beneficiary = accounts[3]
    owner = accounts[4]

    registry = await KeepRegistry.new()
    ethDelegating = await EthDelegating.new(registry.address)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("delegate", async () => {
    it("registers delegate", async () => {
      await ethDelegating.delegate(operator, beneficiary, authorizer, {
        from: owner,
      })

      assert.equal(
        await ethDelegating.ownerOf(operator),
        owner,
        "incorrect owner address"
      )

      assert.equal(
        await ethDelegating.beneficiaryOf(operator),
        beneficiary,
        "incorrect beneficiary address"
      )

      assert.equal(
        await ethDelegating.authorizerOf(operator),
        authorizer,
        "incorrect authorizer address"
      )

      expect(await ethDelegating.balanceOf(operator)).to.eq.BN(
        0,
        "incorrect delegation balance"
      )
    })

    it("emits events", async () => {
      const receipt = await ethDelegating.delegate(
        operator,
        beneficiary,
        authorizer,
        {
          from: owner,
        }
      )

      await expectEvent(receipt, "Delegated", {
        owner: owner,
        operator: operator,
      })

      await expectEvent(receipt, "OperatorDelegated", {
        operator: operator,
        beneficiary: beneficiary,
        authorizer: authorizer,
      })
    })

    it("allows multiple operators for the same owner", async () => {
      const operator2 = accounts[5]

      await ethDelegating.delegate(operator, beneficiary, authorizer, {
        from: owner,
      })

      await ethDelegating.delegate(operator2, beneficiary, authorizer, {
        from: owner,
      })
    })

    it("reverts if operator is already in use", async () => {
      await ethDelegating.delegate(operator, beneficiary, authorizer, {
        from: owner,
      })

      await expectRevert(
        ethDelegating.delegate(operator, accounts[5], accounts[5]),
        "Operator already in use"
      )
    })
  })
})
