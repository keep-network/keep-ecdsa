const { accounts, contract, web3 } = require("@openzeppelin/test-environment")
const { createSnapshot, restoreSnapshot } = require("./helpers/snapshot")

const KeepRegistry = contract.fromArtifact("KeepRegistry")
const TokenStaking = contract.fromArtifact("TokenStakingStub")
const TokenGrant = contract.fromArtifact("TokenGrantStub")
const ManagedGrant = contract.fromArtifact("ManagedGrantStub")
const KeepBonding = contract.fromArtifact("KeepBonding")
const TestEtherReceiver = contract.fromArtifact("TestEtherReceiver")

const { expectEvent, expectRevert } = require("@openzeppelin/test-helpers")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

describe("KeepBonding", function () {
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

    await tokenStaking.setAuthorizer(operator, authorizer)
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

  describe("withdraw", async () => {
    const value = new BN(1000)

    beforeEach(async () => {
      await tokenStaking.setBeneficiary(operator, beneficiary)
      await keepBonding.deposit(operator, { value: value })
    })

    it("can be called by operator", async () => {
      const tokenOwner = accounts[2]
      await tokenStaking.setOwner(operator, tokenOwner)

      await keepBonding.withdraw(value, operator, { from: operator })
      // ok, no reverts
    })

    it("can be called by token owner", async () => {
      const tokenOwner = accounts[2]
      await tokenStaking.setOwner(operator, tokenOwner)

      await keepBonding.withdraw(value, operator, { from: tokenOwner })
      // ok, no reverts
    })

    it("can be called by grantee", async () => {
      const grantee = accounts[2]
      await tokenGrant.setGranteeOperator(grantee, operator)

      await keepBonding.withdraw(value, operator, { from: grantee })
      // ok, no reverts
    })

    it("cannot be called by third party", async () => {
      const thirdParty = accounts[2]

      await expectRevert(
        keepBonding.withdraw(value, operator, { from: thirdParty }),
        "Only operator or the owner is allowed to withdraw bond"
      )
    })

    it("transfers unbonded value to beneficiary", async () => {
      const expectedUnbonded = 0
      await tokenStaking.setBeneficiary(operator, beneficiary)
      const expectedBeneficiaryBalance = web3.utils
        .toBN(await web3.eth.getBalance(beneficiary))
        .add(value)

      await keepBonding.withdraw(value, operator, { from: operator })

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
        beneficiary: beneficiary,
        amount: value,
      })
    })

    it("reverts if insufficient unbonded value", async () => {
      const invalidValue = value.add(new BN(1))

      await expectRevert(
        keepBonding.withdraw(invalidValue, operator, { from: operator }),
        "Insufficient unbonded value"
      )
    })

    it("reverts if transfer fails", async () => {
      await etherReceiver.setShouldFail(true)
      await tokenStaking.setBeneficiary(operator, etherReceiver.address)

      await expectRevert(
        keepBonding.withdraw(value, operator, { from: operator }),
        "Transfer failed"
      )
    })
  })

  describe("withdrawAsManagedGrantee", async () => {
    const value = new BN(1000)
    const managedGrantee = accounts[2]
    let managedGrant

    beforeEach(async () => {
      await tokenStaking.setBeneficiary(operator, beneficiary)
      await keepBonding.deposit(operator, { value: value })

      managedGrant = await ManagedGrant.new(managedGrantee)
      await tokenGrant.setGranteeOperator(managedGrant.address, operator)
    })

    it("can be called by managed grantee", async () => {
      await keepBonding.withdrawAsManagedGrantee(
        value,
        operator,
        managedGrant.address,
        { from: managedGrantee }
      )
      // ok, no reverts
    })

    it("cannot be called by operator", async () => {
      await expectRevert(
        keepBonding.withdrawAsManagedGrantee(
          value,
          operator,
          managedGrant.address,
          { from: operator }
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
          { from: tokenOwner }
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
          { from: standardGrantee }
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
          { from: thirdParty }
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
        { from: managedGrantee }
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
        { from: managedGrantee }
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
          { from: managedGrantee }
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
          { from: managedGrantee }
        ),
        "Transfer failed"
      )
    })
  })
})
