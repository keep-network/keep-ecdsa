import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const {
  BN,
  constants,
  expectEvent,
  expectRevert,
  time
} = require("@openzeppelin/test-helpers");

const chai = require("chai");
chai.use(require("bn-chai")(BN));
const expect = chai.expect;

const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor");
const BondedECDSAKeepVendorImplV1 = artifacts.require(
  "BondedECDSAKeepVendorImplV1"
);

contract("BondedECDSAKeepVendor", async accounts => {
  const address0 = constants.ZERO_ADDRESS;
  const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d";
  const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27";

  let registryAddress = "0x0000000000000000000000000000000000000001";
  let factoryAddress = "0x0000000000000000000000000000000000000002";

  const proxyAdmin = accounts[1];
  const implOwner = accounts[2];

  let currentAddress;
  let keepVendor;
  let initializeCallData;

  before(async () => {
    const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new({
      from: implOwner
    });

    initializeCallData = bondedECDSAKeepVendorImplV1.contract.methods
      .initialize(registryAddress, factoryAddress)
      .encodeABI();

    keepVendor = await BondedECDSAKeepVendor.new(
      bondedECDSAKeepVendorImplV1.address,
      initializeCallData,
      { from: proxyAdmin }
    );

    currentAddress = bondedECDSAKeepVendorImplV1.address;
  });

  describe("constructor", async () => {
    it("reverts when initialization fails", async () => {
      const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new(
        { from: implOwner }
      );

      const initializeCallData = bondedECDSAKeepVendorImplV1.contract.methods
        .initialize(registryAddress, address0)
        .encodeABI();

      await expectRevert(
        BondedECDSAKeepVendor.new(
          bondedECDSAKeepVendorImplV1.address,
          initializeCallData,
          { from: proxyAdmin }
        ),
        "Incorrect factory address"
      );
    });
  });

  describe("upgradeToAndCall", async () => {
    beforeEach(async () => {
      await createSnapshot();
    });

    afterEach(async () => {
      await restoreSnapshot();
    });

    it("sets timestamp", async () => {
      await keepVendor.upgradeToAndCall(address1, initializeCallData, {
        from: proxyAdmin
      });

      const expectedTimestamp = await time.latest();

      expect(await keepVendor.upgradeInitiatedTimestamp()).to.eq.BN(
        expectedTimestamp
      );
    });

    it("sets new implementation", async () => {
      await keepVendor.upgradeToAndCall(address1, initializeCallData, {
        from: proxyAdmin
      });

      assert.equal(await keepVendor.newImplementation.call(), address1);
      assert.equal(await keepVendor.implementation.call(), currentAddress);
    });

    it("emits an event", async () => {
      const receipt = await keepVendor.upgradeToAndCall(
        address1,
        initializeCallData,
        { from: proxyAdmin }
      );

      const expectedTimestamp = await time.latest();

      expectEvent(receipt, "UpgradeStarted", {
        implementation: address1,
        timestamp: expectedTimestamp
      });
    });

    it("allows new value overwrite", async () => {
      await keepVendor.upgradeToAndCall(address1, initializeCallData, {
        from: proxyAdmin
      });

      await keepVendor.upgradeToAndCall(address2, initializeCallData, {
        from: proxyAdmin
      });

      assert.equal(await keepVendor.newImplementation.call(), address2);
    });

    it("reverts on zero address", async () => {
      await expectRevert(
        keepVendor.upgradeToAndCall(address0, initializeCallData, {
          from: proxyAdmin
        }),
        "Implementation address can't be zero."
      );
      0;
    });

    it("reverts on the same address", async () => {
      await expectRevert(
        keepVendor.upgradeToAndCall(currentAddress, initializeCallData, {
          from: proxyAdmin
        }),
        "Implementation address must be different from the current one."
      );
    });

    it("reverts when called by non-owner", async () => {
      await expectRevert.unspecified(
        keepVendor.upgradeToAndCall(address1, initializeCallData)
      );
    });

    it("reverts when initialization call fails", async () => {
      const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new();

      const failingData = bondedECDSAKeepVendorImplV1.contract.methods
        .initialize(registryAddress, address0)
        .encodeABI();

      await expectRevert(
        keepVendor.upgradeToAndCall(bondedECDSAKeepVendorImplV1.address, failingData, {
          from: proxyAdmin
        }),
        "Incorrect factory address"
      );
    });
  });

  describe("completeUpgrade", async () => {
    beforeEach(async () => {
      await createSnapshot();
    });

    afterEach(async () => {
      await restoreSnapshot();
    });

    it("reverts when upgrade not initiated", async () => {
      await expectRevert(keepVendor.completeUpgrade(), "Upgrade not initiated");
    });

    it("reverts when timer not elapsed", async () => {
      await keepVendor.upgradeToAndCall(address1, initializeCallData, {
        from: proxyAdmin
      });

      await time.increase((await keepVendor.upgradeTimeDelay()).subn(2));

      await expectRevert(keepVendor.completeUpgrade(), "Timer not elapsed");
    });

    it("clears timestamp", async () => {
      await keepVendor.upgradeToAndCall(address1, initializeCallData, {
        from: proxyAdmin
      });
      await time.increase(await keepVendor.upgradeTimeDelay());

      await keepVendor.completeUpgrade();

      expect(await keepVendor.upgradeInitiatedTimestamp()).to.eq.BN(0);
    });

    it("sets implementation", async () => {
      await keepVendor.upgradeToAndCall(address1, initializeCallData, {
        from: proxyAdmin
      });
      await time.increase(await keepVendor.upgradeTimeDelay());

      await keepVendor.completeUpgrade();
      assert.equal(
        await keepVendor.implementation.call({ from: proxyAdmin }),
        address1
      );
    });

    it("emits an event", async () => {
      await keepVendor.upgradeToAndCall(address1, initializeCallData, {
        from: proxyAdmin
      });
      await time.increase(await keepVendor.upgradeTimeDelay());

      const receipt = await keepVendor.completeUpgrade();

      expectEvent(receipt, "Upgraded", {
        implementation: address1
      });
    });
  });
});
