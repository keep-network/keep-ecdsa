const { accounts, contract, web3 } = require("@openzeppelin/test-environment")
const { createSnapshot, restoreSnapshot } = require("./helpers/snapshot")

const KeepRegistry = contract.fromArtifact("KeepRegistry")
const StakingInfoStub = contract.fromArtifact("StakingInfoStub")
const AbstractBonding = contract.fromArtifact("AbstractBondingStub")
const TestEtherReceiver = contract.fromArtifact("TestEtherReceiver")

const {
  constants,
  expectEvent,
  expectRevert,
} = require("@openzeppelin/test-helpers")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect
const assert = chai.assert

describe("AbstractBonding", function () {
  let registry
  let abstractBonding
  let etherReceiver

  let operator
  let authorizer
  let bondCreator
  let sortitionPool
  let beneficiary

  before(async () => {
    operator = accounts[1]
    authorizer = accounts[2]
    beneficiary = accounts[3]
    bondCreator = accounts[4]
    sortitionPool = accounts[5]

    registry = await KeepRegistry.new()
    stakingInfoStub = await StakingInfoStub.new()

    abstractBonding = await AbstractBonding.new(
      registry.address,
      stakingInfoStub.address
    )

    etherReceiver = await TestEtherReceiver.new()

    await registry.approveOperatorContract(bondCreator)

    await stakingInfoStub.setAuthorizer(operator, authorizer)
    await abstractBonding.authorizeSortitionPoolContract(
      operator,
      sortitionPool,
      {
        from: authorizer,
      }
    )

    await stakingInfoStub.authorizeOperatorContract(operator, bondCreator)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("deposit", async () => {
    const value = new BN(100)
    const expectedUnbonded = value

    it("registers unbonded value", async () => {
      await stakingInfoStub.setBeneficiary(operator, beneficiary)
      await abstractBonding.deposit(operator, { value: value })
      const unbonded = await abstractBonding.availableUnbondedValue(
        operator,
        bondCreator,
        sortitionPool
      )

      expect(unbonded).to.eq.BN(expectedUnbonded, "invalid unbonded value")
    })

    it("sums deposits", async () => {
      const value1 = value
      const value2 = new BN(230)

      await stakingInfoStub.setBeneficiary(operator, beneficiary)

      await abstractBonding.deposit(operator, { value: value1 })
      expect(
        await abstractBonding.availableUnbondedValue(
          operator,
          bondCreator,
          sortitionPool
        )
      ).to.eq.BN(value1, "invalid unbonded value after first deposit")

      await abstractBonding.deposit(operator, { value: value2 })
      expect(
        await abstractBonding.availableUnbondedValue(
          operator,
          bondCreator,
          sortitionPool
        )
      ).to.eq.BN(
        value1.add(value2),
        "invalid unbonded value after second deposit"
      )
    })

    it("emits event", async () => {
      await stakingInfoStub.setBeneficiary(operator, beneficiary)
      const receipt = await abstractBonding.deposit(operator, { value: value })
      expectEvent(receipt, "UnbondedValueDeposited", {
        operator: operator,
        beneficiary: beneficiary,
        amount: value,
      })
    })

    it("reverts if beneficiary is not defined", async () => {
      await stakingInfoStub.setBeneficiary(operator, constants.ZERO_ADDRESS)

      await expectRevert(
        abstractBonding.deposit(operator, { value: value }),
        "Beneficiary not defined for the operator"
      )
    })
  })

  describe("availableUnbondedValue", async () => {
    const value = new BN(100)

    beforeEach(async () => {
      await stakingInfoStub.setBeneficiary(operator, beneficiary)
      await abstractBonding.deposit(operator, { value: value })
    })

    it("returns zero for operator with no deposit", async () => {
      const unbondedOperator = "0x0000000000000000000000000000000000000001"
      const expectedUnbonded = 0

      const unbondedValue = await abstractBonding.availableUnbondedValue(
        unbondedOperator,
        bondCreator,
        sortitionPool
      )
      expect(unbondedValue).to.eq.BN(expectedUnbonded, "invalid unbonded value")
    })

    it("return zero when bond creator is not approved by operator", async () => {
      const notApprovedBondCreator =
        "0x0000000000000000000000000000000000000001"
      const expectedUnbonded = 0

      const unbondedValue = await abstractBonding.availableUnbondedValue(
        operator,
        notApprovedBondCreator,
        sortitionPool
      )
      expect(unbondedValue).to.eq.BN(expectedUnbonded, "invalid unbonded value")
    })

    it("returns zero when sortition pool is not authorized", async () => {
      const notAuthorizedSortitionPool =
        "0x0000000000000000000000000000000000000001"
      const expectedUnbonded = 0

      const unbondedValue = await abstractBonding.availableUnbondedValue(
        operator,
        bondCreator,
        notAuthorizedSortitionPool
      )
      expect(unbondedValue).to.eq.BN(expectedUnbonded, "invalid unbonded value")
    })

    it("returns value of operators deposit", async () => {
      const expectedUnbonded = value

      const unbonded = await abstractBonding.availableUnbondedValue(
        operator,
        bondCreator,
        sortitionPool
      )

      expect(unbonded).to.eq.BN(expectedUnbonded, "invalid unbonded value")
    })
  })

  describe("createBond", async () => {
    const holder = accounts[3]
    const value = new BN(100)

    beforeEach(async () => {
      await stakingInfoStub.setBeneficiary(operator, beneficiary)
      await abstractBonding.deposit(operator, { value: value })
    })

    it("creates bond", async () => {
      const reference = new BN(888)

      const expectedUnbonded = 0

      await abstractBonding.createBond(
        operator,
        holder,
        reference,
        value,
        sortitionPool,
        { from: bondCreator }
      )

      const unbonded = await abstractBonding.availableUnbondedValue(
        operator,
        bondCreator,
        sortitionPool
      )
      expect(unbonded).to.eq.BN(expectedUnbonded, "invalid unbonded value")

      const lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(value, "unexpected bond value")
    })

    it("emits event", async () => {
      const reference = new BN(999)

      const receipt = await abstractBonding.createBond(
        operator,
        holder,
        reference,
        value,
        sortitionPool,
        { from: bondCreator }
      )

      expectEvent(receipt, "BondCreated", {
        operator: operator,
        holder: holder,
        sortitionPool: sortitionPool,
        referenceID: reference,
        amount: value,
      })
    })

    it("creates two bonds with the same reference for different operators", async () => {
      const operator2 = accounts[6]
      const authorizer2 = accounts[7]
      const bondValue = new BN(10)
      const reference = new BN(777)

      const expectedUnbonded = value.sub(bondValue)

      await stakingInfoStub.setBeneficiary(operator2, etherReceiver.address)
      await stakingInfoStub.setAuthorizer(operator2, authorizer2)

      await abstractBonding.deposit(operator2, { value: value })

      await stakingInfoStub.authorizeOperatorContract(operator2, bondCreator)
      await abstractBonding.authorizeSortitionPoolContract(
        operator2,
        sortitionPool,
        { from: authorizer2 }
      )
      await abstractBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        { from: bondCreator }
      )
      await abstractBonding.createBond(
        operator2,
        holder,
        reference,
        bondValue,
        sortitionPool,
        { from: bondCreator }
      )

      const unbonded1 = await abstractBonding.availableUnbondedValue(
        operator,
        bondCreator,
        sortitionPool
      )
      expect(unbonded1).to.eq.BN(expectedUnbonded, "invalid unbonded value 1")

      const unbonded2 = await abstractBonding.availableUnbondedValue(
        operator2,
        bondCreator,
        sortitionPool
      )
      expect(unbonded2).to.eq.BN(expectedUnbonded, "invalid unbonded value 2")

      const lockedBonds1 = await abstractBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds1).to.eq.BN(bondValue, "unexpected bond value 1")

      const lockedBonds2 = await abstractBonding.bondAmount(
        operator2,
        holder,
        reference
      )
      expect(lockedBonds2).to.eq.BN(bondValue, "unexpected bond value 2")
    })

    it("fails to create two bonds with the same reference for the same operator", async () => {
      const bondValue = new BN(10)
      const reference = new BN(777)

      await abstractBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        { from: bondCreator }
      )

      await expectRevert(
        abstractBonding.createBond(
          operator,
          holder,
          reference,
          bondValue,
          sortitionPool,
          { from: bondCreator }
        ),
        "Reference ID not unique for holder and operator"
      )
    })

    it("fails if insufficient unbonded value", async () => {
      const bondValue = value.add(new BN(1))

      await expectRevert(
        abstractBonding.createBond(
          operator,
          holder,
          0,
          bondValue,
          sortitionPool,
          {
            from: bondCreator,
          }
        ),
        "Insufficient unbonded value"
      )
    })
  })

  describe("reassignBond", async () => {
    const holder = accounts[6]
    const newHolder = accounts[3]
    const bondValue = new BN(100)
    const reference = new BN(777)
    const newReference = new BN(888)

    beforeEach(async () => {
      await stakingInfoStub.setBeneficiary(operator, beneficiary)
      await abstractBonding.deposit(operator, { value: bondValue })
      await abstractBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        { from: bondCreator }
      )
    })

    it("reassigns bond to a new holder and a new reference", async () => {
      await abstractBonding.reassignBond(
        operator,
        reference,
        newHolder,
        newReference,
        { from: holder }
      )

      let lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        newReference
      )
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await abstractBonding.bondAmount(
        operator,
        newHolder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await abstractBonding.bondAmount(
        operator,
        newHolder,
        newReference
      )
      expect(lockedBonds).to.eq.BN(bondValue, "invalid locked bonds")
    })

    it("reassigns bond to the same holder and a new reference", async () => {
      await abstractBonding.reassignBond(
        operator,
        reference,
        holder,
        newReference,
        { from: holder }
      )

      let lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        newReference
      )
      expect(lockedBonds).to.eq.BN(bondValue, "invalid locked bonds")
    })

    it("reassigns bond to a new holder and the same reference", async () => {
      await abstractBonding.reassignBond(
        operator,
        reference,
        newHolder,
        reference,
        { from: holder }
      )

      let lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await abstractBonding.bondAmount(
        operator,
        newHolder,
        reference
      )
      expect(lockedBonds).to.eq.BN(bondValue, "invalid locked bonds")
    })

    it("emits event", async () => {
      const receipt = await abstractBonding.reassignBond(
        operator,
        reference,
        newHolder,
        newReference,
        { from: holder }
      )

      expectEvent(receipt, "BondReassigned", {
        operator: operator,
        referenceID: reference,
        newHolder: newHolder,
        newReferenceID: newReference,
      })
    })

    it("fails if sender is not the holder", async () => {
      await expectRevert(
        abstractBonding.reassignBond(
          operator,
          reference,
          newHolder,
          newReference,
          {
            from: accounts[0],
          }
        ),
        "Bond not found"
      )
    })

    it("fails if reassigned to the same holder and the same reference", async () => {
      await abstractBonding.deposit(operator, { value: bondValue })
      await abstractBonding.createBond(
        operator,
        holder,
        newReference,
        bondValue,
        sortitionPool,
        { from: bondCreator }
      )

      await expectRevert(
        abstractBonding.reassignBond(
          operator,
          reference,
          holder,
          newReference,
          {
            from: holder,
          }
        ),
        "Reference ID not unique for holder and operator"
      )
    })
  })

  describe("freeBond", async () => {
    const holder = accounts[6]
    const initialUnboundedValue = new BN(500)
    const bondValue = new BN(100)
    const reference = new BN(777)

    beforeEach(async () => {
      await stakingInfoStub.setBeneficiary(operator, beneficiary)
      await abstractBonding.deposit(operator, { value: initialUnboundedValue })
      await abstractBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        { from: bondCreator }
      )
    })

    it("releases bond amount to operator's available bonding value", async () => {
      await abstractBonding.freeBond(operator, reference, { from: holder })

      const lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "unexpected remaining locked bonds")

      const unbondedValue = await abstractBonding.availableUnbondedValue(
        operator,
        bondCreator,
        sortitionPool
      )
      expect(unbondedValue).to.eq.BN(
        initialUnboundedValue,
        "unexpected unbonded value"
      )
    })

    it("emits event", async () => {
      const receipt = await abstractBonding.freeBond(operator, reference, {
        from: holder,
      })

      expectEvent(receipt, "BondReleased", {
        operator: operator,
        referenceID: reference,
      })
    })

    it("fails if sender is not the holder", async () => {
      await expectRevert(
        abstractBonding.freeBond(operator, reference, { from: accounts[0] }),
        "Bond not found"
      )
    })
  })

  describe("seizeBond", async () => {
    const holder = accounts[6]
    const destination = accounts[3]
    const bondValue = new BN(1000)
    const reference = new BN(777)

    beforeEach(async () => {
      await stakingInfoStub.setBeneficiary(operator, beneficiary)
      await abstractBonding.deposit(operator, { value: bondValue })
      await abstractBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        { from: bondCreator }
      )
    })

    it("transfers whole bond amount to destination account", async () => {
      const amount = bondValue
      const expectedBalance = web3.utils
        .toBN(await web3.eth.getBalance(destination))
        .add(amount)

      await abstractBonding.seizeBond(
        operator,
        reference,
        amount,
        destination,
        {
          from: holder,
        }
      )

      const actualBalance = await web3.eth.getBalance(destination)
      expect(actualBalance).to.eq.BN(
        expectedBalance,
        "invalid destination account balance"
      )

      const lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "unexpected remaining bond value")
    })

    it("emits event", async () => {
      const amount = new BN(80)

      const receipt = await abstractBonding.seizeBond(
        operator,
        reference,
        amount,
        destination,
        {
          from: holder,
        }
      )

      expectEvent(receipt, "BondSeized", {
        operator: operator,
        referenceID: reference,
        destination: destination,
        amount: amount,
      })
    })

    it("transfers less than bond amount to destination account", async () => {
      const remainingBond = new BN(1)
      const amount = bondValue.sub(remainingBond)
      const expectedBalance = web3.utils
        .toBN(await web3.eth.getBalance(destination))
        .add(amount)

      await abstractBonding.seizeBond(
        operator,
        reference,
        amount,
        destination,
        {
          from: holder,
        }
      )

      const actualBalance = await web3.eth.getBalance(destination)
      expect(actualBalance).to.eq.BN(
        expectedBalance,
        "invalid destination account balance"
      )

      const lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(
        remainingBond,
        "unexpected remaining bond value"
      )
    })

    it("reverts if seized amount equals zero", async () => {
      const amount = new BN(0)
      await expectRevert(
        abstractBonding.seizeBond(operator, reference, amount, destination, {
          from: holder,
        }),
        "Requested amount should be greater than zero"
      )
    })

    it("reverts if seized amount is greater than bond value", async () => {
      const amount = bondValue.add(new BN(1))
      await expectRevert(
        abstractBonding.seizeBond(operator, reference, amount, destination, {
          from: holder,
        }),
        "Requested amount is greater than the bond"
      )
    })

    it("reverts if transfer fails", async () => {
      await etherReceiver.setShouldFail(true)
      const destination = etherReceiver.address

      await expectRevert(
        abstractBonding.seizeBond(operator, reference, bondValue, destination, {
          from: holder,
        }),
        "Transfer failed"
      )

      const destinationBalance = await web3.eth.getBalance(destination)
      expect(destinationBalance).to.eq.BN(
        0,
        "invalid destination account balance"
      )

      const lockedBonds = await abstractBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(bondValue, "unexpected bond value")
    })
  })

  describe("authorizeSortitionPoolContract", async () => {
    it("reverts when operator is not an authorizer", async () => {
      const authorizer1 = accounts[6]

      await expectRevert(
        abstractBonding.authorizeSortitionPoolContract(
          operator,
          sortitionPool,
          {
            from: authorizer1,
          }
        ),
        "Not authorized"
      )
    })

    it("should authorize sortition pool for the provided operator", async () => {
      await abstractBonding.authorizeSortitionPoolContract(
        operator,
        sortitionPool,
        { from: authorizer }
      )

      assert.isTrue(
        await abstractBonding.hasSecondaryAuthorization(
          operator,
          sortitionPool
        ),
        "Sortition pool should be authorized for the provided operator"
      )
    })
  })

  describe("deauthorizeSortitionPoolContract", async () => {
    it("reverts when operator is not an authorizer", async () => {
      const authorizer1 = accounts[6]

      await expectRevert(
        abstractBonding.deauthorizeSortitionPoolContract(
          operator,
          sortitionPool,
          {
            from: authorizer1,
          }
        ),
        "Not authorized"
      )
    })

    it("should deauthorize sortition pool for the provided operator", async () => {
      await abstractBonding.authorizeSortitionPoolContract(
        operator,
        sortitionPool,
        { from: authorizer }
      )
      await abstractBonding.deauthorizeSortitionPoolContract(
        operator,
        sortitionPool,
        { from: authorizer }
      )
      assert.isFalse(
        await abstractBonding.hasSecondaryAuthorization(
          operator,
          sortitionPool
        ),
        "Sortition pool should be deauthorized for the provided operator"
      )
    })

    describe("withdrawBond", async () => {
      const value = new BN(1000)

      beforeEach(async () => {
        await stakingInfoStub.setBeneficiary(operator, beneficiary)
        await abstractBonding.deposit(operator, { value: value })
      })

      it("transfers unbonded value to beneficiary", async () => {
        const expectedUnbonded = 0
        await stakingInfoStub.setBeneficiary(operator, beneficiary)
        const expectedBeneficiaryBalance = web3.utils
          .toBN(await web3.eth.getBalance(beneficiary))
          .add(value)

        await abstractBonding.withdrawBondExposed(value, operator, {
          from: operator,
        })

        const unbonded = await abstractBonding.availableUnbondedValue(
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

        const receipt = await abstractBonding.withdrawBondExposed(
          value,
          operator,
          {
            from: operator,
          }
        )
        expectEvent(receipt, "UnbondedValueWithdrawn", {
          operator: operator,
          beneficiary: beneficiary,
          amount: value,
        })
      })

      it("reverts if insufficient unbonded value", async () => {
        const invalidValue = value.add(new BN(1))

        await expectRevert(
          abstractBonding.withdrawBondExposed(invalidValue, operator, {
            from: operator,
          }),
          "Insufficient unbonded value"
        )
      })

      it("reverts if transfer fails", async () => {
        await etherReceiver.setShouldFail(true)
        await stakingInfoStub.setBeneficiary(operator, etherReceiver.address)

        await expectRevert(
          abstractBonding.withdrawBondExposed(value, operator, {
            from: operator,
          }),
          "Transfer failed"
        )
      })
    })
  })
})
