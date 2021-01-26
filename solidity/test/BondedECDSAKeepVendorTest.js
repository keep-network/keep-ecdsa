const { accounts, contract } = require("@openzeppelin/test-environment")
const { createSnapshot, restoreSnapshot } = require("./helpers/snapshot")

const {
  BN,
  constants,
  expectEvent,
  expectRevert,
  time,
} = require("@openzeppelin/test-helpers")

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect
const assert = chai.assert

const BondedECDSAKeepVendor = contract.fromArtifact("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1 = contract.fromArtifact(
  "BondedECDSAKeepVendorImplV1"
)

describe("BondedECDSAKeepVendor", function () {
  const address0 = constants.ZERO_ADDRESS
  const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d"
  const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27"

  const registryAddress = "0x0000000000000000000000000000000000000001"
  const factoryAddress = "0x0000000000000000000000000000000000000002"

  const proxyAdmin = accounts[1]
  const implOwner = accounts[2]
  const newAdmin = accounts[3]

  let currentAddress
  let keepVendor
  let initializeCallData

  before(async () => {
    const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new({
      from: implOwner,
    })

    initializeCallData = bondedECDSAKeepVendorImplV1.contract.methods
      .initialize(registryAddress, factoryAddress)
      .encodeABI()

    keepVendor = await BondedECDSAKeepVendor.new(
      bondedECDSAKeepVendorImplV1.address,
      initializeCallData,
      { from: proxyAdmin }
    )

    currentAddress = bondedECDSAKeepVendorImplV1.address
  })

  describe("constructor", async () => {
    it("reverts when initialization fails", async () => {
      const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new(
        { from: implOwner }
      )

      const initializeCallData = bondedECDSAKeepVendorImplV1.contract.methods
        .initialize(registryAddress, address0)
        .encodeABI()

      await expectRevert(
        BondedECDSAKeepVendor.new(
          bondedECDSAKeepVendorImplV1.address,
          initializeCallData,
          { from: proxyAdmin }
        ),
        "Incorrect factory address"
      )
    })
  })

  describe("upgradeTo", async () => {
    beforeEach(async () => {
      await createSnapshot()
    })

    afterEach(async () => {
      await restoreSnapshot()
    })

    it("sets timestamp", async () => {
      await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })

      const expectedTimestamp = await time.latest()

      expect(await keepVendor.upgradeInitiatedTimestamp()).to.eq.BN(
        expectedTimestamp
      )
    })

    it("sets new implementation", async () => {
      await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })

      assert.equal(await keepVendor.newImplementation.call(), address1)
      assert.equal(await keepVendor.implementation.call(), currentAddress)
    })

    it("sets initialization call data", async () => {
      await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })

      assert.equal(
        await keepVendor.initializationData.call(address1),
        initializeCallData
      )
    })

    it("supports empty initialization call data", async () => {
      await keepVendor.upgradeTo(address1, [], {
        from: proxyAdmin,
      })

      assert.notExists(await keepVendor.initializationData.call(address1))
    })

    it("emits an event", async () => {
      const receipt = await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })

      const expectedTimestamp = await time.latest()

      expectEvent(receipt, "UpgradeStarted", {
        implementation: address1,
        timestamp: expectedTimestamp,
      })
    })

    it("allows implementation overwrite", async () => {
      await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })

      await keepVendor.upgradeTo(address2, initializeCallData, {
        from: proxyAdmin,
      })

      assert.equal(await keepVendor.newImplementation.call(), address2)
    })

    it("allows initialization data overwrite", async () => {
      const initializeCallData2 = "0x123456"

      await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })

      await keepVendor.upgradeTo(address1, initializeCallData2, {
        from: proxyAdmin,
      })

      assert.equal(
        await keepVendor.initializationData.call(address1),
        initializeCallData2
      )
    })

    it("reverts on zero address", async () => {
      await expectRevert(
        keepVendor.upgradeTo(address0, initializeCallData, {
          from: proxyAdmin,
        }),
        "Implementation address can't be zero."
      )
      0
    })

    it("reverts on the same address", async () => {
      await expectRevert(
        keepVendor.upgradeTo(currentAddress, initializeCallData, {
          from: proxyAdmin,
        }),
        "Implementation address must be different from the current one."
      )
    })

    it("reverts when called by non-admin", async () => {
      await expectRevert(
        keepVendor.upgradeTo(address1, initializeCallData),
        "Caller is not the admin"
      )
    })
  })

  describe("completeUpgrade", async () => {
    beforeEach(async () => {
      await createSnapshot()
    })

    afterEach(async () => {
      await restoreSnapshot()
    })

    it("reverts when upgrade not initiated", async () => {
      await expectRevert(
        keepVendor.completeUpgrade({ from: proxyAdmin }),
        "Upgrade not initiated"
      )
    })

    it("reverts when timer not elapsed", async () => {
      await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })

      await time.increase((await keepVendor.upgradeTimeDelay()).subn(2))

      await expectRevert(
        keepVendor.completeUpgrade({ from: proxyAdmin }),
        "Timer not elapsed"
      )
    })

    it("clears timestamp", async () => {
      await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })
      await time.increase(await keepVendor.upgradeTimeDelay())

      await keepVendor.completeUpgrade({ from: proxyAdmin })

      expect(await keepVendor.upgradeInitiatedTimestamp()).to.eq.BN(0)
    })

    it("sets implementation", async () => {
      await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })
      await time.increase(await keepVendor.upgradeTimeDelay())

      await keepVendor.completeUpgrade({ from: proxyAdmin })
      assert.equal(
        await keepVendor.implementation.call({ from: proxyAdmin }),
        address1
      )
    })

    it("emits an event", async () => {
      await keepVendor.upgradeTo(address1, initializeCallData, {
        from: proxyAdmin,
      })
      await time.increase(await keepVendor.upgradeTimeDelay())

      const receipt = await keepVendor.completeUpgrade({ from: proxyAdmin })

      expectEvent(receipt, "UpgradeCompleted", {
        implementation: address1,
      })
    })

    it("supports empty initialization call data", async () => {
      await keepVendor.upgradeTo(address1, [], {
        from: proxyAdmin,
      })
      await time.increase(await keepVendor.upgradeTimeDelay())

      await keepVendor.completeUpgrade({ from: proxyAdmin })
    })

    it("reverts when called by non-admin", async () => {
      await expectRevert(
        keepVendor.completeUpgrade(),
        "Caller is not the admin"
      )
    })

    it("reverts when initialization call fails", async () => {
      const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new()

      const failingData = bondedECDSAKeepVendorImplV1.contract.methods
        .initialize(registryAddress, address0)
        .encodeABI()

      keepVendor.upgradeTo(bondedECDSAKeepVendorImplV1.address, failingData, {
        from: proxyAdmin,
      })
      await time.increase(await keepVendor.upgradeTimeDelay())

      await expectRevert(
        keepVendor.completeUpgrade({ from: proxyAdmin }),
        "Incorrect factory address"
      )
    })
  })

  describe("updateAdmin", async () => {
    beforeEach(async () => {
      await createSnapshot()
    })

    afterEach(async () => {
      await restoreSnapshot()
    })

    it("sets new admin when called by admin", async () => {
      await keepVendor.updateAdmin(newAdmin, { from: proxyAdmin })

      assert.equal(await keepVendor.admin(), newAdmin, "Unexpected admin")
    })

    it("reverts when called by non-admin", async () => {
      await expectRevert(
        keepVendor.updateAdmin(newAdmin),
        "Caller is not the admin"
      )
    })

    it("reverts when called by admin after role transfer", async () => {
      await keepVendor.updateAdmin(newAdmin, { from: proxyAdmin })

      await expectRevert(
        keepVendor.updateAdmin(accounts[0], { from: proxyAdmin }),
        "Caller is not the admin"
      )
    })
  })
})
