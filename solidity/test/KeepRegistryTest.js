var KeepRegistry = artifacts.require('KeepRegistry')

contract("KeepRegistry", async accounts => {
    const keepType1 = "ECDSA"
    const keepType2 = "BondedECDSA"
    const address0 = "0x0000000000000000000000000000000000000000"
    const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d"
    const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27"

    let keepRegistry

    describe("setVendor", async () => {
        beforeEach(async () => {
            keepRegistry = await KeepRegistry.new()
        })

        it("sets vendor address for new keep type", async () => {
            await keepRegistry.setVendor(keepType1, address1)

            let result = await keepRegistry.getVendor.call(keepType1)
            assert.deepEqual(result, address1, "unexpected keep vendor address")
        })

        it("replaces vendor address for keep type", async () => {
            await keepRegistry.setVendor(keepType1, address1)
            await keepRegistry.setVendor(keepType1, address2)

            let result = await keepRegistry.getVendor.call(keepType1)
            assert.deepEqual(result, address2, "unexpected keep vendor address")
        })

        it("sets two keep types with different addresses", async () => {
            await keepRegistry.setVendor(keepType1, address1)
            await keepRegistry.setVendor(keepType2, address2)

            let result1 = await keepRegistry.getVendor.call(keepType1)
            assert.deepEqual(result1, address1, "unexpected keep vendor address")

            let result2 = await keepRegistry.getVendor.call(keepType2)
            assert.deepEqual(result2, address2, "unexpected keep vendor address")
        })

        it("cannot be called by non-owner", async () => {
            try {
                await keepRegistry.setVendor(keepType1, address1, { from: accounts[1] })
                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, 'Ownable: caller is not the owner')
            }
        })
    })

    describe("getVendor", async () => {
        before(async () => {
            keepRegistry = await KeepRegistry.deployed()
        })

        it("returns zero for not registered keep type", async () => {
            let result = await keepRegistry.getVendor.call("NOT EXISTING")
            assert.deepEqual(result, address0, "unexpected keep vendor address")
        })
    })
})
