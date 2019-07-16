var ECDSAKeepVendor = artifacts.require('ECDSAKeepVendor')
var ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub')

contract("ECDSAKeepVendor", async accounts => {
    const address0 = "0x0000000000000000000000000000000000000000"
    const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d"
    const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27"

    let keepVendor

    describe("register, remove and get factories", async () => {
        let expectedResult

        beforeEach(async () => {
            keepVendor = await ECDSAKeepVendor.new()
        })

        afterEach(async () => {
            let result = await keepVendor.getFactories.call()
            assert.deepEqual(result, expectedResult, "unexpected registered factories list")
        })

        describe("registerFactory", async () => {
            it("registers one factory address", async () => {
                expectedResult = [address1]

                await keepVendor.registerFactory(address1)
            })

            it("registers two factory addresses", async () => {
                expectedResult = [address1, address2]

                await keepVendor.registerFactory(address1)
                await keepVendor.registerFactory(address2)
            })

            it("fails with zero address", async () => {
                expectedResult = []

                try {
                    await keepVendor.registerFactory(address0)
                    assert(false, 'Test call did not error as expected')
                } catch (e) {
                    assert.include(e.message, 'Factory address cannot be zero')
                }
            })

            it("fails if address already exists", async () => {
                expectedResult = [address1]

                await keepVendor.registerFactory(address1)

                try {
                    await keepVendor.registerFactory(address1)
                    assert(false, 'Test call did not error as expected')
                } catch (e) {
                    assert.include(e.message, 'Factory address already registered')
                }
            })

            it("cannot be called by non owner", async () => {
                expectedResult = []

                try {
                    await keepVendor.registerFactory.call(address1, { from: accounts[1] })
                    assert(false, 'Test call did not error as expected')
                } catch (e) {
                    assert.include(e.message, 'Ownable: caller is not the owner')
                }
            })
        })

        describe("removeFactory", async () => {
            beforeEach(async () => {
                await keepVendor.registerFactory(address1)
            })

            it("remove factory address", async () => {
                expectedResult = []

                await keepVendor.removeFactory(address1)
            })

            it("remove factory address which was not registered", async () => {
                expectedResult = [address1]

                // If address2 has not been registered we don't expect to 
                // get any error on it's removal.
                await keepVendor.removeFactory(address2)
            })

            it("cannot be called by non owner", async () => {
                expectedResult = [address1]

                try {
                    await keepVendor.removeFactory.call(address1, { from: accounts[1] })
                    assert(false, 'Test call did not error as expected')
                } catch (e) {
                    assert.include(e.message, 'Ownable: caller is not the owner')
                }
            })
        })
    })

    describe("selectFactory", async () => {
        before(async () => {
            keepVendor = await ECDSAKeepVendor.new()

            await keepVendor.registerFactory(address1)
            await keepVendor.registerFactory(address2)
        })

        it("returns first factory from the list", async () => {
            let expectedResult = address1

            let result = await keepVendor.selectFactory.call()

            assert.equal(result, expectedResult, "unexpected factory selected")
        })
    })

    describe("openKeep", async () => {
        before(async () => {
            keepVendor = await ECDSAKeepVendor.new()
            let factoryStub1 = await ECDSAKeepFactoryStub.new()
            let factoryStub2 = await ECDSAKeepFactoryStub.new()

            await keepVendor.registerFactory(factoryStub1.address)
            await keepVendor.registerFactory(factoryStub2.address)
        })

        it("calls selected factory", async () => {
            let selectedFactory = await ECDSAKeepFactoryStub.at(
                await keepVendor.selectFactory.call()
            )

            let expectedResult = await selectedFactory.calculateKeepAddress.call()

            let result = await keepVendor.openKeep.call(
                10, // _groupSize
                5, // _honestThreshold
                "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
            )

            assert.equal(result, expectedResult, "unexpected opened keep address")
        })
    })
})
