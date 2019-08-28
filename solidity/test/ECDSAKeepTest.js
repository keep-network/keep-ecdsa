const ECDSAKeep = artifacts.require('./ECDSAKeep.sol')
const truffleAssert = require('truffle-assertions')

contract('ECDSAKeep', (accounts) => {
  const owner = accounts[1]
  const members = [accounts[2], accounts[3]]
  const honestThreshold = 5

  describe('#constructor', async () => {
    it('succeeds', async () => {
      let keep = await ECDSAKeep.new(
        owner,
        members,
        honestThreshold
      )

      assert(web3.utils.isAddress(keep.address), 'invalid keep address')
    })
  })

  describe('#sign', async () => {
    const digest = '0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da'
    let keep

    before(async () => {
      keep = await ECDSAKeep.new(owner, members, honestThreshold)
    })

    it('emits event', async () => {
      const digest = '0xbb0b57005f01018b19c278c55273a60118ffdd3e5790ccc8a48cad03907fa521'

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
    const expectedPublicKey = '0xa899b9539de2a6345dc2ebd14010fe6bcd5d38db9ed75cef4afc6fc68a4c45a4901970bbff307e69048b4d6edf960a6dd7bc5ba9b1cf1b4e0a1e319f68e0741a'

    let keep

    beforeEach(async () => {
      keep = await ECDSAKeep.new(owner, members, honestThreshold);
    })

    it('get public key before it is set', async () => {
      let publicKey = await keep.getPublicKey.call()

      assert.equal(publicKey, undefined, 'incorrect public key')
    })

    it('set public key and get it', async () => {
      await keep.setPublicKey(expectedPublicKey, { from: members[0] })

      let publicKey = await keep.getPublicKey.call()

      assert.equal(
        publicKey,
        expectedPublicKey,
        'incorrect public key'
      )
    })

    describe('setPublicKey', async () => {
      it('emits an event', async () => {
        let res = await keep.setPublicKey(expectedPublicKey, { from: members[0] })

        truffleAssert.eventEmitted(res, 'PublicKeyPublished', (ev) => {
          return ev.publicKey == expectedPublicKey
        })
      })

      describe('cannot be called by non member', async () => {
        it('cannot be called by default account', async () => {
          try {
            await keep.setPublicKey(expectedPublicKey)
            assert(false, 'Test call did not error as expected')
          } catch (e) {
            assert.include(e.message, 'Caller is not the keep member')
          }
        })

        it('cannot be called by owner', async () => {
          try {
            await keep.setPublicKey(expectedPublicKey, { from: owner })
            assert(false, 'Test call did not error as expected')
          } catch (e) {
            assert.include(e.message, 'Caller is not the keep member')
          }
        })
      })
    })
  })

  describe('submitSignature', () => {
    const digest = '0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da'
    const publicKey = '0xda7deb4491a25c6d3ef36e291c825aa7cfbba776f54c4373ebee9d15cb4b988858879cb2c60c61cdaac1524e9f5e5b0ca2e52e0172cdec932ee6180383eca830'
    const signatureR = '0x7a31b539b7b67ae1d998f1929d95a89b497b600db36e8b474f4b9e703decd82a'
    const signatureS = '0x597fbca252abeedab90ef2e8b8088eeef9e2e3ac824b4626396fc46f2934f185'

    let keep

    before(async () => {
      keep = await ECDSAKeep.new(owner, members, honestThreshold)

      await keep.setPublicKey(publicKey, { from: members[0] })
    })

    it('emits an event', async () => {
      let res = await keep.submitSignature(digest, signatureR, signatureS, { from: members[0] })

      truffleAssert.eventEmitted(res, 'SignatureSubmitted', (ev) => {
        return ev.digest == digest && ev.r == signatureR && ev.s == signatureS
      })
    })

    describe('validates signature', async () => {
      const bytes31 = '0x01020304050607080910111213141516171819202122232425262728293031'
      const bytes33 = '0x010203040506070809101112131415161718192021222324252627282930313233'

      it('rejects R of 31-byte length', async () => {
        try {
          await keep.submitSignature(digest, bytes31, signatureS, { from: members[0] })
          assert(false, 'Test call did not error as expected')
        } catch (e) {
          assert.include(e.message, "Signature's R value is not 32-bytes long")
        }
      })

      it('rejects R of 33-byte length', async () => {
        try {
          await keep.submitSignature(digest, bytes33, signatureS, { from: members[0] })
          assert(false, 'Test call did not error as expected')
        } catch (e) {
          assert.include(e.message, "Signature's R value is not 32-bytes long")
        }
      })

      it('rejects S of 31-byte length', async () => {
        try {
          await keep.submitSignature(digest, signatureR, bytes31, { from: members[0] })
          assert(false, 'Test call did not error as expected')
        } catch (e) {
          assert.include(e.message, "Signature's S value is not 32-bytes long")
        }
      })

      it('rejects S of 33-byte length', async () => {
        try {
          await keep.submitSignature(digest, signatureR, bytes33, { from: members[0] })
          assert(false, 'Test call did not error as expected')
        } catch (e) {
          assert.include(e.message, "Signature's S value is not 32-bytes long")
        }
      })
    })

    describe('cannot be called by non member', async () => {
      it('cannot be called by default account', async () => {
        try {
          await keep.submitSignature(digest, signatureR, signatureS)
          assert(false, 'Test call did not error as expected')
        } catch (e) {
          assert.include(e.message, 'Caller is not the keep member')
        }
      })

      it('cannot be called by owner', async () => {
        try {
          await keep.submitSignature(digest, signatureR, signatureS, { from: owner })
          assert(false, 'Test call did not error as expected')
        } catch (e) {
          assert.include(e.message, 'Caller is not the keep member')
        }
      })
    })
  })
})
