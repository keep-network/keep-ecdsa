var KeepRegistry = artifacts.require('KeepRegistry')

contract("KeepRegistry", async accounts => {
    const keepType1 = "ECDSA"
    const keepType2 = "BondedECDSA"
    const address0 = "0x0000000000000000000000000000000000000000"
    const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d"
    const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27"

    let keepRegistry

    before(async () => {
        keepRegistry = await KeepRegistry.deployed()
    })

    describe("setKeepTypeVendor", async () => {
        it("sets vendor address for new keep type", async () => {
            keepRegistry = await KeepRegistry.new()

            await keepRegistry.setKeepTypeVendor(keepType1, address1)

            let result = await keepRegistry.getKeepTypeVendor.call(keepType1)
            assert.deepEqual(result, address1, "unexpected keep vendor address")
        })

        it("replaces vendor address for keep type", async () => {
            await keepRegistry.setKeepTypeVendor(keepType1, address1)
            await keepRegistry.setKeepTypeVendor(keepType1, address2)

            let result = await keepRegistry.getKeepTypeVendor.call(keepType1)
            assert.deepEqual(result, address2, "unexpected keep vendor address")
        })

        it("sets two keep types with different addresses", async () => {
            await keepRegistry.setKeepTypeVendor(keepType1, address1)
            await keepRegistry.setKeepTypeVendor(keepType2, address2)

            let result1 = await keepRegistry.getKeepTypeVendor.call(keepType1)
            assert.deepEqual(result1, address1, "unexpected keep vendor address")

            let result2 = await keepRegistry.getKeepTypeVendor.call(keepType2)
            assert.deepEqual(result2, address2, "unexpected keep vendor address")
        })

        it("sets two keep types with the same addresses", async () => {
            await keepRegistry.setKeepTypeVendor(keepType1, address1)
            await keepRegistry.setKeepTypeVendor(keepType2, address1)

            let result1 = await keepRegistry.getKeepTypeVendor.call(keepType1)
            assert.deepEqual(result1, address1, "unexpected keep vendor address")

            let result2 = await keepRegistry.getKeepTypeVendor.call(keepType2)
            assert.deepEqual(result2, address1, "unexpected keep vendor address")
        })

        it("fails with zero address", async () => {
            await keepRegistry.setKeepTypeVendor(keepType1, address1)

            try {
                await keepRegistry.setKeepTypeVendor(keepType1, address0)
                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, 'Vendor address cannot be zero')
            }

            let result = await keepRegistry.getKeepTypeVendor.call(keepType1)
            assert.deepEqual(result, address1, "unexpected keep vendor address")
        })

        it("cannot be called by non owner", async () => {
            try {
                await keepRegistry.setKeepTypeVendor.call(keepType1, address1, { from: accounts[1] })
                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, 'Ownable: caller is not the owner')
            }
        })
    })

    describe("getKeepTypeVendor", async () => {
        it("returns zero for not registered keep type", async () => {
            let result = await keepRegistry.getKeepTypeVendor.call("NOT EXISTING")
            assert.deepEqual(result, address0, "unexpected keep vendor address")
        })
    })

    describe("removeKeepType", async () => {
        before(async () => {
            await keepRegistry.setKeepTypeVendor(keepType1, address1)
            await keepRegistry.setKeepTypeVendor(keepType2, address2)
        })

        it("removes keep type address", async () => {
            await keepRegistry.removeKeepType(keepType1)

            let result = await keepRegistry.getKeepTypeVendor.call(keepType1)
            assert.deepEqual(result, address0, "unexpected keep vendor address")
        })

        it("doesn't fail for not registered keep type", async () => {
            await keepRegistry.removeKeepType("NOT EXISTING")
        })

        it("cannot be called by non owner", async () => {
            try {
                await keepRegistry.removeKeepType.call(keepType1, { from: accounts[1] })
                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, 'Ownable: caller is not the owner')
            }
        })
    })
})
