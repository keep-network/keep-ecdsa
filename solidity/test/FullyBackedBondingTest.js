const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("./helpers/snapshot")

const KeepRegistry = contract.fromArtifact("KeepRegistry")
const FullyBackedBonding = contract.fromArtifact("FullyBackedBonding")
const TestEtherReceiver = contract.fromArtifact("TestEtherReceiver")

const {expectEvent, expectRevert, time} = require("@openzeppelin/test-helpers")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect
const assert = chai.assert

describe("FullyBackedBonding", function () {
  const initializationPeriod = new BN(60)

  let minimumDelegationValue

  let registry
  let bonding
  let etherReceiver

  let operator
  let authorizer
  let beneficiary
  let owner
  let bondCreator
  let sortitionPool
  let thirdParty

  before(async () => {
    operator = accounts[1]
    authorizer = accounts[2]
    beneficiary = accounts[3]
    owner = accounts[4]
    bondCreator = accounts[5]
    sortitionPool = accounts[6]
    thirdParty = accounts[7]

    registry = await KeepRegistry.new()
    bonding = await FullyBackedBonding.new(
      registry.address,
      initializationPeriod
    )
    etherReceiver = await TestEtherReceiver.new()

    minimumDelegationValue = await bonding.MINIMUM_DELEGATION_DEPOSIT()

    await registry.approveOperatorContract(bondCreator)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("delegate", async () => {
    it("registers delegation", async () => {
      const {receipt} = await bonding.delegate(
        operator,
        beneficiary,
        authorizer,
        {
          from: owner,
          value: minimumDelegationValue,
        }
      )

      assert.equal(
        await bonding.ownerOf(operator),
        owner,
        "incorrect owner address"
      )

      assert.equal(
        await bonding.beneficiaryOf(operator),
        beneficiary,
        "incorrect beneficiary address"
      )

      assert.equal(
        await bonding.authorizerOf(operator),
        authorizer,
        "incorrect authorizer address"
      )

      expect(await bonding.balanceOf(operator)).to.eq.BN(
        0,
        "incorrect delegation balance"
      )

      const {timestamp: expectedCreatedAt} = await web3.eth.getBlock(
        receipt.blockNumber
      )

      const {createdAt, undelegatedAt} = await bonding.getDelegationInfo(
        operator
      )

      expect(createdAt).to.eq.BN(
        expectedCreatedAt,
        "incorrect created at value"
      )

      expect(undelegatedAt).to.eq.BN(0, "incorrect undelegated at value")
    })

    it("emits events", async () => {
      const receipt = await bonding.delegate(
        operator,
        beneficiary,
        authorizer,
        {
          from: owner,
          value: minimumDelegationValue,
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
        value: minimumDelegationValue,
      })
    })

    it("deposits passed value as unbonded value", async () => {
      const value = minimumDelegationValue

      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: value,
      })

      expect(await bonding.unbondedValue(operator)).to.eq.BN(
        value,
        "invalid unbonded value"
      )

      await bonding.authorizeOperatorContract(operator, bondCreator, {
        from: authorizer,
      })

      await bonding.authorizeSortitionPoolContract(operator, sortitionPool, {
        from: authorizer,
      })

      expect(
        await bonding.availableUnbondedValue(
          operator,
          bondCreator,
          sortitionPool
        )
      ).to.eq.BN(value, "invalid available unbonded value")
    })

    it("reverts if insufficient value passed", async () => {
      const value = minimumDelegationValue.subn(1)

      await expectRevert(
        bonding.delegate(operator, beneficiary, authorizer, {
          from: owner,
          value: value,
        }),
        "Insufficient delegation value"
      )
    })

    it("allows multiple operators for the same owner", async () => {
      const operator2 = accounts[5]

      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: minimumDelegationValue,
      })

      await bonding.delegate(operator2, beneficiary, authorizer, {
        from: owner,
        value: minimumDelegationValue,
      })
    })

    it("reverts if operator is already in use", async () => {
      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: minimumDelegationValue,
      })

      await expectRevert(
        bonding.delegate(operator, accounts[5], accounts[5]),
        "Operator already in use"
      )
    })
  })

  describe("topUp", async () => {
    const value = new BN(123)

    let initialDeposit

    beforeEach(async () => {
      initialDeposit = minimumDelegationValue

      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: initialDeposit,
      })
    })

    it("adds value to deposited on delegation", async () => {
      const expectedFinalBalance = initialDeposit.add(value)

      await bonding.topUp(operator, {
        value: value,
      })

      expect(await bonding.unbondedValue(operator)).to.eq.BN(
        expectedFinalBalance,
        "invalid final unbonded value"
      )
    })

    it("emits event", async () => {
      const receipt = await bonding.topUp(operator, {
        value: value,
      })

      expectEvent(receipt, "OperatorToppedUp", {
        operator: operator,
        value: value,
      })
    })

    it("reverts when no delegation happened", async () => {
      await expectRevert(
        bonding.topUp(thirdParty, {
          value: new BN(123),
        }),
        "Beneficiary not defined for the operator"
      )
    })
  })

  describe("deposit", async () => {
    it("adds value to deposited on delegation", async () => {
      const initialDeposit = minimumDelegationValue
      const value = new BN(123)
      const expectedFinalBalance = initialDeposit.add(value)

      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: initialDeposit,
      })

      expect(await bonding.unbondedValue(operator)).to.eq.BN(
        initialDeposit,
        "invalid initial unbonded value"
      )

      await bonding.deposit(operator, {
        value: value,
      })

      expect(await bonding.unbondedValue(operator)).to.eq.BN(
        expectedFinalBalance,
        "invalid final unbonded value"
      )

      await bonding.authorizeOperatorContract(operator, bondCreator, {
        from: authorizer,
      })

      await bonding.authorizeSortitionPoolContract(operator, sortitionPool, {
        from: authorizer,
      })

      expect(
        await bonding.availableUnbondedValue(
          operator,
          bondCreator,
          sortitionPool
        )
      ).to.eq.BN(expectedFinalBalance, "invalid final available unbonded value")
    })
  })

  describe("withdraw", async () => {
    const value = new BN(1000)

    let initialDeposit
    let delegationLockPeriod

    beforeEach(async () => {
      initialDeposit = minimumDelegationValue
      delegationLockPeriod = await bonding.DELEGATION_LOCK_PERIOD.call()

      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: initialDeposit,
      })

      await bonding.authorizeOperatorContract(operator, bondCreator, {
        from: authorizer,
      })

      await time.increase(delegationLockPeriod.addn(1))
    })

    it("can be called by operator", async () => {
      await bonding.withdraw(value, operator, {from: operator})
      // ok, no reverts
    })

    it("can be called by delegation owner", async () => {
      await bonding.withdraw(value, operator, {from: owner})
      // ok, no reverts
    })

    it("cannot be called before delegation lock period passes", async () => {
      const operator2 = await web3.eth.personal.newAccount("pass")

      await bonding.delegate(operator2, beneficiary, authorizer, {
        from: owner,
        value: initialDeposit,
      })

      await bonding.authorizeOperatorContract(operator2, bondCreator, {
        from: authorizer,
      })

      await time.increase(delegationLockPeriod.subn(1))

      await expectRevert(
        bonding.withdraw(value, operator2, {from: owner}),
        "Delegation lock period has not passed yet"
      )
    })

    it("cannot be called by authorizer", async () => {
      await expectRevert(
        bonding.withdraw(value, operator, {from: authorizer}),
        "Only operator or the owner is allowed to withdraw bond"
      )
    })

    it("cannot be called by beneficiary", async () => {
      await expectRevert(
        bonding.withdraw(value, operator, {from: beneficiary}),
        "Only operator or the owner is allowed to withdraw bond"
      )
    })

    it("cannot be called by third party", async () => {
      const thirdParty = accounts[7]

      await expectRevert(
        bonding.withdraw(value, operator, {from: thirdParty}),
        "Only operator or the owner is allowed to withdraw bond"
      )
    })

    it("transfers unbonded value to beneficiary", async () => {
      const expectedUnbonded = initialDeposit.sub(value)

      const expectedBeneficiaryBalance = web3.utils
        .toBN(await web3.eth.getBalance(beneficiary))
        .add(value)

      expect(await bonding.unbondedValue(operator)).to.eq.BN(
        initialDeposit,
        "invalid unbonded value"
      )

      await bonding.withdraw(value, operator, {from: operator})

      expect(await bonding.unbondedValue(operator)).to.eq.BN(
        expectedUnbonded,
        "invalid unbonded value"
      )

      const actualBeneficiaryBalance = await web3.eth.getBalance(beneficiary)
      expect(actualBeneficiaryBalance).to.eq.BN(
        expectedBeneficiaryBalance,
        "invalid beneficiary balance"
      )
    })

    it("emits event", async () => {
      const value = new BN(90)

      const receipt = await bonding.withdraw(value, operator, {
        from: operator,
      })
      expectEvent(receipt, "UnbondedValueWithdrawn", {
        operator: operator,
        beneficiary: beneficiary,
        amount: value,
      })
    })

    it("reverts if insufficient unbonded value", async () => {
      const invalidValue = initialDeposit.addn(1)

      await expectRevert(
        bonding.withdraw(invalidValue, operator, {from: operator}),
        "Insufficient unbonded value"
      )
    })

    it("reverts if transfer fails", async () => {
      const operator2 = accounts[7]

      await etherReceiver.setShouldFail(true)

      await bonding.delegate(operator2, etherReceiver.address, authorizer, {
        from: owner,
        value: initialDeposit,
      })

      await time.increase(delegationLockPeriod.addn(1))

      await expectRevert(
        bonding.withdraw(value, operator2, {from: operator2}),
        "Transfer failed"
      )
    })
  })

  describe("isInitialized", async () => {
    it("returns true when authorized and initialization period passed", async () => {
      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: minimumDelegationValue,
      })

      await bonding.authorizeOperatorContract(operator, bondCreator, {
        from: authorizer,
      })

      await time.increase(initializationPeriod.addn(1))

      assert.isTrue(await bonding.isInitialized(operator, bondCreator))
    })

    it("returns false when authorized but initialization period not passed yet", async () => {
      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: minimumDelegationValue,
      })

      await bonding.authorizeOperatorContract(operator, bondCreator, {
        from: authorizer,
      })

      await time.increase(initializationPeriod.subn(1))

      assert.isFalse(await bonding.isInitialized(operator, bondCreator))
    })

    it("returns false when initialization period passed but not authorized", async () => {
      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: minimumDelegationValue,
      })

      await time.increase(initializationPeriod.addn(1))

      assert.isFalse(await bonding.isInitialized(operator, bondCreator))
    })

    it("returns false when not authorized and initialization period not passed", async () => {
      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: minimumDelegationValue,
      })

      assert.isFalse(await bonding.isInitialized(operator, bondCreator))
    })

    it("returns false when initialization period passed but other contract authorized", async () => {
      await bonding.delegate(operator, beneficiary, authorizer, {
        from: owner,
        value: minimumDelegationValue,
      })

      await registry.approveOperatorContract(thirdParty)
      await bonding.authorizeOperatorContract(operator, thirdParty, {
        from: authorizer,
      })

      await time.increase(initializationPeriod.addn(1))

      assert.isFalse(await bonding.isInitialized(operator, bondCreator))
    })

    describe("getDelegationInfo", async () => {
      it("returns delegation details", async () => {
        await bonding.delegate(operator, beneficiary, authorizer, {
          from: owner,
          value: minimumDelegationValue,
        })

        const delegationInfo = await bonding.getDelegationInfo(operator)
        expect(delegationInfo.createdAt).to.eq.BN(await time.latest())
        expect(delegationInfo.undelegatedAt).to.eq.BN(0)
      })
    })
  })
})
