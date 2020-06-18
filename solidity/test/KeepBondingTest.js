import {createSnapshot, restoreSnapshot} from "./helpers/snapshot"

const KeepRegistry = artifacts.require("./KeepRegistry.sol")
const TokenStaking = artifacts.require("./TokenStakingStub.sol")
const TokenGrant = artifacts.require("./TokenGrantStub.sol")
const ManagedGrant = artifacts.require("./ManagedGrantStub.sol")
const KeepBonding = artifacts.require("./KeepBonding.sol")
const TestEtherReceiver = artifacts.require("./TestEtherReceiver.sol")

const {expectEvent, expectRevert} = require("@openzeppelin/test-helpers")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

contract("KeepBonding", (accounts) => {
  let registry
  let tokenStaking
  let tokenGrant
  let keepBonding
  let etherReceiver

  let operator
  let authorizer
  let bondCreator
  let sortitionPool
  let beneficiary

  before(async () => {
    operator = accounts[1]
    authorizer = operator
    beneficiary = accounts[3]
    bondCreator = accounts[4]
    sortitionPool = accounts[5]

    registry = await KeepRegistry.new()
    tokenStaking = await TokenStaking.new()
    tokenGrant = await TokenGrant.new()
    keepBonding = await KeepBonding.new(
      registry.address,
      tokenStaking.address,
      tokenGrant.address
    )
    etherReceiver = await TestEtherReceiver.new()

    await registry.approveOperatorContract(bondCreator)
    await keepBonding.authorizeSortitionPoolContract(operator, sortitionPool, {
      from: authorizer,
    })

    await tokenStaking.authorizeOperatorContract(operator, bondCreator)
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
      await keepBonding.deposit(operator, {value: value})

      const unbonded = await keepBonding.availableUnbondedValue(
        operator,
        bondCreator,
        sortitionPool
      )

      expect(unbonded).to.eq.BN(expectedUnbonded, "invalid unbonded value")
    })

    it("emits event", async () => {
      const receipt = await keepBonding.deposit(operator, {value: value})
      expectEvent(receipt, "UnbondedValueDeposited", {
        operator: operator,
        amount: value,
      })
    })
  })

  describe("withdraw", async () => {
    const value = new BN(1000)

    beforeEach(async () => {
      await keepBonding.deposit(operator, {value: value})
    })

    it("can be called by operator", async () => {
      const tokenOwner = accounts[2]
      await tokenStaking.setOwner(operator, tokenOwner)

      await keepBonding.withdraw(value, operator, {from: operator})
      // ok, no reverts
    })

    it("can be called by token owner", async () => {
      const tokenOwner = accounts[2]
      await tokenStaking.setOwner(operator, tokenOwner)

      await keepBonding.withdraw(value, operator, {from: tokenOwner})
      // ok, no reverts
    })

    it("can be called by grantee", async () => {
      const grantee = accounts[2]
      await tokenGrant.setGranteeOperator(grantee, operator)

      await keepBonding.withdraw(value, operator, {from: grantee})
      // ok, no reverts
    })

    it("cannot be called by third party", async () => {
      const thirdParty = accounts[2]

      await expectRevert(
        keepBonding.withdraw(value, operator, {from: thirdParty}),
        "Only operator or the owner is allowed to withdraw bond"
      )
    })

    it("transfers unbonded value to beneficiary", async () => {
      const expectedUnbonded = 0
      await tokenStaking.setBeneficiary(operator, beneficiary)
      const expectedBeneficiaryBalance = web3.utils
        .toBN(await web3.eth.getBalance(beneficiary))
        .add(value)

      await keepBonding.withdraw(value, operator, {from: operator})

      const unbonded = await keepBonding.availableUnbondedValue(
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

      const receipt = await keepBonding.withdraw(value, operator, {
        from: operator,
      })
      expectEvent(receipt, "UnbondedValueWithdrawn", {
        operator: operator,
        amount: value,
      })
    })

    it("reverts if insufficient unbonded value", async () => {
      const invalidValue = value.add(new BN(1))

      await expectRevert(
        keepBonding.withdraw(invalidValue, operator, {from: operator}),
        "Insufficient unbonded value"
      )
    })

    it("reverts if transfer fails", async () => {
      await etherReceiver.setShouldFail(true)
      await tokenStaking.setBeneficiary(operator, etherReceiver.address)

      await expectRevert(
        keepBonding.withdraw(value, operator, {from: operator}),
        "Transfer failed"
      )
    })
  })

  describe("withdrawAsManagedGrantee", async () => {
    const value = new BN(1000)
    const managedGrantee = accounts[2]
    let managedGrant

    beforeEach(async () => {
      await keepBonding.deposit(operator, {value: value})

      managedGrant = await ManagedGrant.new(managedGrantee)
      await tokenGrant.setGranteeOperator(managedGrant.address, operator)
    })

    it("can be called by managed grantee", async () => {
      await keepBonding.withdrawAsManagedGrantee(
        value,
        operator,
        managedGrant.address,
        {from: managedGrantee}
      )
      // ok, no reverts
    })

    it("cannot be called by operator", async () => {
      await expectRevert(
        keepBonding.withdrawAsManagedGrantee(
          value,
          operator,
          managedGrant.address,
          {from: operator}
        ),
        "Not a grantee of the provided contract"
      )
    })

    it("cannot be called by token owner", async () => {
      const tokenOwner = accounts[0]
      await tokenStaking.setOwner(operator, tokenOwner)

      await expectRevert(
        keepBonding.withdrawAsManagedGrantee(
          value,
          operator,
          managedGrant.address,
          {from: tokenOwner}
        ),
        "Not a grantee of the provided contract"
      )
    })

    it("cannot be called by a standard grantee", async () => {
      const standardGrantee = accounts[0]
      await tokenGrant.setGranteeOperator(standardGrantee, operator)

      await expectRevert(
        keepBonding.withdrawAsManagedGrantee(
          value,
          operator,
          managedGrant.address,
          {from: standardGrantee}
        ),
        "Not a grantee of the provided contract"
      )
    })

    it("cannot be called by third party", async () => {
      const thirdParty = accounts[0]

      await expectRevert(
        keepBonding.withdrawAsManagedGrantee(
          value,
          operator,
          managedGrant.address,
          {from: thirdParty}
        ),
        "Not a grantee of the provided contract"
      )
    })

    it("transfers unbonded value to beneficiary", async () => {
      const expectedUnbonded = 0
      await tokenStaking.setBeneficiary(operator, beneficiary)
      const expectedBeneficiaryBalance = web3.utils
        .toBN(await web3.eth.getBalance(beneficiary))
        .add(value)

      await keepBonding.withdrawAsManagedGrantee(
        value,
        operator,
        managedGrant.address,
        {from: managedGrantee}
      )

      const unbonded = await keepBonding.availableUnbondedValue(
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

      const receipt = await keepBonding.withdrawAsManagedGrantee(
        value,
        operator,
        managedGrant.address,
        {from: managedGrantee}
      )
      expectEvent(receipt, "UnbondedValueWithdrawn", {
        operator: operator,
        amount: value,
      })
    })

    it("reverts if insufficient unbonded value", async () => {
      const invalidValue = value.add(new BN(1))

      await expectRevert(
        keepBonding.withdrawAsManagedGrantee(
          invalidValue,
          operator,
          managedGrant.address,
          {from: managedGrantee}
        ),
        "Insufficient unbonded value"
      )
    })

    it("reverts if transfer fails", async () => {
      await etherReceiver.setShouldFail(true)
      await tokenStaking.setBeneficiary(operator, etherReceiver.address)

      await expectRevert(
        keepBonding.withdrawAsManagedGrantee(
          value,
          operator,
          managedGrant.address,
          {from: managedGrantee}
        ),
        "Transfer failed"
      )
    })
  })

  describe("availableUnbondedValue", async () => {
    const value = new BN(100)

    beforeEach(async () => {
      await keepBonding.deposit(operator, {value: value})
    })

    it("returns zero for operator with no deposit", async () => {
      const unbondedOperator = "0x0000000000000000000000000000000000000001"
      const expectedUnbonded = 0

      const unbondedValue = await keepBonding.availableUnbondedValue(
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

      const unbondedValue = await keepBonding.availableUnbondedValue(
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

      const unbondedValue = await keepBonding.availableUnbondedValue(
        operator,
        bondCreator,
        notAuthorizedSortitionPool
      )
      expect(unbondedValue).to.eq.BN(expectedUnbonded, "invalid unbonded value")
    })

    it("returns value of operators deposit", async () => {
      const expectedUnbonded = value

      const unbonded = await keepBonding.availableUnbondedValue(
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
      await keepBonding.deposit(operator, {value: value})
    })

    it("creates bond", async () => {
      const reference = new BN(888)

      const expectedUnbonded = 0

      await keepBonding.createBond(
        operator,
        holder,
        reference,
        value,
        sortitionPool,
        {from: bondCreator}
      )

      const unbonded = await keepBonding.availableUnbondedValue(
        operator,
        bondCreator,
        sortitionPool
      )
      expect(unbonded).to.eq.BN(expectedUnbonded, "invalid unbonded value")

      const lockedBonds = await keepBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(value, "unexpected bond value")
    })

    it("emits event", async () => {
      const reference = new BN(999)

      const receipt = await keepBonding.createBond(
        operator,
        holder,
        reference,
        value,
        sortitionPool,
        {from: bondCreator}
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
      const operator2 = accounts[2]
      const authorizer2 = accounts[2]
      const bondValue = new BN(10)
      const reference = new BN(777)

      const expectedUnbonded = value.sub(bondValue)

      await keepBonding.deposit(operator2, {value: value})

      await tokenStaking.authorizeOperatorContract(operator2, bondCreator)
      await keepBonding.authorizeSortitionPoolContract(
        operator2,
        sortitionPool,
        {from: authorizer2}
      )
      await keepBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        {from: bondCreator}
      )
      await keepBonding.createBond(
        operator2,
        holder,
        reference,
        bondValue,
        sortitionPool,
        {from: bondCreator}
      )

      const unbonded1 = await keepBonding.availableUnbondedValue(
        operator,
        bondCreator,
        sortitionPool
      )
      expect(unbonded1).to.eq.BN(expectedUnbonded, "invalid unbonded value 1")

      const unbonded2 = await keepBonding.availableUnbondedValue(
        operator2,
        bondCreator,
        sortitionPool
      )
      expect(unbonded2).to.eq.BN(expectedUnbonded, "invalid unbonded value 2")

      const lockedBonds1 = await keepBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds1).to.eq.BN(bondValue, "unexpected bond value 1")

      const lockedBonds2 = await keepBonding.bondAmount(
        operator2,
        holder,
        reference
      )
      expect(lockedBonds2).to.eq.BN(bondValue, "unexpected bond value 2")
    })

    it("fails to create two bonds with the same reference for the same operator", async () => {
      const bondValue = new BN(10)
      const reference = new BN(777)

      await keepBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        {from: bondCreator}
      )

      await expectRevert(
        keepBonding.createBond(
          operator,
          holder,
          reference,
          bondValue,
          sortitionPool,
          {from: bondCreator}
        ),
        "Reference ID not unique for holder and operator"
      )
    })

    it("fails if insufficient unbonded value", async () => {
      const bondValue = value.add(new BN(1))

      await expectRevert(
        keepBonding.createBond(operator, holder, 0, bondValue, sortitionPool, {
          from: bondCreator,
        }),
        "Insufficient unbonded value"
      )
    })
  })

  describe("reassignBond", async () => {
    const holder = accounts[2]
    const newHolder = accounts[3]
    const bondValue = new BN(100)
    const reference = new BN(777)
    const newReference = new BN(888)

    beforeEach(async () => {
      await keepBonding.deposit(operator, {value: bondValue})
      await keepBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        {from: bondCreator}
      )
    })

    it("reassigns bond to a new holder and a new reference", async () => {
      await keepBonding.reassignBond(
        operator,
        reference,
        newHolder,
        newReference,
        {from: holder}
      )

      let lockedBonds = await keepBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await keepBonding.bondAmount(operator, holder, newReference)
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await keepBonding.bondAmount(operator, newHolder, reference)
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await keepBonding.bondAmount(
        operator,
        newHolder,
        newReference
      )
      expect(lockedBonds).to.eq.BN(bondValue, "invalid locked bonds")
    })

    it("reassigns bond to the same holder and a new reference", async () => {
      await keepBonding.reassignBond(
        operator,
        reference,
        holder,
        newReference,
        {from: holder}
      )

      let lockedBonds = await keepBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await keepBonding.bondAmount(operator, holder, newReference)
      expect(lockedBonds).to.eq.BN(bondValue, "invalid locked bonds")
    })

    it("reassigns bond to a new holder and the same reference", async () => {
      await keepBonding.reassignBond(
        operator,
        reference,
        newHolder,
        reference,
        {from: holder}
      )

      let lockedBonds = await keepBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "invalid locked bonds")

      lockedBonds = await keepBonding.bondAmount(operator, newHolder, reference)
      expect(lockedBonds).to.eq.BN(bondValue, "invalid locked bonds")
    })

    it("emits event", async () => {
      const receipt = await keepBonding.reassignBond(
        operator,
        reference,
        newHolder,
        newReference,
        {from: holder}
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
        keepBonding.reassignBond(operator, reference, newHolder, newReference, {
          from: accounts[0],
        }),
        "Bond not found"
      )
    })

    it("fails if reassigned to the same holder and the same reference", async () => {
      await keepBonding.deposit(operator, {value: bondValue})
      await keepBonding.createBond(
        operator,
        holder,
        newReference,
        bondValue,
        sortitionPool,
        {from: bondCreator}
      )

      await expectRevert(
        keepBonding.reassignBond(operator, reference, holder, newReference, {
          from: holder,
        }),
        "Reference ID not unique for holder and operator"
      )
    })
  })

  describe("freeBond", async () => {
    const holder = accounts[2]
    const initialUnboundedValue = new BN(500)
    const bondValue = new BN(100)
    const reference = new BN(777)

    beforeEach(async () => {
      await keepBonding.deposit(operator, {value: initialUnboundedValue})
      await keepBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        {from: bondCreator}
      )
    })

    it("releases bond amount to operator's available bonding value", async () => {
      await keepBonding.freeBond(operator, reference, {from: holder})

      const lockedBonds = await keepBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "unexpected remaining locked bonds")

      const unbondedValue = await keepBonding.availableUnbondedValue(
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
      const receipt = await keepBonding.freeBond(operator, reference, {
        from: holder,
      })

      expectEvent(receipt, "BondReleased", {
        operator: operator,
        referenceID: reference,
      })
    })

    it("fails if sender is not the holder", async () => {
      await expectRevert(
        keepBonding.freeBond(operator, reference, {from: accounts[0]}),
        "Bond not found"
      )
    })
  })

  describe("seizeBond", async () => {
    const holder = accounts[2]
    const destination = accounts[3]
    const bondValue = new BN(1000)
    const reference = new BN(777)

    beforeEach(async () => {
      await keepBonding.deposit(operator, {value: bondValue})
      await keepBonding.createBond(
        operator,
        holder,
        reference,
        bondValue,
        sortitionPool,
        {from: bondCreator}
      )
    })

    it("transfers whole bond amount to destination account", async () => {
      const amount = bondValue
      const expectedBalance = web3.utils
        .toBN(await web3.eth.getBalance(destination))
        .add(amount)

      await keepBonding.seizeBond(operator, reference, amount, destination, {
        from: holder,
      })

      const actualBalance = await web3.eth.getBalance(destination)
      expect(actualBalance).to.eq.BN(
        expectedBalance,
        "invalid destination account balance"
      )

      const lockedBonds = await keepBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(0, "unexpected remaining bond value")
    })

    it("emits event", async () => {
      const amount = new BN(80)

      const receipt = await keepBonding.seizeBond(
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

      await keepBonding.seizeBond(operator, reference, amount, destination, {
        from: holder,
      })

      const actualBalance = await web3.eth.getBalance(destination)
      expect(actualBalance).to.eq.BN(
        expectedBalance,
        "invalid destination account balance"
      )

      const lockedBonds = await keepBonding.bondAmount(
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
        keepBonding.seizeBond(operator, reference, amount, destination, {
          from: holder,
        }),
        "Requested amount should be greater than zero"
      )
    })

    it("reverts if seized amount is greater than bond value", async () => {
      const amount = bondValue.add(new BN(1))
      await expectRevert(
        keepBonding.seizeBond(operator, reference, amount, destination, {
          from: holder,
        }),
        "Requested amount is greater than the bond"
      )
    })

    it("reverts if transfer fails", async () => {
      await etherReceiver.setShouldFail(true)
      const destination = etherReceiver.address

      await expectRevert(
        keepBonding.seizeBond(operator, reference, bondValue, destination, {
          from: holder,
        }),
        "Transfer failed"
      )

      const destinationBalance = await web3.eth.getBalance(destination)
      expect(destinationBalance).to.eq.BN(
        0,
        "invalid destination account balance"
      )

      const lockedBonds = await keepBonding.bondAmount(
        operator,
        holder,
        reference
      )
      expect(lockedBonds).to.eq.BN(bondValue, "unexpected bond value")
    })
  })

  describe("authorizeSortitionPoolContract", async () => {
    it("reverts when operator is not an authorizer", async () => {
      const authorizer1 = accounts[2]

      await expectRevert(
        keepBonding.authorizeSortitionPoolContract(operator, sortitionPool, {
          from: authorizer1,
        }),
        "Not authorized"
      )
    })

    it("should authorize sortition pool for the provided operator", async () => {
      await keepBonding.authorizeSortitionPoolContract(
        operator,
        sortitionPool,
        {from: authorizer}
      )

      assert.isTrue(
        await keepBonding.hasSecondaryAuthorization(operator, sortitionPool),
        "Sortition pool should be authorized for the provided operator"
      )
    })
  })

  describe("deauthorizeSortitionPoolContract", async () => {
    it("reverts when operator is not an authorizer", async () => {
      const authorizer1 = accounts[2]

      await expectRevert(
        keepBonding.deauthorizeSortitionPoolContract(operator, sortitionPool, {
          from: authorizer1,
        }),
        "Not authorized"
      )
    })

    it("should deauthorize sortition pool for the provided operator", async () => {
      await keepBonding.authorizeSortitionPoolContract(
        operator,
        sortitionPool,
        {from: authorizer}
      )
      await keepBonding.deauthorizeSortitionPoolContract(
        operator,
        sortitionPool,
        {from: authorizer}
      )
      assert.isFalse(
        await keepBonding.hasSecondaryAuthorization(operator, sortitionPool),
        "Sortition pool should be deauthorized for the provided operator"
      )
    })
  })
})
