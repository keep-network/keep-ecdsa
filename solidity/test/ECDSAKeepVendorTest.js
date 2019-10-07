var ECDSAKeepVendor = artifacts.require('ECDSAKeepVendorStub')
var ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub')

contract("ECDSAKeepVendor", async accounts => {
    const address0 = "0x0000000000000000000000000000000000000000"
    const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d"
    const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27"

    let keepVendor

    describe("registerFactory", async () => {
        beforeEach(async () => {
            keepVendor = await ECDSAKeepVendor.new()
        })

        it("registers one factory address", async () => {
            let expectedResult = [address1]

            await keepVendor.registerFactory(address1)

            assertFactories(expectedResult)
        })

        it("registers factory with zero address", async () => {
            let expectedResult = [address0]

            await keepVendor.registerFactory(address0)

            assertFactories(expectedResult)
        })

        it("registers two factory addresses", async () => {
            let expectedResult = [address1, address2]

            await keepVendor.registerFactory(address1)
            await keepVendor.registerFactory(address2)

            assertFactories(expectedResult)
        })

        it("fails if address already exists", async () => {
            let expectedResult = [address1]

            await keepVendor.registerFactory(address1)

            try {
                await keepVendor.registerFactory(address1)
                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, 'Factory address already registered')
            }

            assertFactories(expectedResult)
        })

        it("cannot be called by non-owner", async () => {
            let expectedResult = []

            try {
                await keepVendor.registerFactory(address1, { from: accounts[1] })
                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, 'Ownable: caller is not the owner')
            }

            assertFactories(expectedResult)
        })

        async function assertFactories(expectedFactories) {
            let result = await keepVendor.getFactories.call()
            assert.deepEqual(result, expectedFactories, "unexpected registered factories list")
        }
    })

    describe("selectFactory", async () => {
        before(async () => {
            keepVendor = await ECDSAKeepVendor.new()
        })

        it("returns last factory from the list", async () => {
            await keepVendor.registerFactory(address1)
            await keepVendor.registerFactory(address2)

            let expectedResult = address2

            let result = await keepVendor.selectFactoryPublic()

            assert.equal(result, expectedResult, "unexpected factory selected")
        })
    })

    describe("openKeep", async () => {
        before(async () => {
            keepVendor = await ECDSAKeepVendor.new()
        })

        it("reverts if no factories registered", async () => {
            try {
                await keepVendor.openKeep(
                    10, // _groupSize
                    5, // _honestThreshold
                    "0xbc4862697a1099074168d54A555c4A60169c18BD", // _owner
                )

                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, 'No factories registered')
            }
        })

        it("calls selected factory", async () => {
            let factoryStub = await ECDSAKeepFactoryStub.new()
            await keepVendor.registerFactory(factoryStub.address)

            let selectedFactory = await ECDSAKeepFactoryStub.at(
                await keepVendor.selectFactoryPublic.call()
            )

            let expectedResult = await selectedFactory.calculateKeepAddress.call()

            const result = await keepVendor.openKeep.call(
                10, // _groupSize
                5, // _honestThreshold
                "0xbc4862697a1099074168d54A555c4A60169c18BD", // _owner
            )

            assert.equal(result, expectedResult, "unexpected opened keep address")
        })

        it.skip("transfers value to factory", async () => {
            // TODO: Write test
        })
    })
})
