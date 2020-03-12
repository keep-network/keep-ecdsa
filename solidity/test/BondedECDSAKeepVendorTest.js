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

  let currentAddress;
  let keepVendor;

  before(async () => {
    const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new();
    keepVendor = await BondedECDSAKeepVendor.new(
      bondedECDSAKeepVendorImplV1.address
    );

    currentAddress = bondedECDSAKeepVendorImplV1.address;
  });

  describe("upgradeTo", async () => {
    beforeEach(async () => {
      await createSnapshot();
    });

    afterEach(async () => {
      await restoreSnapshot();
    });

    it("sets timestamp", async () => {
      await keepVendor.upgradeTo(address1);

      const expectedTimestamp = await time.latest();

      expect(await keepVendor.upgradeInitiatedTimestamp()).to.eq.BN(
        expectedTimestamp
      );
    });

    it("sets new implementation", async () => {
      await keepVendor.upgradeTo(address1);

      assert.equal(await keepVendor.newImplementation.call(), address1);
      assert.equal(await keepVendor.implementation.call(), currentAddress);
    });

    it("emits an event", async () => {
      const receipt = await keepVendor.upgradeTo(address1);

      const expectedTimestamp = await time.latest();

      expectEvent(receipt, "UpgradeStarted", {
        implementation: address1,
        timestamp: expectedTimestamp
      });
    });

    it("allows new value overwrite", async () => {
      await keepVendor.upgradeTo(address1);

      await keepVendor.upgradeTo(address2);

      assert.equal(await keepVendor.newImplementation.call(), address2);
    });

    it("reverts on zero address", async () => {
      await expectRevert(
        keepVendor.upgradeTo(address0),
        "Implementation address can't be zero."
      );
      0;
    });

    it("reverts on the same address", async () => {
      await expectRevert(
        keepVendor.upgradeTo(currentAddress),
        "Implementation address must be different from the current one."
      );
    });

    it("reverts when called by non-owner", async () => {
      await expectRevert(
        keepVendor.upgradeTo(address1, { from: accounts[1] }),
        "Caller is not the owner"
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
      await keepVendor.upgradeTo(address1);

      await time.increase((await keepVendor.upgradeTimeDelay()).subn(1));

      await expectRevert(keepVendor.completeUpgrade(), "Timer not elapsed");
    });

    it("clears timestamp", async () => {
      await keepVendor.upgradeTo(address1);
      await time.increase(await keepVendor.upgradeTimeDelay());

      await keepVendor.completeUpgrade();

      expect(await keepVendor.upgradeInitiatedTimestamp()).to.eq.BN(0);
    });

    it("sets implementation", async () => {
      await keepVendor.upgradeTo(address1);
      await time.increase(await keepVendor.upgradeTimeDelay());

      await keepVendor.completeUpgrade();

      assert.equal(await keepVendor.implementation(), address1);
    });

    it("emits an event", async () => {
      await keepVendor.upgradeTo(address1);
      await time.increase(await keepVendor.upgradeTimeDelay());

      const receipt = await keepVendor.completeUpgrade();

      expectEvent(receipt, "Upgraded", {
        implementation: address1
      });
    });

    it("completes when called by non-owner", async () => {
      await keepVendor.upgradeTo(address1);
      await time.increase(await keepVendor.upgradeTimeDelay());

      await keepVendor.completeUpgrade({ from: accounts[1] });

      assert.equal(await keepVendor.implementation(), address1);
    });
  });
});
