const BondedECDSAKeepVendorImplV1 = artifacts.require('BondedECDSAKeepVendorImplV1')
const { expectRevert } = require('openzeppelin-test-helpers');

contract("BondedECDSAKeepVendorImplV1", async accounts => {
    const address0 = "0x0000000000000000000000000000000000000000"
    const address1 = "0xF2D3Af2495E286C7820643B963FB9D34418c871d"
    const address2 = "0x4566716c07617c5854fe7dA9aE5a1219B19CCd27"

    let keepVendor

    beforeEach(async () => {
        keepVendor = await BondedECDSAKeepVendorImplV1.new()
    })

    describe("keep factory registration", async () => {

        it("registers one factory address", async () => {
            let expectedResult = address1

            await keepVendor.registerFactory(address1)

            assertFactory(expectedResult)
        })

        it("does not register factory with zero address", async () => {
            await expectRevert(
                keepVendor.registerFactory(address0),
                "Incorrect factory address"
            )
        })

        it("replaces previous factory address", async () => {
            await keepVendor.registerFactory(address1)
            await keepVendor.registerFactory(address2)

            assertFactory(address2)
        })

        it("cannot be called by non-owner", async () => {
            await expectRevert(
                keepVendor.registerFactory(address1, { from: accounts[1] }),
                "caller is not the owner"
            )
        })

        async function assertFactory(expectedFactory) {
            let actualFactory = await keepVendor.selectFactory.call()
            assert.equal(actualFactory, expectedFactory, "unexpected registered factory")
        }
    })
})
