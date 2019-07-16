const ECDSAKeep = artifacts.require('./ECDSAKeep.sol')

const truffleAssert = require('truffle-assertions')

contract('ECDSAKeep', function(accounts) {
  describe('#constructor', async function() {
    it('succeeds', async () => {
      const owner = accounts[0]
      const members = [owner]
      const threshold = 1

      const instance = await ECDSAKeep.new(
        owner,
        members,
        threshold
      )

      expect(instance.address).to.be.not.empty
    })
  })

  describe('#sign', async function() {
    let instance

    before(async () => {
      const owner = accounts[0]
      const members = [owner]
      const threshold = 1

      instance = await ECDSAKeep.new(
        owner,
        members,
        threshold
      )
    })

    it('emits event', async () => {
      const res = await instance.sign('0x00')
      truffleAssert.eventEmitted(res, 'SignatureRequested', (ev) => {
        return ev.digest == '0x00'
      })
    })
  })

  describe('public key', () => {
    const expectedPublicKey = web3.utils.hexToBytes('0x67656e657261746564207075626c6963206b6579')
    const owner = '0xbc4862697a1099074168d54A555c4A60169c18BD'
    const members = ['0x774700a36A96037936B8666dCFdd3Fb6687b08cb']
    const honestThreshold = 5

    it('get public key before it is set', async () => {
      const keep = await ECDSAKeep.new(owner, members, honestThreshold)

      const publicKey = await keep.getPublicKey.call().catch((err) => {
        assert.fail(`ecdsa keep creation failed: ${err}`)
      })

      assert.equal(publicKey, undefined, 'incorrect public key')
    })

    it('set public key and get it', async () => {
      const keep = await ECDSAKeep.new(owner, members, honestThreshold)

      await keep.setPublicKey(expectedPublicKey).catch((err) => {
        assert.fail(`ecdsa keep creation failed: ${err}`)
      })

      const publicKey = await keep.getPublicKey.call().catch((err) => {
        assert.fail(`cannot get public key: ${err}`)
      })

      assert.equal(
        publicKey,
        web3.utils.bytesToHex(expectedPublicKey),
        'incorrect public key'
      )
    })
  })
})
