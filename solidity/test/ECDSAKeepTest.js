const ECDSAKeep = artifacts.require('./ECDSAKeep.sol')
const truffleAssert = require('truffle-assertions')

contract('ECDSAKeep', (accounts) => {
  const owner = accounts[1]
  const members = [accounts[2], accounts[3]]
  const threshold = 5

  describe('#constructor', async () => {
    it('succeeds', async () => {
      let keep = await ECDSAKeep.new(
        owner,
        members,
        threshold
      )

      assert(web3.utils.isAddress(keep.address), 'invalid keep address')
    })
  })

  describe.only('#sign', async () => {
    const digest = '0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da'
    let keep

    before(async () => {
      keep = await ECDSAKeep.new(owner, members, threshold)
    })

    it('emits event', async () => {
      let res = await keep.sign(digest, { from: owner })
      truffleAssert.eventEmitted(res, 'SignatureRequested', (ev) => {
        return ev.digest == digest
      })
    })

    describe('cannot be called by non owner', async () => {
      it('cannot be called by default account', async () => {
        try {
          await keep.sign(digest)
          assert(false, 'Test call did not error as expected')
        } catch (e) {
          assert.include(e.message, 'Ownable: caller is not the owner.')
        }
      })

      it('cannot be called by member', async () => {
        try {
          await keep.sign(digest, { from: members[0] })
          assert(false, 'Test call did not error as expected')
        } catch (e) {
          assert.include(e.message, 'Ownable: caller is not the owner.')
        }
      })
    })
  })

  describe('public key', () => {
    const publicKey = web3.utils.hexToBytes('0xa899b9539de2a6345dc2ebd14010fe6bcd5d38db9ed75cef4afc6fc68a4c45a4901970bbff307e69048b4d6edf960a6dd7bc5ba9b1cf1b4e0a1e319f68e0741a')

    let keep

    beforeEach(async () => {
      keep = await ECDSAKeep.new(owner, members, threshold)
    })

    it('get public key before it is set', async () => {
      let publicKey = await keep.getPublicKey.call()

      assert.equal(publicKey, undefined, 'incorrect public key')
    })

    it.only('set public key and get it', async () => {
      await keep.setPublicKey(publicKey, { from: members[0] })

      const result = await keep.getPublicKey.call()

      assert.equal(
        result,
        web3.utils.bytesToHex(publicKey),
        'incorrect public key'
      )

      describe('setPublicKey', async () => {
        describe('cannot be called by non member', async () => {
          it('cannot be called by default account', async () => {
            try {
              await keep.setPublicKey(publicKey)
              assert(false, 'Test call did not error as expected')
            } catch (e) {
              assert.include(e.message, 'Caller is not the keep member')
            }
          })

          it('cannot be called by owner', async () => {
            try {
              await keep.setPublicKey(publicKey, { from: owner })
              assert(false, 'Test call did not error as expected')
            } catch (e) {
              assert.include(e.message, 'Caller is not the keep member')
            }
          })
        })
      })
    })
  })
})
