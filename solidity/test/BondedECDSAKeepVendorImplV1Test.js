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
const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor");
const BondedECDSAKeepVendorImplV1Stub = artifacts.require(
    "BondedECDSAKeepVendorImplV1Stub"
);

contract("BondedECDSAKeepVendorImplV1", async accounts => {
    const address0 = constants.ZERO_ADDRESS;
    const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d";
    const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27";
    const address3 = "0x65ea55c1f10491038425725dc00dffeab2a1e28a";

    let registry, keepVendor;

    beforeEach(async () => {
        registry = await Registry.new();

        await newVendor();

        await keepVendor.initialize(registry.address, address0);
    });

    describe("registerFactory", async () => {
        it("sets timestamp", async () => {
            await keepVendor.registerFactory(address1);

            const expectedTimestamp = await time.latest();

            const actualTimestamp = await keepVendor.getFactoryRegistrationInitiatedTimestamp();

            expect(actualTimestamp).to.eq.BN(expectedTimestamp);
        });

        it("sets new factory address", async () => {
            await keepVendor.registerFactory(address1);

            assert.equal(
                await keepVendor.getNewKeepFactory(),
                address1,
                "unexpected registered new factory"
            );
        });

        it("allows new value overwrite", async () => {
            await keepVendor.registerFactory(address1);

            await keepVendor.registerFactory(address2);

            assert.equal(await keepVendor.getNewKeepFactory(), address2);
        });

        it("emits event", async () => {
            const receipt = await keepVendor.registerFactory(address1);

            const expectedTimestamp = await time.latest();

            expectEvent(receipt, "FactoryRegistrationStarted", {
                factory: address1,
                timestamp: expectedTimestamp
            });
        });

        it("does not register factory with zero address", async () => {
            await expectRevert(
                keepVendor.registerFactory(address0),
                "Incorrect factory address"
            );
        });

        it("does not register factory that is already registered", async () => {
            await newVendor();

            await keepVendor.initialize(registry.address, address1);

            await expectRevert(
                keepVendor.registerFactory(address1),
                "Factory already registered"
            );
        });

        it("does not register factory not approved in registry", async () => {
            await expectRevert(
                keepVendor.registerFactory(address3),
                "Factory contract is not approved"
            );
        });

        it("does not update current factory address", async () => {
            await keepVendor.registerFactory(address1);

            assert.equal(await keepVendor.getKeepFactory(), address0);
        });

        it("cannot be called by non authorized upgrader", async () => {
            await expectRevert(
                keepVendor.registerFactory(address1, { from: accounts[1] }),
                "Caller is not operator contract upgrader"
            );
        });
    });

    describe("completeFactoryRegistration", async () => {
        it("reverts when upgrade not initiated", async () => {
            await expectRevert(
                keepVendor.completeFactoryRegistration(),
                "Upgrade not initiated"
            );
        });

        it("reverts when timer not elapsed", async () => {
            await keepVendor.registerFactory(address1);

            await time.increase(
                (await keepVendor.factoryRegistrationTimeDelay()).subn(1)
            );

            await expectRevert(
                keepVendor.completeFactoryRegistration(),
                "Timer not elapsed"
            );
        });

        it("clears timestamp", async () => {
            await keepVendor.registerFactory(address1);
            await time.increase(await keepVendor.factoryRegistrationTimeDelay());

            await keepVendor.completeFactoryRegistration();

            expect(
                await keepVendor.getFactoryRegistrationInitiatedTimestamp()
            ).to.eq.BN(0);
        });

        it("sets factory address", async () => {
            await keepVendor.registerFactory(address1);
            await time.increase(await keepVendor.factoryRegistrationTimeDelay());

            await keepVendor.completeFactoryRegistration();

            assert.equal(await keepVendor.getKeepFactory(), address1);
        });

        it("emits an event", async () => {
            await keepVendor.registerFactory(address1);
            await time.increase(await keepVendor.factoryRegistrationTimeDelay());

            const receipt = await keepVendor.completeFactoryRegistration();

            expectEvent(receipt, "FactoryRegistered", {
                factory: address1
            });
        });

        it("completes when called by non-owner", async () => {
            await keepVendor.registerFactory(address1);
            await time.increase(await keepVendor.factoryRegistrationTimeDelay());

            await keepVendor.completeFactoryRegistration({ from: accounts[1] });

            assert.equal(await keepVendor.getKeepFactory(), address1);
        });
    });

    async function newVendor() {
        const bondedECDSAKeepVendorImplV1Stub = await BondedECDSAKeepVendorImplV1Stub.new();
        const bondedECDSAKeepVendorProxy = await BondedECDSAKeepVendor.new(
            bondedECDSAKeepVendorImplV1Stub.address
        );
        keepVendor = await BondedECDSAKeepVendorImplV1Stub.at(
            bondedECDSAKeepVendorProxy.address
        );

        await registry.setOperatorContractUpgrader(keepVendor.address, accounts[0]);

        await registry.approveOperatorContract(address0);
        await registry.approveOperatorContract(address1);
        await registry.approveOperatorContract(address2);
    }
});
