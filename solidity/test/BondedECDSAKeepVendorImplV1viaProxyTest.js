import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const {
    constants,
    expectRevert,
    time
} = require("@openzeppelin/test-helpers");

const Registry = artifacts.require("Registry");

const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor");
const BondedECDSAKeepVendorImplV1Stub = artifacts.require(
    "BondedECDSAKeepVendorImplV1Stub"
);

// These tests are calling BondedECDSAKeepVendorImplV1 via proxy contract.
contract("BondedECDSAKeepVendorImplV1viaProxy", async accounts => {
    const address0 = constants.ZERO_ADDRESS;
    const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d";
    const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27";

    const proxyAdmin = accounts[1];
    const implOwner = accounts[2];
    const upgrader = accounts[3];

    let registry, keepVendor;

    before(async () => {
        registry = await Registry.new();
    });

    describe("constructor", async () => {
        beforeEach(async () => {
            await createSnapshot();
        });

        afterEach(async () => {
            await restoreSnapshot();
        });

        it("marks implementation contract as initialized", async () => {
            const keepVendor = await deployVendorProxy()

            assert.isTrue(await keepVendor.initialized());
        });
    });

    describe("registerFactory", async () => {
        before(async () => {
            keepVendor = await newVendor();
        });

        beforeEach(async () => {
            await createSnapshot();
        });

        afterEach(async () => {
            await restoreSnapshot();
        });

        it("succeeds", async () => {
            await keepVendor.registerFactory(address2, { from: upgrader });

            assert.equal(
                await keepVendor.getNewKeepFactory(),
                address2,
                "unexpected registered new factory"
            );
        });

        it("cannot be called by non authorized upgrader", async () => {
            await expectRevert(
                keepVendor.registerFactory(address2),
                "Caller is not operator contract upgrader"
            );
        });
    });

    describe("completeFactoryRegistration", async () => {
        before(async () => {
            keepVendor = await newVendor();
        });

        beforeEach(async () => {
            await createSnapshot();
        });

        afterEach(async () => {
            await restoreSnapshot();
        });

        it("succeeds", async () => {
            await keepVendor.registerFactory(address2, { from: upgrader });
            await time.increase(await keepVendor.factoryRegistrationTimeDelay());

            await keepVendor.completeFactoryRegistration();

            assert.equal(await keepVendor.getKeepFactory(), address2);
        });
    });

    async function deployVendorProxy() {
        const bondedECDSAKeepVendorImplV1Stub = await BondedECDSAKeepVendorImplV1Stub.new(
            { from: implOwner }
        );

        const initializeCallData = bondedECDSAKeepVendorImplV1Stub
            .contract.methods.initialize(registry.address, address1).encodeABI()

        const bondedECDSAKeepVendorProxy = await BondedECDSAKeepVendor.new(
            bondedECDSAKeepVendorImplV1Stub.address,
            initializeCallData,
            { from: proxyAdmin }
        );
        const keepVendor = await BondedECDSAKeepVendorImplV1Stub.at(
            bondedECDSAKeepVendorProxy.address
        );

        return keepVendor
    }

    async function newVendor() {
        const keepVendor = await deployVendorProxy()

        await registry.setOperatorContractUpgrader(keepVendor.address, upgrader);

        await registry.approveOperatorContract(address0);
        await registry.approveOperatorContract(address1);
        await registry.approveOperatorContract(address2);

        return keepVendor;
    }
});
