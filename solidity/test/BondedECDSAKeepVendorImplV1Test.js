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

const Registry = artifacts.require("Registry");
const BondedECDSAKeepVendorImplV1Stub = artifacts.require(
    "BondedECDSAKeepVendorImplV1Stub"
);

contract("BondedECDSAKeepVendorImplV1", async accounts => {
    const address0 = constants.ZERO_ADDRESS;
    const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d";
    const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27";
    const address3 = "0x65EA55c1f10491038425725dC00dFFEAb2A1e28A";
    const address4 = "0x4f76Eb125610290301a6F70F535Af9838F48DbC1";

    let registry, keepVendor;

    const implOwner = accounts[1];

    before(async () => {
        registry = await Registry.new();
    });

    describe("initialize", async () => {
        before(async () => {
            keepVendor = await BondedECDSAKeepVendorImplV1Stub.new({ from: implOwner });
        });

        beforeEach(async () => {
            await createSnapshot();
        });

        afterEach(async () => {
            await restoreSnapshot();
        });

        it("does not register registry with zero address", async () => {
            await expectRevert(
                keepVendor.initialize(address0, address1),
                "Incorrect registry address"
            );
        });

        it("does not register factory with zero address", async () => {
            await expectRevert(
                keepVendor.initialize(address1, address0),
                "Incorrect factory address"
            );
        });

        it("marks contract as initialized", async () => {
            await keepVendor.initialize(address1, address1);

            assert.isTrue(await keepVendor.initialized({ from: implOwner }));
        });

        it("can be called only once", async () => {
            await keepVendor.initialize(address1, address1);

            await expectRevert(
                keepVendor.initialize(address1, address1),
                "Contract is already initialized."
            );
        });
    });

    describe("registerFactory", async () => {
        before(async () => {
            keepVendor = await newVendor();

            await keepVendor.initialize(registry.address, address1);
        });

        beforeEach(async () => {
            await createSnapshot();
        });

        afterEach(async () => {
            await restoreSnapshot();
        });

        it("sets timestamp", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });

            const expectedTimestamp = await time.latest();

            const actualTimestamp = await keepVendor.getFactoryRegistrationInitiatedTimestamp();

            expect(actualTimestamp).to.eq.BN(expectedTimestamp);
        });

        it("sets new factory address", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });

            assert.equal(
                await keepVendor.getNewKeepFactory(),
                address2,
                "unexpected registered new factory"
            );
        });

        it("allows new value overwrite", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });

            await keepVendor.registerFactory(address3, { from: implOwner });

            assert.equal(await keepVendor.getNewKeepFactory(), address3);
        });

        it("allows change back to current factory", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });

            await keepVendor.registerFactory(address1, { from: implOwner });

            assert.equal(await keepVendor.getNewKeepFactory(), address1);
        });

        it("emits event", async () => {
            const receipt = await keepVendor.registerFactory(address2, { from: implOwner });

            const expectedTimestamp = await time.latest();

            expectEvent(receipt, "FactoryRegistrationStarted", {
                factory: address2,
                timestamp: expectedTimestamp
            });
        });

        it("does not register factory with zero address", async () => {
            await expectRevert(
                keepVendor.registerFactory(address0, { from: implOwner }),
                "Incorrect factory address"
            );
        });

        it("does not register factory that is already registered", async () => {
            const keepVendor = await newVendor();

            await keepVendor.initialize(registry.address, address2);

            await expectRevert(
                keepVendor.registerFactory(address2, { from: implOwner }),
                "Factory already registered"
            );
        });

        it("does not register factory not approved in registry", async () => {
            await expectRevert(
                keepVendor.registerFactory(address4, { from: implOwner }),
                "Factory contract is not approved"
            );
        });

        it("does not update current factory address", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });

            assert.equal(await keepVendor.getKeepFactory(), address1);
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

            await keepVendor.initialize(registry.address, address1);
        });

        beforeEach(async () => {
            await createSnapshot();
        });

        afterEach(async () => {
            await restoreSnapshot();
        });

        it("reverts when upgrade not initiated", async () => {
            await expectRevert(
                keepVendor.completeFactoryRegistration(),
                "Upgrade not initiated"
            );
        });

        it("reverts when timer not elapsed", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });

            await time.increase(
                (await keepVendor.factoryRegistrationTimeDelay()).subn(2)
            );

            await expectRevert(
                keepVendor.completeFactoryRegistration(),
                "Timer not elapsed"
            );
        });

        it("clears new factory", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });
            await time.increase(await keepVendor.factoryRegistrationTimeDelay());

            await keepVendor.completeFactoryRegistration();

            assert.equal(await keepVendor.getNewKeepFactory(), address0);
        });

        it("clears timestamp", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });
            await time.increase(await keepVendor.factoryRegistrationTimeDelay());

            await keepVendor.completeFactoryRegistration();

            expect(
                await keepVendor.getFactoryRegistrationInitiatedTimestamp()
            ).to.eq.BN(0);
        });

        it("sets factory address", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });
            await time.increase(await keepVendor.factoryRegistrationTimeDelay());

            await keepVendor.completeFactoryRegistration();

            assert.equal(await keepVendor.getKeepFactory(), address2);
        });

        it("emits an event", async () => {
            await keepVendor.registerFactory(address2, { from: implOwner });
            await time.increase(await keepVendor.factoryRegistrationTimeDelay());

            const receipt = await keepVendor.completeFactoryRegistration();

            expectEvent(receipt, "FactoryRegistered", {
                factory: address2
            });
        });
    });

    async function newVendor() {
        const keepVendor = await BondedECDSAKeepVendorImplV1Stub.new();

        await registry.setOperatorContractUpgrader(keepVendor.address, implOwner);

        await registry.approveOperatorContract(address0);
        await registry.approveOperatorContract(address1);
        await registry.approveOperatorContract(address2);
        await registry.approveOperatorContract(address3);

        return keepVendor;
    }
});
