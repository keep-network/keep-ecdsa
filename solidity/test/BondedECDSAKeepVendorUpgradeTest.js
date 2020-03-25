import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const {
  expectEvent,
  expectRevert,
  time
} = require("@openzeppelin/test-helpers");

const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor");
const BondedECDSAKeepVendorImplV1Stub = artifacts.require(
  "BondedECDSAKeepVendorImplV1Stub"
);
const BondedECDSAKeepVendorImplV2Stub = artifacts.require(
  "BondedECDSAKeepVendorImplV2Stub"
);

contract("BondedECDSAKeepVendorUpgrade", async accounts => {
  let registryAddress = "0x0000000000000000000000000000000000000001";
  let factoryAddress = "0x0000000000000000000000000000000000000002";

  const proxyAdmin = accounts[1];

  let keepVendorProxy;

  let implV1, implV2;
  const V1 = "V1";
  const V2 = "V2";

  before(async () => {
    implV1 = await BondedECDSAKeepVendorImplV1Stub.new();
    implV2 = await BondedECDSAKeepVendorImplV2Stub.new();

    const initializeCallData = implV1.contract.methods
      .initialize(registryAddress, factoryAddress)
      .encodeABI();

    keepVendorProxy = await BondedECDSAKeepVendor.new(
      V1,
      implV1.address,
      initializeCallData,
      { from: proxyAdmin }
    );
  });

  describe("upgrade process", async () => {
    beforeEach(async () => {
      await createSnapshot();
    });

    afterEach(async () => {
      await restoreSnapshot();
    });

    it("upgrades to new version", async () => {
      const initializeCallData = implV2.contract.methods
        .initialize(false)
        .encodeABI();

      await keepVendorProxy.upgradeToAndCall(V2, implV2.address, initializeCallData, {
        from: proxyAdmin
      });
      await time.increase(await keepVendorProxy.upgradeTimeDelay());

      const receipt = await keepVendorProxy.completeUpgrade({ from: proxyAdmin });
      expectEvent(receipt, "UpgradeCompleted", {
        implementation: implV2.address
      });

      assert.equal(
        await keepVendorProxy.implementation(),
        implV2.address
      );

      const keepVendor = await BondedECDSAKeepVendorImplV2Stub.at(
        keepVendorProxy.address
      );

      assert.equal(
        await keepVendor.version(),
        "V2"
      );

      assert.isTrue(await keepVendor.initialized(), "implementation not initialized");
    });

    it("reverts when call delegated to wrong contract", async () => {
      const initializeCallData = implV1.contract.methods
        .initialize(registryAddress, factoryAddress)
        .encodeABI();

      keepVendorProxy.upgradeToAndCall(V2, implV2.address, initializeCallData, {
        from: proxyAdmin
      })
      await time.increase(await keepVendorProxy.upgradeTimeDelay());

      await expectRevert(
        keepVendorProxy.completeUpgrade({ from: proxyAdmin }),
        "revert"
      );
    });
  });
});
