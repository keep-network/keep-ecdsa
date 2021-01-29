const { accounts, contract } = require("@openzeppelin/test-environment")
const { createSnapshot, restoreSnapshot } = require("./helpers/snapshot")

const {
  expectEvent,
  expectRevert,
  time,
} = require("@openzeppelin/test-helpers")

const BondedECDSAKeepVendor = contract.fromArtifact("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1Stub = contract.fromArtifact(
  "BondedECDSAKeepVendorImplV1Stub"
)
const BondedECDSAKeepVendorImplV2Stub = contract.fromArtifact(
  "BondedECDSAKeepVendorImplV2Stub"
)

const chai = require("chai")
const assert = chai.assert

describe("BondedECDSAKeepVendorUpgrade", function () {
  const registryAddress = "0x0000000000000000000000000000000000000001"
  const factoryAddress = "0x0000000000000000000000000000000000000002"

  const proxyAdmin = accounts[1]

  let keepVendorProxy

  let implV1
  let implV2

  before(async () => {
    implV1 = await BondedECDSAKeepVendorImplV1Stub.new()
    implV2 = await BondedECDSAKeepVendorImplV2Stub.new()

    const initializeCallData = implV1.contract.methods
      .initialize(registryAddress, factoryAddress)
      .encodeABI()

    keepVendorProxy = await BondedECDSAKeepVendor.new(
      implV1.address,
      initializeCallData,
      { from: proxyAdmin }
    )
  })

  describe("upgrade process", async () => {
    beforeEach(async () => {
      await createSnapshot()
    })

    afterEach(async () => {
      await restoreSnapshot()
    })

    it("upgrades to new version", async () => {
      const initializeCallData = implV2.contract.methods
        .initialize(false)
        .encodeABI()

      await keepVendorProxy.upgradeTo(implV2.address, initializeCallData, {
        from: proxyAdmin,
      })
      await time.increase(await keepVendorProxy.upgradeTimeDelay())

      const receipt = await keepVendorProxy.completeUpgrade({
        from: proxyAdmin,
      })
      expectEvent(receipt, "UpgradeCompleted", {
        implementation: implV2.address,
      })

      assert.equal(await keepVendorProxy.implementation(), implV2.address)

      const keepVendor = await BondedECDSAKeepVendorImplV2Stub.at(
        keepVendorProxy.address
      )

      assert.equal(await keepVendor.version(), "V2")

      assert.isTrue(
        await keepVendor.initialized(),
        "implementation not initialized"
      )
    })

    it("reverts when call delegated to wrong contract", async () => {
      const initializeCallData = implV1.contract.methods
        .initialize(registryAddress, factoryAddress)
        .encodeABI()

      keepVendorProxy.upgradeTo(implV2.address, initializeCallData, {
        from: proxyAdmin,
      })
      await time.increase(await keepVendorProxy.upgradeTimeDelay())

      await expectRevert(
        keepVendorProxy.completeUpgrade({ from: proxyAdmin }),
        "revert"
      )
    })

    it("reverts when upgraded to a previously initialized contract", async () => {
      const initializeCallData1 = implV1.contract.methods
        .initialize(registryAddress, factoryAddress)
        .encodeABI()

      const initializeCallData2 = implV2.contract.methods
        .initialize(false)
        .encodeABI()

      keepVendorProxy.upgradeTo(implV2.address, initializeCallData2, {
        from: proxyAdmin,
      })
      await time.increase(await keepVendorProxy.upgradeTimeDelay())

      await keepVendorProxy.completeUpgrade({ from: proxyAdmin })

      await keepVendorProxy.upgradeTo(implV1.address, initializeCallData1, {
        from: proxyAdmin,
      })
      await time.increase(await keepVendorProxy.upgradeTimeDelay())

      await expectRevert(
        keepVendorProxy.completeUpgrade({ from: proxyAdmin }),
        "Contract is already initialized"
      )
    })
  })
})
