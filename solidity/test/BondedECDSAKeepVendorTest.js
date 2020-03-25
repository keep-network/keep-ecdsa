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

const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendorStub");
const BondedECDSAKeepVendorImplV1 = artifacts.require(
  "BondedECDSAKeepVendorImplV1"
);

contract("BondedECDSAKeepVendor", async accounts => {
  const address0 = constants.ZERO_ADDRESS;
  const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d";
  const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27";
  const address3 = "0xc005694f9b7806af17aa05df09b6e83207cb2fd8";

  let registryAddress = "0x0000000000000000000000000000000000000001";
  let factoryAddress = "0x0000000000000000000000000000000000000002";

  const proxyAdmin = accounts[1];
  const implOwner = accounts[2];

  let currentAddress;
  let keepVendor;
  let initializeCallData;

  const V1 = "V1";
  const V2 = "V2";
  const V3 = "V3";

  before(async () => {
    const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new({
      from: implOwner
    });

    initializeCallData = bondedECDSAKeepVendorImplV1.contract.methods
      .initialize(registryAddress, factoryAddress)
      .encodeABI();

    keepVendor = await BondedECDSAKeepVendor.new(
      V1,
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
          V2,
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
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });

      const expectedTimestamp = await time.latest();

      expect(await keepVendor.upgradeInitiatedTimestamp()).to.eq.BN(
        expectedTimestamp
      );
    });

    it("sets new version", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });

      assert.equal(await keepVendor.version.call(), V1);
      assert.equal(await keepVendor.upgradeVersion.call(), V2);
    });

    it("sets new implementation", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });

      assert.equal(await keepVendor.implementation.call(), currentAddress);
      assert.equal(await keepVendor.upgradeImplementation.call(), address1);
    });

    it("sets initialization call data", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });

      assert.equal(await keepVendor.upgradeInitialization.call(), initializeCallData);
    });

    it("supports empty initialization call data", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, [], {
        from: proxyAdmin
      });

      assert.notExists(await keepVendor.upgradeInitialization.call());
    });

    it("emits an event", async () => {
      const receipt = await keepVendor.upgradeToAndCall(
        V2,
        address1,
        initializeCallData,
        { from: proxyAdmin }
      );

      const expectedTimestamp = await time.latest();

      expectEvent(receipt, "UpgradeStarted", {
        version: V2,
        implementation: address1,
        timestamp: expectedTimestamp
      });
    });

    it("allows upgrade overwrite for the same version", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });

      // With the same implementation address
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });

      assert.equal(await keepVendor.version.call(), V1);
      assert.equal(await keepVendor.upgradeVersion.call(), V2);
      assert.equal(await keepVendor.upgradeImplementation.call(), address1);

      // With different implementation address
      await keepVendor.upgradeToAndCall(V2, address2, initializeCallData, {
        from: proxyAdmin
      });

      assert.equal(await keepVendor.version.call(), V1);
      assert.equal(await keepVendor.upgradeVersion.call(), V2);
      assert.equal(await keepVendor.upgradeImplementation.call(), address2);
    });

    it("allows upgrade overwrite with new version", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });

      await keepVendor.upgradeToAndCall(V3, address2, initializeCallData, {
        from: proxyAdmin
      });

      assert.equal(await keepVendor.version.call(), V1);
      assert.equal(await keepVendor.upgradeVersion.call(), V3);
      assert.equal(await keepVendor.upgradeImplementation.call(), address2);
    });

    it("reverts on empty version", async () => {
      await expectRevert(
        keepVendor.upgradeToAndCall("", address2, initializeCallData, {
          from: proxyAdmin
        }),
        "Version can't be empty string."
      );
    });

    it("reverts on zero address", async () => {
      await expectRevert(
        keepVendor.upgradeToAndCall(V2, address0, initializeCallData, {
          from: proxyAdmin
        }),
        "Implementation address can't be zero."
      );
    });

    it("reverts on the same version", async () => {
      await expectRevert(
        keepVendor.upgradeToAndCall(V1, address2, initializeCallData, {
          from: proxyAdmin
        }),
        "Implementation version must be different from the current one"
      );
    });

    it("reverts on the same version as one of historic", async () => {
      await keepVendor.upgradeToAndCall(V2, address2, initializeCallData, {
        from: proxyAdmin
      });
      await time.increase(await keepVendor.upgradeTimeDelay());
      await keepVendor.completeUpgrade({ from: proxyAdmin });

      await expectRevert(
        keepVendor.upgradeToAndCall(V1, address3, initializeCallData, {
          from: proxyAdmin
        }),
        "Implementation version has already been registered before"
      );
    });

    it("allows the same implementation address as the current", async () => {
      await keepVendor.upgradeToAndCall(V2, currentAddress, initializeCallData, {
        from: proxyAdmin
      })
    });

    it("reverts when called by non-admin", async () => {
      await expectRevert(
        keepVendor.upgradeToAndCall(V2, address1, initializeCallData),
        "Caller is not the admin"
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
      await expectRevert(keepVendor.completeUpgrade({ from: proxyAdmin }), "Upgrade not initiated");
    });

    it("reverts when timer not elapsed", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });

      await time.increase((await keepVendor.upgradeTimeDelay()).subn(2));

      await expectRevert(keepVendor.completeUpgrade({ from: proxyAdmin }), "Timer not elapsed");
    });

    it("clears timestamp", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });
      await time.increase(await keepVendor.upgradeTimeDelay());

      await keepVendor.completeUpgrade({ from: proxyAdmin });

      expect(await keepVendor.upgradeInitiatedTimestamp()).to.eq.BN(0);
    });

    it("sets implementation", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });
      await time.increase(await keepVendor.upgradeTimeDelay());

      await keepVendor.completeUpgrade({ from: proxyAdmin });
      assert.equal(
        await keepVendor.implementation.call({ from: proxyAdmin }),
        address1
      );
    });

    it("emits an event", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, initializeCallData, {
        from: proxyAdmin
      });
      await time.increase(await keepVendor.upgradeTimeDelay());

      const receipt = await keepVendor.completeUpgrade({ from: proxyAdmin });

      expectEvent(receipt, "UpgradeCompleted", {
        version: V2,
        implementation: address1
      });
    });

    it("supports empty initialization call data", async () => {
      await keepVendor.upgradeToAndCall(V2, address1, [], {
        from: proxyAdmin
      });
      await time.increase(await keepVendor.upgradeTimeDelay());

      await keepVendor.completeUpgrade({ from: proxyAdmin });
      assert.equal(
        await keepVendor.implementation.call({ from: proxyAdmin }),
        address1
      );
    });

    it("reverts when called by non-admin", async () => {
      await expectRevert(
        keepVendor.completeUpgrade(),
        "Caller is not the admin"
      );
    });

    it("reverts when initialization call fails", async () => {
      const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new();

      const failingData = bondedECDSAKeepVendorImplV1.contract.methods
        .initialize(registryAddress, address0)
        .encodeABI();


      keepVendor.upgradeToAndCall(V2, bondedECDSAKeepVendorImplV1.address, failingData, {
        from: proxyAdmin
      })
      await time.increase(await keepVendor.upgradeTimeDelay());

      await expectRevert(
        keepVendor.completeUpgrade({ from: proxyAdmin }),
        "Incorrect factory address"
      );
    });
  });
});
