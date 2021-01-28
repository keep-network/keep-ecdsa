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

const KeepRegistry = contract.fromArtifact("KeepRegistry")

const BondedECDSAKeepVendor = contract.fromArtifact("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1Stub = contract.fromArtifact(
  "BondedECDSAKeepVendorImplV1Stub"
)

// These tests are calling BondedECDSAKeepVendorImplV1 via proxy contract.
describe("BondedECDSAKeepVendorImplV1viaProxy", function () {
  const address0 = constants.ZERO_ADDRESS
  const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d"
  const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27"
  const address3 = "0x65EA55c1f10491038425725dC00dFFEAb2A1e28A"
  const address4 = "0x4f76Eb125610290301a6F70F535Af9838F48DbC1"

  const proxyAdmin = accounts[1]
  const implOwner = accounts[2]
  const upgrader = accounts[3]

  let registry
  let keepVendor

  before(async () => {
    registry = await KeepRegistry.new()
    await registry.approveOperatorContract(address0)
    await registry.approveOperatorContract(address1)
    await registry.approveOperatorContract(address2)
    await registry.approveOperatorContract(address3)

    keepVendor = await newVendor()
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("initialize", async () => {
    it("marks implementation contract as initialized", async () => {
      const keepVendor = await deployVendorProxy(registry.address, address1)

      assert.isTrue(await keepVendor.initialized())
    })

    it("does not register registry with zero address", async () => {
      await expectRevert(
        deployVendorProxy(address0, address1),
        "Incorrect registry address"
      )
    })

    it("does not register factory with zero address", async () => {
      await expectRevert(
        deployVendorProxy(address1, address0),
        "Incorrect factory address"
      )
    })
  })

  describe("upgradeFactory", async () => {
    it("sets timestamp", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })

      const expectedTimestamp = await time.latest()

      const actualTimestamp = await keepVendor.getFactoryRegistrationInitiatedTimestamp()

      expect(actualTimestamp).to.eq.BN(expectedTimestamp)
    })

    it("sets new factory address", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })

      assert.equal(
        await keepVendor.getNewKeepFactory(),
        address2,
        "unexpected registered new factory"
      )
    })

    it("allows new value overwrite", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })

      await keepVendor.upgradeFactory(address3, { from: upgrader })

      assert.equal(await keepVendor.getNewKeepFactory(), address3)
    })

    it("allows change back to current factory", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })

      await keepVendor.upgradeFactory(address1, { from: upgrader })

      assert.equal(await keepVendor.getNewKeepFactory(), address1)
    })

    it("emits event", async () => {
      const receipt = await keepVendor.upgradeFactory(address2, {
        from: upgrader,
      })

      const expectedTimestamp = await time.latest()

      expectEvent(receipt, "FactoryUpgradeStarted", {
        factory: address2,
        timestamp: expectedTimestamp,
      })
    })

    it("does not register factory with zero address", async () => {
      await expectRevert(
        keepVendor.upgradeFactory(address0, { from: upgrader }),
        "Incorrect factory address"
      )
    })

    it("does not register factory that is already registered", async () => {
      const keepVendor = await newVendor(registry.address, address2)

      await expectRevert(
        keepVendor.upgradeFactory(address2, { from: upgrader }),
        "Factory already registered"
      )
    })

    it("does not register factory not approved in registry", async () => {
      await expectRevert(
        keepVendor.upgradeFactory(address4, { from: upgrader }),
        "Factory contract is not approved"
      )
    })

    it("does not update current factory address", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })

      assert.equal(await keepVendor.getKeepFactory(), address1)
    })

    it("cannot be called by proxy admin", async () => {
      await expectRevert(
        keepVendor.completeFactoryUpgrade({ from: proxyAdmin }),
        "Upgrade not initiated"
      )
    })

    it("cannot be called by implementation owner", async () => {
      await expectRevert(
        keepVendor.completeFactoryUpgrade({ from: implOwner }),
        "Upgrade not initiated"
      )
    })

    it("cannot be called by non-authorized upgrader", async () => {
      await expectRevert(
        keepVendor.upgradeFactory(address2),
        "Caller is not operator contract upgrader"
      )
    })
  })

  describe("completeFactoryUpgrade", async () => {
    it("reverts when upgrade not initiated", async () => {
      await expectRevert(
        keepVendor.completeFactoryUpgrade(),
        "Upgrade not initiated"
      )
    })

    it("reverts when timer not elapsed", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })

      await time.increase((await keepVendor.factoryUpgradeTimeDelay()).subn(2))

      await expectRevert(
        keepVendor.completeFactoryUpgrade(),
        "Timer not elapsed"
      )
    })

    it("clears new factory", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })
      await time.increase(await keepVendor.factoryUpgradeTimeDelay())

      await keepVendor.completeFactoryUpgrade()

      assert.equal(await keepVendor.getNewKeepFactory(), address0)
    })

    it("clears timestamp", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })
      await time.increase(await keepVendor.factoryUpgradeTimeDelay())

      await keepVendor.completeFactoryUpgrade()

      expect(
        await keepVendor.getFactoryRegistrationInitiatedTimestamp()
      ).to.eq.BN(0)
    })

    it("sets factory address", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })
      await time.increase(await keepVendor.factoryUpgradeTimeDelay())

      await keepVendor.completeFactoryUpgrade()

      assert.equal(await keepVendor.getKeepFactory(), address2)
    })

    it("emits an event", async () => {
      await keepVendor.upgradeFactory(address2, { from: upgrader })
      await time.increase(await keepVendor.factoryUpgradeTimeDelay())

      const receipt = await keepVendor.completeFactoryUpgrade()

      expectEvent(receipt, "FactoryUpgradeCompleted", {
        factory: address2,
      })
    })
  })

  async function deployVendorProxy(registryAddress, factoryAddress) {
    const bondedECDSAKeepVendorImplV1Stub = await BondedECDSAKeepVendorImplV1Stub.new(
      { from: implOwner }
    )

    const initializeCallData = bondedECDSAKeepVendorImplV1Stub.contract.methods
      .initialize(registryAddress, factoryAddress)
      .encodeABI()

    const bondedECDSAKeepVendorProxy = await BondedECDSAKeepVendor.new(
      bondedECDSAKeepVendorImplV1Stub.address,
      initializeCallData,
      { from: proxyAdmin }
    )
    const keepVendor = await BondedECDSAKeepVendorImplV1Stub.at(
      bondedECDSAKeepVendorProxy.address
    )

    return keepVendor
  }

  async function newVendor(
    registryAddress = registry.address,
    factoryAddress = address1
  ) {
    const keepVendor = await deployVendorProxy(registryAddress, factoryAddress)

    await registry.setOperatorContractUpgrader(keepVendor.address, upgrader)

    return keepVendor
  }
})
