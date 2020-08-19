const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("./helpers/snapshot")

const KeepRegistry = contract.fromArtifact("KeepRegistry")
const EthBonding = contract.fromArtifact("EthBonding")
const TestEtherReceiver = contract.fromArtifact("TestEtherReceiver")

const {expectEvent, expectRevert} = require("@openzeppelin/test-helpers")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect
const assert = chai.assert

describe("EthBonding", function () {
  let registry
  let ethBonding
  let etherReceiver

  let operator
  let authorizer
  let beneficiary
  let owner
  let bondCreator
  let sortitionPool

  before(async () => {
    operator = accounts[1]
    authorizer = accounts[2]
    beneficiary = accounts[3]
    owner = accounts[4]
    bondCreator = accounts[5]
    sortitionPool = accounts[6]

    registry = await KeepRegistry.new()
    ethBonding = await EthBonding.new(registry.address)
    etherReceiver = await TestEtherReceiver.new()

    await registry.approveOperatorContract(bondCreator)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("delegate", async () => {
    it("registers delegate", async () => {
      await ethBonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
      })

      assert.equal(
        await ethBonding.ownerOf(operator),
        owner,
        "incorrect owner address"
      )

      assert.equal(
        await ethBonding.beneficiaryOf(operator),
        beneficiary,
        "incorrect beneficiary address"
      )

      assert.equal(
        await ethBonding.authorizerOf(operator),
        authorizer,
        "incorrect authorizer address"
      )

      expect(await ethBonding.balanceOf(operator)).to.eq.BN(
        0,
        "incorrect delegation balance"
      )
    })

    it("emits events", async () => {
      const receipt = await ethBonding.delegate(
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

      await ethBonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
      })

      await ethBonding.delegate(operator2, beneficiary, authorizer, {
        from: owner,
      })
    })

    it("reverts if operator is already in use", async () => {
      await ethBonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
      })

      await expectRevert(
        ethBonding.delegate(operator, accounts[5], accounts[5]),
        "Operator already in use"
      )
    })
  })

  describe("withdraw", async () => {
    const value = new BN(1000)

    beforeEach(async () => {
      await ethBonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
      })

      await ethBonding.deposit(operator, {value: value})

      await ethBonding.authorizeOperatorContract(operator, bondCreator, {
        from: authorizer,
      })
    })

    it("can be called by operator", async () => {
      await ethBonding.withdraw(value, operator, {from: operator})
      // ok, no reverts
    })

    it("can be called by delegation owner", async () => {
      await ethBonding.withdraw(value, operator, {from: owner})
      // ok, no reverts
    })

    it("cannot be called by authorizer", async () => {
      await expectRevert(
        ethBonding.withdraw(value, operator, {from: authorizer}),
        "Only operator or the owner is allowed to withdraw bond"
      )
    })

    it("cannot be called by beneficiary", async () => {
      await expectRevert(
        ethBonding.withdraw(value, operator, {from: beneficiary}),
        "Only operator or the owner is allowed to withdraw bond"
      )
    })

    it("cannot be called by third party", async () => {
      const thirdParty = accounts[7]

      await expectRevert(
        ethBonding.withdraw(value, operator, {from: thirdParty}),
        "Only operator or the owner is allowed to withdraw bond"
      )
    })

    it("transfers unbonded value to beneficiary", async () => {
      const expectedUnbonded = 0

      const expectedBeneficiaryBalance = web3.utils
        .toBN(await web3.eth.getBalance(beneficiary))
        .add(value)

      await ethBonding.withdraw(value, operator, {from: operator})

      const unbonded = await ethBonding.availableUnbondedValue(
        operator,
        bondCreator,
        sortitionPool
      )
      expect(unbonded).to.eq.BN(expectedUnbonded, "invalid unbonded value")

      const actualBeneficiaryBalance = await web3.eth.getBalance(beneficiary)
      expect(actualBeneficiaryBalance).to.eq.BN(
        expectedBeneficiaryBalance,
        "invalid beneficiary balance"
      )
    })

    it("emits event", async () => {
      const value = new BN(90)

      const receipt = await ethBonding.withdraw(value, operator, {
        from: operator,
      })
      expectEvent(receipt, "UnbondedValueWithdrawn", {
        operator: operator,
        beneficiary: beneficiary,
        amount: value,
      })
    })

    it("reverts if insufficient unbonded value", async () => {
      const invalidValue = value.add(new BN(1))

      await expectRevert(
        ethBonding.withdraw(invalidValue, operator, {from: operator}),
        "Insufficient unbonded value"
      )
    })

    it("reverts if transfer fails", async () => {
      const operator2 = accounts[7]

      await etherReceiver.setShouldFail(true)

      await ethBonding.delegate(operator2, etherReceiver.address, authorizer, {
        from: owner,
      })

      await ethBonding.deposit(operator2, {value: value})

      await expectRevert(
        ethBonding.withdraw(value, operator2, {from: operator2}),
        "Transfer failed"
      )
    })
  })
})
