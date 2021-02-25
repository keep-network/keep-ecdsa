const { accounts, contract } = require("@openzeppelin/test-environment")
const { createSnapshot, restoreSnapshot } = require("./helpers/snapshot")

const { constants, expectRevert } = require("@openzeppelin/test-helpers")

const KeepRegistry = contract.fromArtifact("KeepRegistry")
const BondedECDSAKeepVendorImplV1Stub = contract.fromArtifact(
  "BondedECDSAKeepVendorImplV1Stub"
)

const chai = require("chai")
const assert = chai.assert

// These tests are calling BondedECDSAKeepVendorImplV1 directly.
describe("BondedECDSAKeepVendorImplV1", function () {
  const address0 = constants.ZERO_ADDRESS
  const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d"

  let registry
  let keepVendor

  const implOwner = accounts[1]
  const upgrader = accounts[2]

  before(async () => {
    registry = await KeepRegistry.new()

    await registry.approveOperatorContract(address0)
    await registry.approveOperatorContract(address1)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("initialize", async () => {
    it("marks contract as initialized", async () => {
      keepVendor = await BondedECDSAKeepVendorImplV1Stub.new()

      assert.isTrue(await keepVendor.initialized())
    })
  })

  describe("initialize", async () => {
    before(async () => {
      keepVendor = await BondedECDSAKeepVendorImplV1Stub.new({
        from: implOwner,
      })
    })

    it("reverts as contract is already initialized", async () => {
      await expectRevert(
        keepVendor.initialize(address1, address1),
        "Contract is already initialized."
      )
    })
  })

  describe("upgradeFactory", async () => {
    before(async () => {
      keepVendor = await newVendor()
    })

    it("reverts when called directly", async () => {
      await expectRevert(
        keepVendor.upgradeFactory(address1, { from: upgrader }),
        "KeepRegistry address is not registered"
      )
    })
  })

  describe("completeFactoryUpgrade", async () => {
    before(async () => {
      keepVendor = await newVendor()
    })

    it("reverts when called directly", async () => {
      await expectRevert(
        keepVendor.completeFactoryUpgrade(),
        "Upgrade not initiated"
      )
    })
  })

  async function newVendor() {
    const keepVendor = await BondedECDSAKeepVendorImplV1Stub.new()

    await registry.setOperatorContractUpgrader(keepVendor.address, upgrader)

    return keepVendor
  }
})
