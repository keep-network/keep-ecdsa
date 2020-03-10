import { createSnapshot, restoreSnapshot } from './helpers/snapshot'

const Registry = artifacts.require('Registry')
const BondedECDSAKeepVendor = artifacts.require('BondedECDSAKeepVendor')
const BondedECDSAKeepVendorImplV1 = artifacts.require(
  'BondedECDSAKeepVendorImplV1'
)
const { expectRevert } = require('openzeppelin-test-helpers')

contract('BondedECDSAKeepVendorImplV1', async accounts => {
  const address0 = '0x0000000000000000000000000000000000000000'
  const address1 = '0xF2D3Af2495E286C7820643B963FB9D34418c871d'
  const address2 = '0x4566716c07617c5854fe7dA9aE5a1219B19CCd27'
  const address3 = '0x65ea55c1f10491038425725dc00dffeab2a1e28a'

  let registry, keepVendor

  before(async () => {
    registry = await Registry.new()
  })

  describe('registerFactory', async () => {
    before(async () => {
      const bondedECDSAKeepVendorImplV1 = await BondedECDSAKeepVendorImplV1.new()
      const bondedECDSAKeepVendorProxy = await BondedECDSAKeepVendor.new(
        bondedECDSAKeepVendorImplV1.address
      )
      keepVendor = await BondedECDSAKeepVendorImplV1.at(
        bondedECDSAKeepVendorProxy.address
      )

      await keepVendor.initialize(registry.address)
      await registry.setOperatorContractUpgrader(
        keepVendor.address,
        accounts[0]
      )
      await registry.approveOperatorContract(address0)
      await registry.approveOperatorContract(address1)
      await registry.approveOperatorContract(address2)
    })

    beforeEach(async () => {
      await createSnapshot()
    })

    afterEach(async () => {
      await restoreSnapshot()
    })

    it('registers one factory address', async () => {
      let expectedResult = address1

      await keepVendor.registerFactory(address1)

      await assertFactory(expectedResult)
    })

    it('does not register factory with zero address', async () => {
      await expectRevert(
        keepVendor.registerFactory(address0),
        'Incorrect factory address'
      )
    })

    it('does not register factory not approved in registry', async () => {
      await expectRevert(
        keepVendor.registerFactory(address3),
        'Factory contract is not approved'
      )
    })

    it('replaces previous factory address', async () => {
      await keepVendor.registerFactory(address1)
      await keepVendor.registerFactory(address2)

      await assertFactory(address2)
    })

    it('cannot be called by non-owner', async () => {
      await expectRevert(
        keepVendor.registerFactory(address1, { from: accounts[1] }),
        'Caller is not operator contract upgrader'
      )
    })

    async function assertFactory(expectedFactory) {
      let actualFactory = await keepVendor.selectFactory.call()
      assert.equal(
        actualFactory,
        expectedFactory,
        'unexpected registered factory'
      )
    }
  })
})
