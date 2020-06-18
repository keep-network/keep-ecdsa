import {
  getETHBalancesFromList,
  getERC20BalancesFromList,
  addToBalances,
} from "./helpers/listBalanceUtils"

import {mineBlocks} from "./helpers/mineBlocks"
import {createSnapshot, restoreSnapshot} from "./helpers/snapshot"
import {duration, increaseTime} from "./helpers/increaseTime"

const {expectRevert, constants, time} = require("@openzeppelin/test-helpers")

const KeepRegistry = artifacts.require("KeepRegistry")
const BondedECDSAKeepStub = artifacts.require("BondedECDSAKeepStub")
const TestToken = artifacts.require("TestToken")
const KeepBonding = artifacts.require("KeepBonding")
const TestEtherReceiver = artifacts.require("TestEtherReceiver")
const TokenStakingStub = artifacts.require("TokenStakingStub")
const TokenGrantStub = artifacts.require("TokenGrantStub")
const BondedECDSAKeepCloneFactory = artifacts.require(
  "BondedECDSAKeepCloneFactory"
)

const truffleAssert = require("truffle-assertions")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

contract("BondedECDSAKeep", (accounts) => {
  const bondCreator = accounts[0]
  const owner = accounts[1]
  const nonOwner = accounts[2]
  const members = [accounts[2], accounts[3], accounts[4]]
  const authorizers = [accounts[2], accounts[3], accounts[4]]
  const signingPool = accounts[5]
  const honestThreshold = 1

  const stakeLockDuration = time.duration.days(180)

  let registry
  let keepBonding
  let tokenStaking
  let tokenGrant
  let bondedECDSAKeepStubMaster
  let keep
  let factoryStub
  let memberStake

  async function newKeep(
    owner,
    members,
    honestThreshold,
    memberStake,
    stakeLockDuration,
    tokenStaking,
    keepBonding,
    keepFactory
  ) {
    const startBlock = await web3.eth.getBlockNumber()

    await factoryStub.newKeep(
      owner,
      members,
      honestThreshold,
      memberStake,
      stakeLockDuration,
      tokenStaking,
      keepBonding,
      keepFactory
    )

    const events = await factoryStub.getPastEvents("BondedECDSAKeepCreated", {
      fromBlock: startBlock,
      toBlock: "latest",
    })
    assert.lengthOf(
      events,
      1,
      "unexpected length of BondedECDSAKeepCreated events"
    )
    const keepAddress = events[0].returnValues.keepAddress

    return await BondedECDSAKeepStub.at(keepAddress)
  }

  before(async () => {
    registry = await KeepRegistry.new()
    tokenStaking = await TokenStakingStub.new()
    tokenGrant = await TokenGrantStub.new()
    keepBonding = await KeepBonding.new(
      registry.address,
      tokenStaking.address,
      tokenGrant.address
    )
    bondedECDSAKeepStubMaster = await BondedECDSAKeepStub.new()
    factoryStub = await BondedECDSAKeepCloneFactory.new(
      bondedECDSAKeepStubMaster.address
    )

    await registry.approveOperatorContract(bondCreator)

    memberStake = await factoryStub.minimumStake.call()
    await stakeOperators(members, memberStake)

    await keepBonding.authorizeSortitionPoolContract(members[0], signingPool, {
      from: authorizers[0],
    })
    await keepBonding.authorizeSortitionPoolContract(members[1], signingPool, {
      from: authorizers[1],
    })
    await keepBonding.authorizeSortitionPoolContract(members[2], signingPool, {
      from: authorizers[2],
    })
  })

  beforeEach(async () => {
    await createSnapshot()

    keep = await newKeep(
      owner,
      members,
      honestThreshold,
      memberStake,
      stakeLockDuration,
      tokenStaking.address,
      keepBonding.address,
      factoryStub.address
    )
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("initialize", async () => {
    it("succeeds", async () => {
      const expectedKeyGenerationTimeout = new BN(9000) // 9000 = 150*60 = 2.5h in seconds
      const expectedSigningTimeout = new BN(5400) // 5400 = 90 * 60 = 1.5h in seconds

      keep = await BondedECDSAKeepStub.new()
      await keep.initialize(
        owner,
        members,
        honestThreshold,
        memberStake,
        stakeLockDuration,
        tokenStaking.address,
        keepBonding.address,
        factoryStub.address
      )

      expect(
        await keep.keyGenerationTimeout(),
        "incorrect key generation timeout"
      ).to.eq.BN(expectedKeyGenerationTimeout)
      expect(await keep.signingTimeout(), "incorrect signing timeout").to.eq.BN(
        expectedSigningTimeout
      )
    })

    it("claims token staking delegated authority", async () => {
      keep = await BondedECDSAKeepStub.new()
      await keep.initialize(
        owner,
        members,
        honestThreshold,
        memberStake,
        stakeLockDuration,
        tokenStaking.address,
        keepBonding.address,
        factoryStub.address
      )

      assert.equal(
        await tokenStaking.delegatedAuthority(),
        factoryStub.address,
        "incorrect token staking delegated authority"
      )
    })

    it("locks members token stakes", async () => {
      keep = await BondedECDSAKeepStub.new()
      await keep.initialize(
        owner,
        members,
        honestThreshold,
        memberStake,
        stakeLockDuration,
        tokenStaking.address,
        keepBonding.address,
        factoryStub.address
      )

      for (let i = 0; i < members.length; i++) {
        expect(
          await tokenStaking.operatorLocks(members[i]),
          "incorrect token stake lock"
        ).to.eq.BN(stakeLockDuration)
      }
    })

    it("reverts if called for the second time", async () => {
      // first call was a part of beforeEach
      await expectRevert(
        keep.initialize(
          owner,
          members,
          honestThreshold,
          memberStake,
          stakeLockDuration,
          tokenStaking.address,
          keepBonding.address,
          factoryStub.address
        ),
        "Contract already initialized"
      )
    })
  })

  describe("sign", async () => {
    const publicKey =
      "0x657282135ed640b0f5a280874c7e7ade110b5c3db362e0552e6b7fff2cc8459328850039b734db7629c31567d7fc5677536b7fc504e967dc11f3f2289d3d4051"
    const digest =
      "0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da"

    it("reverts if public key was not set", async () => {
      await expectRevert(
        keep.sign(digest, {from: owner}),
        "Public key was not set yet"
      )
    })

    it("emits event", async () => {
      await submitMembersPublicKeys(publicKey)

      const res = await keep.sign(digest, {from: owner})
      truffleAssert.eventEmitted(res, "SignatureRequested", (ev) => {
        return ev.digest == digest
      })
    })

    it("sets block number for digest", async () => {
      await submitMembersPublicKeys(publicKey)

      const signTx = await keep.sign(digest, {from: owner})

      const blockNumber = await keep.digests.call(digest)

      expect(blockNumber, "incorrect block number").to.eq.BN(
        signTx.receipt.blockNumber
      )
    })

    it("cannot be requested if keep is closed", async () => {
      await createMembersBonds(keep)

      await keep.closeKeep({from: owner})

      await expectRevert(keep.sign(digest, {from: owner}), "Keep is not active")
    })

    it("cannot be called by non-owner", async () => {
      await expectRevert(keep.sign(digest), "Caller is not the keep owner")
    })

    it("cannot be called by non-owner member", async () => {
      await expectRevert(
        keep.sign(digest, {from: members[0]}),
        "Caller is not the keep owner"
      )
    })

    it("cannot be requested if already in progress", async () => {
      await submitMembersPublicKeys(publicKey)

      await keep.sign(digest, {from: owner})

      await expectRevert(keep.sign("0x02", {from: owner}), "Signer is busy")
    })
  })

  describe("isAwaitingSignature", async () => {
    const digest1 =
      "0x54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b"
    const publicKey =
      "0x657282135ed640b0f5a280874c7e7ade110b5c3db362e0552e6b7fff2cc8459328850039b734db7629c31567d7fc5677536b7fc504e967dc11f3f2289d3d4051"
    const signatureR =
      "0x9b32c3623b6a16e87b4d3a56cd67c666c9897751e24a51518136185403b1cba2"
    const signatureS =
      "0x6f7c776efde1e382f2ecc99ec0db13534a70ee86bd91d7b3a4059bccbed5d70c"
    const signatureRecoveryID = 1

    const digest2 =
      "0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da"

    beforeEach(async () => {
      await submitMembersPublicKeys(publicKey)
    })

    it("returns false if signing was not requested", async () => {
      assert.isFalse(await keep.isAwaitingSignature(digest1))
    })

    it("returns true if signing was requested for the digest", async () => {
      await keep.sign(digest1, {from: owner})

      assert.isTrue(await keep.isAwaitingSignature(digest1))
    })

    it("returns false if signing was requested for other digest", async () => {
      await keep.sign(digest2, {from: owner})

      assert.isFalse(await keep.isAwaitingSignature(digest1))
    })

    it("returns false if valid signature has been already submitted", async () => {
      await keep.sign(digest1, {from: owner})

      await keep.submitSignature(signatureR, signatureS, signatureRecoveryID, {
        from: members[0],
      })

      assert.isFalse(await keep.isAwaitingSignature(digest1))
    })

    it("returns true if invalid signature was submitted before", async () => {
      await keep.sign(digest1, {from: owner})

      await expectRevert(
        keep.submitSignature(signatureR, signatureS, 0, {from: members[0]}),
        "Invalid signature"
      )

      assert.isTrue(await keep.isAwaitingSignature(digest1))
    })
  })

  describe("hasSigningTimedOut", async () => {
    const digest1 =
      "0x54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b"
    const publicKey =
      "0x657282135ed640b0f5a280874c7e7ade110b5c3db362e0552e6b7fff2cc8459328850039b734db7629c31567d7fc5677536b7fc504e967dc11f3f2289d3d4051"
    const signatureR =
      "0x9b32c3623b6a16e87b4d3a56cd67c666c9897751e24a51518136185403b1cba2"
    const signatureS =
      "0x6f7c776efde1e382f2ecc99ec0db13534a70ee86bd91d7b3a4059bccbed5d70c"
    const signatureRecoveryID = 1

    const digest2 =
      "0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da"

    let signingTimeout

    beforeEach(async () => {
      signingTimeout = await keep.signingTimeout.call()

      await submitMembersPublicKeys(publicKey)
    })

    it("returns false if signing was not requested", async () => {
      assert.isFalse(await keep.hasSigningTimedOut())
    })

    it("returns false if signing was requested and have not timed out yet", async () => {
      await keep.sign(digest1, {from: owner})

      await increaseTime(duration.seconds(signingTimeout - 1))

      assert.isFalse(await keep.hasSigningTimedOut())
    })

    it("returns true if signing was requested and have timed out", async () => {
      await keep.sign(digest2, {from: owner})

      await increaseTime(duration.seconds(signingTimeout))

      assert.isTrue(await keep.hasSigningTimedOut())
    })

    it("returns false if signing was requested and signature have been submitted", async () => {
      await keep.sign(digest1, {from: owner})

      await keep.submitSignature(signatureR, signatureS, signatureRecoveryID, {
        from: members[0],
      })

      await increaseTime(duration.seconds(signingTimeout))

      assert.isFalse(await keep.hasSigningTimedOut())
    })
  })

  describe("public key", () => {
    const publicKey0 =
      "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
    const publicKey1 =
      "0xa899b9539de2a6345dc2ebd14010fe6bcd5d38db9ed75cef4afc6fc68a4c45a4901970bbff307e69048b4d6edf960a6dd7bc5ba9b1cf1b4e0a1e319f68e0741a"
    const publicKey2 =
      "0x999999539de2a6345dc2ebd14010fe6bcd5d38db9ed75cef4afc6fc68a4c45a4901970bbff307e69048b4d6edf960a6dd7bc5ba9b1cf1b4e0a1e319f68e0741a"
    const publicKey3 =
      "0x657282135ed640b0f5a280874c7e7ade110b5c3db362e0552e6b7fff2cc8459328850039b734db7629c31567d7fc5677536b7fc504e967dc11f3f2289d3d4051"

    it("get public key before it is set", async () => {
      const publicKey = await keep.getPublicKey.call()

      assert.equal(publicKey, undefined, "public key should not be set")
    })

    it("get the public key when all members submitted", async () => {
      await submitMembersPublicKeys(publicKey1)

      const publicKey = await keep.getPublicKey.call()

      assert.equal(publicKey, publicKey1, "incorrect public key")
    })

    describe("submitPublicKey", async () => {
      it("does not emit an event nor sets the key when keys were not submitted by all members", async () => {
        const res = await keep.submitPublicKey(publicKey1, {from: members[1]})
        truffleAssert.eventNotEmitted(res, "PublicKeyPublished")

        const publicKey = await keep.getPublicKey.call()
        assert.equal(publicKey, null, "incorrect public key")
      })

      it("does not emit an event nor sets the key when inconsistent keys were submitted by all members", async () => {
        const startBlock = await web3.eth.getBlockNumber()

        await keep.submitPublicKey(publicKey1, {from: members[0]})
        await keep.submitPublicKey(publicKey2, {from: members[1]})
        await keep.submitPublicKey(publicKey3, {from: members[2]})

        assert.isNull(await keep.getPublicKey(), "incorrect public key")

        assert.isEmpty(
          await keep.getPastEvents("PublicKeyPublished", {
            fromBlock: startBlock,
            toBlock: "latest",
          }),
          "unexpected events emitted"
        )
      })

      it("does not emit an event nor sets the key when just one inconsistent key was submitted", async () => {
        const startBlock = await web3.eth.getBlockNumber()

        await keep.submitPublicKey(publicKey1, {from: members[0]})
        await keep.submitPublicKey(publicKey2, {from: members[1]})
        await keep.submitPublicKey(publicKey1, {from: members[2]})

        assert.isNull(await keep.getPublicKey(), "incorrect public key")

        assert.isEmpty(
          await keep.getPastEvents("PublicKeyPublished", {
            fromBlock: startBlock,
            toBlock: "latest",
          }),
          "unexpected events emitted"
        )
      })

      it("emits event and sets a key when all submitted keys are the same", async () => {
        let res = await keep.submitPublicKey(publicKey1, {from: members[2]})
        truffleAssert.eventNotEmitted(res, "PublicKeyPublished")

        res = await keep.submitPublicKey(publicKey1, {from: members[0]})
        truffleAssert.eventNotEmitted(res, "PublicKeyPublished")

        const actualPublicKey = await keep.getPublicKey()
        assert.isNull(actualPublicKey, "incorrect public key")

        res = await keep.submitPublicKey(publicKey1, {from: members[1]})
        truffleAssert.eventEmitted(res, "PublicKeyPublished", {
          publicKey: publicKey1,
        })

        assert.equal(
          await keep.getPublicKey(),
          publicKey1,
          "incorrect public key"
        )
      })

      it("does not allow submitting public key more than once", async () => {
        await keep.submitPublicKey(publicKey0, {from: members[0]})

        await expectRevert(
          keep.submitPublicKey(publicKey1, {from: members[0]}),
          "Member already submitted a public key"
        )
      })

      it("does not emit conflict event for first all zero key ", async () => {
        // Event should not be emitted as other keys are not yet submitted.
        const res = await keep.submitPublicKey(publicKey0, {from: members[2]})
        truffleAssert.eventNotEmitted(res, "ConflictingPublicKeySubmitted")

        // One event should be emitted as just one other key is submitted.
        const startBlock = await web3.eth.getBlockNumber()
        await keep.submitPublicKey(publicKey1, {from: members[0]})
        assert.lengthOf(
          await keep.getPastEvents("ConflictingPublicKeySubmitted", {
            fromBlock: startBlock,
            toBlock: "latest",
          }),
          1,
          "unexpected events"
        )
      })

      it("emits conflict events for submitted values", async () => {
        // In this test it's important that members don't submit in the same order
        // as they are registered in the keep. We want to stress this scenario
        // and confirm that logic works correctly in such sophisticated scenario.

        // First member submits a public key, there are not conflicts.
        let startBlock = await web3.eth.getBlockNumber()
        await keep.submitPublicKey(publicKey1, {from: members[2]})
        assert.lengthOf(
          await keep.getPastEvents("ConflictingPublicKeySubmitted", {
            fromBlock: startBlock,
            toBlock: "latest",
          }),
          0,
          "unexpected events for the first submitted key"
        )
        await mineBlocks(1)

        // Second member submits another public key, there is one conflict.
        startBlock = await web3.eth.getBlockNumber()
        await keep.submitPublicKey(publicKey2, {from: members[1]})
        assert.lengthOf(
          await keep.getPastEvents("ConflictingPublicKeySubmitted", {
            fromBlock: startBlock,
            toBlock: "latest",
          }),
          1,
          "unexpected events for the second submitted key"
        )
        await mineBlocks(1)

        // Third member submits yet another public key, there are two conflicts.
        startBlock = await web3.eth.getBlockNumber()
        await keep.submitPublicKey(publicKey3, {from: members[0]})
        assert.lengthOf(
          await keep.getPastEvents("ConflictingPublicKeySubmitted", {
            fromBlock: startBlock,
            toBlock: "latest",
          }),
          2,
          "unexpected events for the third submitted key"
        )

        assert.isNull(await keep.getPublicKey(), "incorrect public key")
      })

      it("reverts when public key already set", async () => {
        await submitMembersPublicKeys(publicKey1)

        await expectRevert(
          keep.submitPublicKey(publicKey1, {from: members[0]}),
          "Member already submitted a public key"
        )
      })

      it("cannot be called by non-member", async () => {
        await expectRevert(
          keep.submitPublicKey(publicKey1),
          "Caller is not the keep member"
        )
      })

      it("cannot be called by non-member owner", async () => {
        await expectRevert(
          keep.submitPublicKey(publicKey1, {from: owner}),
          "Caller is not the keep member"
        )
      })

      it("cannot be different than 64 bytes", async () => {
        const badPublicKey =
          "0x9b9539de2a6345dc2ebd14010fe6bcd5d38db9ed75cef4afc6fc68a4c45a4901970bbff307e69048b4d6edf960a6dd7bc5ba9b1cf1b4e0a1e319f68e0741a"
        await keep.submitPublicKey(publicKey1, {from: members[1]})
        await expectRevert(
          keep.submitPublicKey(badPublicKey, {from: members[2]}),
          "Public key must be 64 bytes long"
        )
      })

      it("can be called just before the timeout", async () => {
        const keyGenerationTimeout = await keep.keyGenerationTimeout.call()

        await keep.submitPublicKey(publicKey1, {from: members[0]})
        await keep.submitPublicKey(publicKey1, {from: members[1]})

        // 5 seconds before the timeout
        await increaseTime(duration.seconds(keyGenerationTimeout - 5))

        await keep.submitPublicKey(publicKey1, {from: members[2]})

        assert.equal(
          await keep.getPublicKey(),
          publicKey1,
          "incorrect public key"
        )
      })

      it("cannot be called after timeout", async () => {
        const keyGenerationTimeout = await keep.keyGenerationTimeout.call()

        await keep.submitPublicKey(publicKey1, {from: members[0]})
        await keep.submitPublicKey(publicKey1, {from: members[1]})

        await increaseTime(duration.seconds(keyGenerationTimeout))

        await expectRevert(
          keep.submitPublicKey(publicKey1, {from: members[2]}),
          "Key generation timeout elapsed"
        )
      })
    })
  })

  describe("checkBondAmount", () => {
    it("should return bond amount", async () => {
      const expectedBondsSum = await createMembersBonds(keep)

      const actual = await keep.checkBondAmount.call()

      expect(actual).to.eq.BN(expectedBondsSum, "incorrect bond amount")
    })
  })

  describe("seizeSignerBonds", () => {
    const digest =
      "0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da"
    const publicKey =
      "0xa899b9539de2a6345dc2ebd14010fe6bcd5d38db9ed75cef4afc6fc68a4c45a4901970bbff307e69048b4d6edf960a6dd7bc5ba9b1cf1b4e0a1e319f68e0741a"

    let initialBondsSum

    beforeEach(async () => {
      await submitMembersPublicKeys(publicKey)
      initialBondsSum = await createMembersBonds(keep)
    })

    it("should seize signer bond", async () => {
      const expectedBondsSum = initialBondsSum
      const ownerBalanceBefore = await web3.eth.getBalance(owner)

      expect(await keep.checkBondAmount()).to.eq.BN(
        expectedBondsSum,
        "incorrect bond amount before seizure"
      )

      const gasPrice = await web3.eth.getGasPrice()

      const txHash = await keep.seizeSignerBonds({from: owner})
      const seizedSignerBondsFee = new BN(txHash.receipt.gasUsed).mul(
        new BN(gasPrice)
      )
      const ownerBalanceDiff = new BN(await web3.eth.getBalance(owner))
        .add(seizedSignerBondsFee)
        .sub(new BN(ownerBalanceBefore))

      expect(ownerBalanceDiff).to.eq.BN(
        expectedBondsSum,
        "incorrect owner balance"
      )

      expect(await keep.checkBondAmount()).to.eq.BN(
        0,
        "should zero all the bonds"
      )
    })

    it("terminates a keep", async () => {
      await keep.seizeSignerBonds({from: owner})
      assert.isTrue(await keep.isTerminated(), "keep should be terminated")
      assert.isFalse(await keep.isActive(), "keep should no longer be active")
      assert.isFalse(await keep.isClosed(), "keep should not be closed")
    })

    it("releases locks on members token stakes", async () => {
      await keep.seizeSignerBonds({from: owner})

      for (let i = 0; i < members.length; i++) {
        expect(
          await tokenStaking.operatorLocks(members[i]),
          "incorrect token stake lock"
        ).to.eq.BN(-1)
      }
    })

    it("emits an event", async () => {
      truffleAssert.eventEmitted(
        await keep.seizeSignerBonds({from: owner}),
        "KeepTerminated"
      )
    })

    it("can be called only by owner", async () => {
      await expectRevert(
        keep.seizeSignerBonds({from: nonOwner}),
        "Caller is not the keep owner"
      )
    })

    it("reverts when signing is in progress", async () => {
      keep.sign(digest, {from: owner})
      await expectRevert(
        keep.seizeSignerBonds({from: owner}),
        "Requested signing has not timed out yet"
      )
    })

    it("reverts when signing was requested but has not timed out yet", async () => {
      keep.sign(digest, {from: owner})

      const signingTimeout = await keep.signingTimeout.call()
      await increaseTime(duration.seconds(signingTimeout - 1))

      await expectRevert(
        keep.seizeSignerBonds({from: owner}),
        "Requested signing has not timed out yet"
      )
    })

    it("succeeds when signing was requested but timed out", async () => {
      keep.sign(digest, {from: owner})

      const signingTimeout = await keep.signingTimeout.call()
      await increaseTime(duration.seconds(signingTimeout))

      await keep.seizeSignerBonds({from: owner})
    })

    it("reverts when already seized", async () => {
      await keep.seizeSignerBonds({from: owner})

      await expectRevert(
        keep.seizeSignerBonds({from: owner}),
        "Keep is not active"
      )
    })

    it("reverts when already closed", async () => {
      await keep.closeKeep({from: owner})

      await expectRevert(
        keep.seizeSignerBonds({from: owner}),
        "Keep is not active"
      )
    })
  })

  describe("checkSignatureFraud", () => {
    // Private key: 0x937FFE93CFC943D1A8FC0CB8BAD44A978090A4623DA81EEFDFF5380D0A290B41
    // Public key:
    //  Curve: secp256k1
    //  X: 0x9A0544440CC47779235CCB76D669590C2CD20C7E431F97E17A1093FAF03291C4
    //  Y: 0x73E661A208A8A565CA1E384059BD2FF7FF6886DF081FF1229250099D388C83DF

    // TODO: Extract test data to a test data file and use them consistently across other tests.

    const publicKey1 =
      "0x9a0544440cc47779235ccb76d669590c2cd20c7e431f97e17a1093faf03291c473e661a208a8a565ca1e384059bd2ff7ff6886df081ff1229250099d388c83df"
    const preimage1 =
      "0xfdaf2feee2e37c24f2f8d15ad5814b49ba04b450e67b859976cbf25c13ea90d8"
    // hash256Digest1 = sha256(preimage1)
    const hash256Digest1 =
      "0x8bacaa8f02ef807f2f61ae8e00a5bfa4528148e0ae73b2bd54b71b8abe61268e"

    const signature1 = {
      R: "0xedc074a86380cc7e2e4702eaf1bec87843bc0eb7ebd490f5bdd7f02493149170",
      S: "0x3f5005a26eb6f065ea9faea543e5ddb657d13892db2656499a43dfebd6e12efc",
      V: 28,
    }

    const hash256Digest2 =
      "0x14a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b"
    const preimage2 = "0x1111636820506f7a6e616e"

    it("reverts if public key was not set", async () => {
      await expectRevert(
        keep.checkSignatureFraud.call(
          signature1.V,
          signature1.R,
          signature1.S,
          hash256Digest1,
          preimage1
        ),
        "Public key was not set yet"
      )
    })

    it("should return true when signature is valid but was not requested", async () => {
      await submitMembersPublicKeys(publicKey1)

      await keep.sign(hash256Digest2, {from: owner})

      const res = await keep.checkSignatureFraud.call(
        signature1.V,
        signature1.R,
        signature1.S,
        hash256Digest1,
        preimage1
      )

      assert.isTrue(
        res,
        "Signature is fraudulent because is valid but was not requested."
      )
    })

    it("should return an error when preimage does not match digest", async () => {
      await submitMembersPublicKeys(publicKey1)

      await keep.sign(hash256Digest2, {from: owner})

      await expectRevert(
        keep.checkSignatureFraud.call(
          signature1.V,
          signature1.R,
          signature1.S,
          hash256Digest1,
          preimage2
        ),
        "Signed digest does not match sha256 hash of the preimage"
      )
    })

    it("should return false when signature is invalid and was requested", async () => {
      await submitMembersPublicKeys(publicKey1)

      const badSignatureR =
        "0x1112c3623b6a16e87b4d3a56cd67c666c9897751e24a51518136185403b1cba2"

      assert.isFalse(
        await keep.checkSignatureFraud.call(
          signature1.V,
          badSignatureR,
          signature1.S,
          hash256Digest1,
          preimage1
        ),
        "signature is not fraudulent"
      )
    })

    it("should return false when signature is invalid and was not requested", async () => {
      await submitMembersPublicKeys(publicKey1)

      await keep.sign(hash256Digest2, {from: owner})
      const badSignatureR =
        "0x1112c3623b6a16e87b4d3a56cd67c666c9897751e24a51518136185403b1cba2"

      assert.isFalse(
        await keep.checkSignatureFraud.call(
          signature1.V,
          badSignatureR,
          signature1.S,
          hash256Digest1,
          preimage1
        ),
        "signature is not fraudulent"
      )
    })

    it("should return false when signature is valid and was requested", async () => {
      await submitMembersPublicKeys(publicKey1)

      await keep.sign(hash256Digest1, {from: owner})

      assert.isFalse(
        await keep.checkSignatureFraud.call(
          signature1.V,
          signature1.R,
          signature1.S,
          hash256Digest1,
          preimage1
        ),
        "signature is not fraudulent"
      )
    })

    it("should return false when signature is valid, was requested and was submitted", async () => {
      await submitMembersPublicKeys(publicKey1)

      await keep.sign(hash256Digest1, {from: owner})
      await keep.submitSignature(
        signature1.R,
        signature1.S,
        signature1.V - 27,
        {
          from: members[0],
        }
      )

      assert.isFalse(
        await keep.checkSignatureFraud.call(
          signature1.V,
          signature1.R,
          signature1.S,
          hash256Digest1,
          preimage1
        ),
        "signature is not fraudulent"
      )
    })
  })

  describe("submitSignatureFraud", () => {
    // Private key: 0x937FFE93CFC943D1A8FC0CB8BAD44A978090A4623DA81EEFDFF5380D0A290B41
    // Public key:
    //  Curve: secp256k1
    //  X: 0x9A0544440CC47779235CCB76D669590C2CD20C7E431F97E17A1093FAF03291C4
    //  Y: 0x73E661A208A8A565CA1E384059BD2FF7FF6886DF081FF1229250099D388C83DF

    // TODO: Extract test data to a test data file and use them consistently across other tests.

    const publicKey1 =
      "0x9a0544440cc47779235ccb76d669590c2cd20c7e431f97e17a1093faf03291c473e661a208a8a565ca1e384059bd2ff7ff6886df081ff1229250099d388c83df"
    const preimage1 =
      "0xfdaf2feee2e37c24f2f8d15ad5814b49ba04b450e67b859976cbf25c13ea90d8"
    // hash256Digest1 = sha256(preimage1)
    const hash256Digest1 =
      "0x8bacaa8f02ef807f2f61ae8e00a5bfa4528148e0ae73b2bd54b71b8abe61268e"

    const signature1 = {
      R: "0xedc074a86380cc7e2e4702eaf1bec87843bc0eb7ebd490f5bdd7f02493149170",
      S: "0x3f5005a26eb6f065ea9faea543e5ddb657d13892db2656499a43dfebd6e12efc",
      V: 28,
    }

    it("should return true and slash members when the signature is a fraud", async () => {
      await submitMembersPublicKeys(publicKey1)

      const res = await keep.submitSignatureFraud.call(
        signature1.V,
        signature1.R,
        signature1.S,
        hash256Digest1,
        preimage1
      )

      await keep.submitSignatureFraud(
        signature1.V,
        signature1.R,
        signature1.S,
        hash256Digest1,
        preimage1
      )

      assert.isTrue(res, "incorrect returned result")

      for (let i = 0; i < members.length; i++) {
        const actualStake = await tokenStaking.eligibleStake(
          members[i],
          constants.ZERO_ADDRESS
        )
        expect(actualStake).to.eq.BN(0, `incorrect stake for member ${i}`)
      }
    })

    it("should revert when the signature is not a fraud", async () => {
      await submitMembersPublicKeys(publicKey1)

      await keep.sign(hash256Digest1, {from: owner})

      await expectRevert(
        keep.submitSignatureFraud(
          signature1.V,
          signature1.R,
          signature1.S,
          hash256Digest1,
          preimage1
        ),
        "Signature is not fraudulent"
      )

      for (let i = 0; i < members.length; i++) {
        const actualStake = await tokenStaking.eligibleStake(
          members[i],
          constants.ZERO_ADDRESS
        )
        expect(actualStake).to.eq.BN(
          memberStake,
          `incorrect stake for member ${i}`
        )
      }
    })

    it("reverts if called for closed keep", async () => {
      await keep.publicMarkAsClosed()

      await expectRevert(
        keep.submitSignatureFraud(
          signature1.V,
          signature1.R,
          signature1.S,
          hash256Digest1,
          preimage1
        ),
        "Keep is not active"
      )
    })

    it("reverts if called for terminated keep", async () => {
      await keep.publicMarkAsTerminated()

      await expectRevert(
        keep.submitSignatureFraud(
          signature1.V,
          signature1.R,
          signature1.S,
          hash256Digest1,
          preimage1
        ),
        "Keep is not active"
      )
    })

    it("does not revert if slashing failed", async () => {
      await submitMembersPublicKeys(publicKey1)

      await tokenStaking.setSlashingShouldFail(true)

      const res = await keep.submitSignatureFraud.call(
        signature1.V,
        signature1.R,
        signature1.S,
        hash256Digest1,
        preimage1
      )

      const tx = await keep.submitSignatureFraud(
        signature1.V,
        signature1.R,
        signature1.S,
        hash256Digest1,
        preimage1
      )

      assert.isTrue(res, "incorrect returned result")
      truffleAssert.eventEmitted(tx, "SlashingFailed")
    })
  })

  describe("submitSignature", () => {
    const digest =
      "0x54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b"
    const publicKey =
      "0x657282135ed640b0f5a280874c7e7ade110b5c3db362e0552e6b7fff2cc8459328850039b734db7629c31567d7fc5677536b7fc504e967dc11f3f2289d3d4051"
    const signatureR =
      "0x9b32c3623b6a16e87b4d3a56cd67c666c9897751e24a51518136185403b1cba2"
    const signatureS =
      "0x6f7c776efde1e382f2ecc99ec0db13534a70ee86bd91d7b3a4059bccbed5d70c"
    const signatureRecoveryID = 1

    // This malleable signature details corresponds to the signature above but
    // it's calculated that `S` is in the higher half of curve's order. We use
    // this to check malleability.
    // `malleableS = secp256k1.N - signatureS`
    // To read more see [EIP-2](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2.md).
    const malleableS =
      "0x90838891021e1c7d0d1336613f24ecab703dee5ff1b6c8881bccc2c011606a35"
    const malleableRecoveryID = 0

    beforeEach(async () => {
      await submitMembersPublicKeys(publicKey)
    })

    it("emits an event", async () => {
      await keep.sign(digest, {from: owner})

      const res = await keep.submitSignature(
        signatureR,
        signatureS,
        signatureRecoveryID,
        {from: members[0]}
      )

      truffleAssert.eventEmitted(res, "SignatureSubmitted", (ev) => {
        return (
          ev.digest == digest &&
          ev.r == signatureR &&
          ev.s == signatureS &&
          ev.recoveryID == signatureRecoveryID
        )
      })
    })

    it("clears signing lock after submission", async () => {
      await keep.sign(digest, {from: owner})

      await keep.submitSignature(signatureR, signatureS, signatureRecoveryID, {
        from: members[0],
      })

      await keep.sign(digest, {from: owner})
    })

    it("can be called just before the timeout", async () => {
      await keep.sign(digest, {from: owner})

      const signingTimeout = await keep.signingTimeout.call()

      await increaseTime(duration.seconds(signingTimeout - 1))

      await keep.submitSignature(signatureR, signatureS, signatureRecoveryID, {
        from: members[0],
      })
    })

    it("cannot be called after the timeout passed", async () => {
      await keep.sign(digest, {from: owner})

      const signingTimeout = await keep.signingTimeout.call()

      await increaseTime(duration.seconds(signingTimeout))

      await expectRevert(
        keep.submitSignature(signatureR, signatureS, signatureRecoveryID, {
          from: members[0],
        }),
        "Signing timeout elapsed"
      )
    })

    it("cannot be submitted if signing was not requested", async () => {
      await expectRevert(
        keep.submitSignature(signatureR, signatureS, signatureRecoveryID, {
          from: members[0],
        }),
        "Not awaiting a signature"
      )
    })

    describe("validates signature", async () => {
      beforeEach(async () => {
        await keep.sign(digest, {from: owner})
      })

      it("rejects recovery ID out of allowed range", async () => {
        await expectRevert(
          keep.submitSignature(signatureR, signatureS, 4, {from: members[0]}),
          "Recovery ID must be one of {0, 1, 2, 3}"
        )
      })

      it("rejects invalid signature", async () => {
        await expectRevert(
          keep.submitSignature(signatureR, signatureS, 0, {from: members[0]}),
          "Invalid signature"
        )
      })

      it("rejects malleable signature", async () => {
        try {
          await keep.submitSignature(
            signatureR,
            malleableS,
            malleableRecoveryID,
            {from: members[0]}
          )
          assert(false, "Test call did not error as expected")
        } catch (e) {
          assert.include(
            e.message,
            "Malleable signature - s should be in the low half of secp256k1 curve's order"
          )
        }
      })
    })

    it("cannot be called by non-member", async () => {
      await keep.sign(digest, {from: owner})

      await expectRevert(
        keep.submitSignature(signatureR, signatureS, signatureRecoveryID),
        "Caller is not the keep member"
      )
    })

    it("cannot be called by non-member owner", async () => {
      await keep.sign(digest, {from: owner})

      await expectRevert(
        keep.submitSignature(signatureR, signatureS, signatureRecoveryID, {
          from: owner,
        }),
        "Caller is not the keep member"
      )
    })
  })

  describe("closeKeep", () => {
    const digest =
      "0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da"
    const publicKey =
      "0xa899b9539de2a6345dc2ebd14010fe6bcd5d38db9ed75cef4afc6fc68a4c45a4901970bbff307e69048b4d6edf960a6dd7bc5ba9b1cf1b4e0a1e319f68e0741a"

    const bondValue0 = new BN(10)
    const bondValue1 = new BN(20)
    const bondValue2 = new BN(20)

    beforeEach(async () => {
      await createMembersBonds(keep, bondValue0, bondValue1, bondValue2)
      await submitMembersPublicKeys(publicKey)
    })

    it("emits an event", async () => {
      truffleAssert.eventEmitted(
        await keep.closeKeep({from: owner}),
        "KeepClosed"
      )
    })

    it("marks keep as closed", async () => {
      await keep.closeKeep({from: owner})
      assert.isTrue(await keep.isClosed(), "keep should be closed")
      assert.isFalse(await keep.isActive(), "keep should no longer be active")
      assert.isFalse(await keep.isTerminated(), "keep should not be terminated")
    })

    it("frees members bonds", async () => {
      await keep.closeKeep({from: owner})

      expect(await keep.checkBondAmount()).to.eq.BN(
        0,
        "incorrect bond amount for keep"
      )

      expect(
        await keepBonding.availableUnbondedValue(
          members[0],
          bondCreator,
          signingPool
        )
      ).to.eq.BN(bondValue0, "incorrect unbonded amount for member 0")

      expect(
        await keepBonding.availableUnbondedValue(
          members[1],
          bondCreator,
          signingPool
        )
      ).to.eq.BN(bondValue1, "incorrect unbonded amount for member 1")

      expect(
        await keepBonding.availableUnbondedValue(
          members[2],
          bondCreator,
          signingPool
        )
      ).to.eq.BN(bondValue2, "incorrect unbonded amount for member 2")
    })

    it("releases locks on members token stakes", async () => {
      await keep.closeKeep({from: owner})

      for (let i = 0; i < members.length; i++) {
        expect(
          await tokenStaking.operatorLocks(members[i]),
          "incorrect token stake lock"
        ).to.eq.BN(-1)
      }
    })

    it("reverts when signing is in progress", async () => {
      keep.sign(digest, {from: owner})

      await expectRevert(
        keep.closeKeep({from: owner}),
        "Requested signing has not timed out yet"
      )
    })

    it("reverts when signing was requested but has not timed out yet", async () => {
      keep.sign(digest, {from: owner})

      const signingTimeout = await keep.signingTimeout.call()
      await increaseTime(duration.seconds(signingTimeout - 1))

      await expectRevert(
        keep.closeKeep({from: owner}),
        "Requested signing has not timed out yet"
      )
    })

    it("succeeds when signing was requested but timed out", async () => {
      keep.sign(digest, {from: owner})

      const signingTimeout = await keep.signingTimeout.call()
      await increaseTime(duration.seconds(signingTimeout))

      await keep.closeKeep({from: owner})
    })

    it("cannot be called by non-owner", async () => {
      await expectRevert(keep.closeKeep(), "Caller is not the keep owner")
    })

    it("reverts when already closed", async () => {
      await keep.closeKeep({from: owner})

      await expectRevert(keep.closeKeep({from: owner}), "Keep is not active")
    })

    it("reverts when already seized", async () => {
      await keep.seizeSignerBonds({from: owner})

      await expectRevert(keep.closeKeep({from: owner}), "Keep is not active")
    })
  })

  describe("returnPartialSignerBonds", async () => {
    const singleReturnedBondValue = new BN(2000)
    const allReturnedBondsValue = singleReturnedBondValue.mul(
      new BN(members.length)
    )

    const member1Unbonded = new BN(100)
    const member2Unbounded = new BN(200)
    const member3Unbounded = new BN(700)

    beforeEach(async () => {
      await authorizeBondCreator()
      await depositForBonding(
        member1Unbonded,
        member2Unbounded,
        member3Unbounded
      )
    })

    it("correctly distributes ETH", async () => {
      await keep.returnPartialSignerBonds({value: allReturnedBondsValue})

      const member1UnbondedAfter = await keepBonding.availableUnbondedValue(
        members[0],
        bondCreator,
        signingPool
      )
      const member2UnbondedAfter = await keepBonding.availableUnbondedValue(
        members[1],
        bondCreator,
        signingPool
      )
      const member3UnbondedAfter = await keepBonding.availableUnbondedValue(
        members[2],
        bondCreator,
        signingPool
      )

      expect(
        member1UnbondedAfter,
        "incorrect unbounded balance for member 1"
      ).to.eq.BN(2100) // 2000 + 100
      expect(
        member2UnbondedAfter,
        "incorrect unbounded balance for member 2"
      ).to.eq.BN(2200) // 2000 + 200
      expect(
        member3UnbondedAfter,
        "incorrect unbounded balance for member 3"
      ).to.eq.BN(2700) // 2000 + 700
    })

    it("correctly handles remainder", async () => {
      const remainder = new BN(2)

      await keep.returnPartialSignerBonds({
        value: allReturnedBondsValue.add(remainder),
      })

      const member1UnbondedAfter = await keepBonding.availableUnbondedValue(
        members[0],
        bondCreator,
        signingPool
      )
      const member2UnbondedAfter = await keepBonding.availableUnbondedValue(
        members[1],
        bondCreator,
        signingPool
      )
      const member3UnbondedAfter = await keepBonding.availableUnbondedValue(
        members[2],
        bondCreator,
        signingPool
      )

      expect(
        member1UnbondedAfter,
        "incorrect unbounded balance for member 1"
      ).to.eq.BN(2100) // 2000 + 100
      expect(
        member2UnbondedAfter,
        "incorrect unbounded balance for member 2"
      ).to.eq.BN(2200) // 2000 + 200
      expect(
        member3UnbondedAfter,
        "incorrect unbounded balance for member 3"
      ).to.eq.BN(2702) // 2000 + 700 + 2
    })

    it("reverts with zero value", async () => {
      await expectRevert(
        keep.returnPartialSignerBonds({value: 0}),
        "Partial signer bond must be non-zero"
      )
    })

    it("reverts with zero value per member", async () => {
      await expectRevert(
        keep.returnPartialSignerBonds({value: members.length - 1}),
        "Partial signer bond must be non-zero"
      )
    })
  })

  describe("distributeETHReward", async () => {
    const singleValue = new BN(1000)
    const ethValue = singleValue.mul(new BN(members.length))

    it("emits event", async () => {
      const startBlock = await web3.eth.getBlockNumber()

      const res = await keep.distributeETHReward({value: ethValue})
      truffleAssert.eventEmitted(res, "ETHRewardDistributed", (event) => {
        return web3.utils.toBN(event.amount).eq(ethValue)
      })

      assert.lengthOf(
        await keep.getPastEvents("ETHRewardDistributed", {
          fromBlock: startBlock,
          toBlock: "latest",
        }),
        1,
        "unexpected events emitted"
      )
    })

    it("correctly distributes ETH", async () => {
      const initialBalances = await getETHBalancesFromList(members)

      await keep.distributeETHReward({value: ethValue})

      const newBalances = await getETHBalancesFromList(members)

      assert.deepEqual(newBalances, initialBalances)

      expect(
        await web3.eth.getBalance(keep.address),
        "incorrect keep balance"
      ).to.eq.BN(ethValue)

      expect(
        await keep.getMemberETHBalance(members[0]),
        "incorrect member 0 balance"
      ).to.eq.BN(singleValue)

      expect(
        await keep.getMemberETHBalance(members[1]),
        "incorrect member 1 balance"
      ).to.eq.BN(singleValue)

      expect(
        await keep.getMemberETHBalance(members[2]),
        "incorrect member 2 balance"
      ).to.eq.BN(singleValue)
    })

    it("correctly handles unused remainder", async () => {
      const expectedRemainder = new BN(members.length - 1)
      const valueWithRemainder = ethValue.add(expectedRemainder)

      await keep.distributeETHReward({value: valueWithRemainder})

      expect(
        await web3.eth.getBalance(keep.address),
        "incorrect keep balance"
      ).to.eq.BN(valueWithRemainder)

      expect(
        await keep.getMemberETHBalance(members[0]),
        "incorrect member 0 balance"
      ).to.eq.BN(singleValue)

      expect(
        await keep.getMemberETHBalance(members[1]),
        "incorrect member 1 balance"
      ).to.eq.BN(singleValue)

      expect(
        await keep.getMemberETHBalance(members[2]),
        "incorrect member 2 balance"
      ).to.eq.BN(singleValue.add(expectedRemainder))
    })

    it("reverts with zero value", async () => {
      await expectRevert(
        keep.distributeETHReward(),
        "Dividend value must be non-zero"
      )
    })

    it("reverts with zero dividend", async () => {
      const msgValue = members.length - 1
      await expectRevert(
        keep.distributeETHReward({value: msgValue}),
        "Dividend value must be non-zero"
      )
    })
  })

  describe("withdraw", async () => {
    const singleValue = new BN(1000)
    const ethValue = singleValue.mul(new BN(members.length))

    beforeEach(async () => {
      await keep.distributeETHReward({value: ethValue})
    })

    it("correctly transfers value", async () => {
      const initialMemberBalance = new BN(await web3.eth.getBalance(members[0]))

      await keep.withdraw(members[0])

      expect(
        await web3.eth.getBalance(keep.address),
        "incorrect keep balance"
      ).to.eq.BN(ethValue.sub(singleValue))

      expect(
        await keep.getMemberETHBalance(members[0]),
        "incorrect member balance"
      ).to.eq.BN(0)

      expect(
        await web3.eth.getBalance(members[0]),
        "incorrect member account balance"
      ).to.eq.BN(initialMemberBalance.add(singleValue))
    })

    it("sends ETH to beneficiary", async () => {
      const valueWithRemainder = ethValue.add(new BN(1))

      const member1 = accounts[2]
      const member2 = accounts[3]
      const beneficiary = accounts[4]

      const testMembers = [member1, member2]

      const accountsInTest = [member1, member2, beneficiary]
      const expectedBalances = [
        new BN(await web3.eth.getBalance(member1)),
        new BN(await web3.eth.getBalance(member2)),
        new BN(await web3.eth.getBalance(beneficiary)).add(valueWithRemainder),
      ]

      const keep = await newKeep(
        owner,
        testMembers,
        honestThreshold,
        memberStake,
        stakeLockDuration,
        tokenStaking.address,
        keepBonding.address,
        factoryStub.address
      )

      await tokenStaking.setBeneficiary(member1, beneficiary)
      await tokenStaking.setBeneficiary(member2, beneficiary)

      await keep.distributeETHReward({value: valueWithRemainder})

      await keep.withdraw(member1)
      expect(
        await keep.getMemberETHBalance(member1),
        "incorrect member 1 balance"
      ).to.eq.BN(0)

      await keep.withdraw(member2)
      expect(
        await keep.getMemberETHBalance(member2),
        "incorrect member 2 balance"
      ).to.eq.BN(0)

      // Check balances of all keep members' and beneficiary.
      const newBalances = await getETHBalancesFromList(accountsInTest)
      assert.deepEqual(newBalances, expectedBalances)
    })

    it("reverts in case of zero balance", async () => {
      const member = members[0]

      const keep = await newKeep(
        owner,
        [member],
        honestThreshold,
        memberStake,
        stakeLockDuration,
        tokenStaking.address,
        keepBonding.address,
        factoryStub.address
      )

      await expectRevert(keep.withdraw(member), "No funds to withdraw")
    })

    it("reverts in case of transfer failure", async () => {
      const etherReceiver = await TestEtherReceiver.new()
      await etherReceiver.setShouldFail(true)

      const member = etherReceiver.address // a receiver which we expect to reject the transfer

      const keep = await newKeep(
        owner,
        [member],
        honestThreshold,
        memberStake,
        stakeLockDuration,
        tokenStaking.address,
        keepBonding.address,
        factoryStub.address
      )

      await keep.distributeETHReward({value: ethValue})

      await expectRevert(keep.withdraw(member), "Transfer failed")

      // Check balances of keep member's account.
      expect(
        await web3.eth.getBalance(member),
        "incorrect member's account balance"
      ).to.eq.BN(0)

      // Check that value which failed transfer remained in the keep contract.
      expect(
        await web3.eth.getBalance(keep.address),
        "incorrect keep's account balance"
      ).to.eq.BN(ethValue)
    })
  })

  describe("distributeERC20Reward", async () => {
    const erc20Value = new BN(2000).mul(new BN(members.length))
    let token

    beforeEach(async () => {
      token = await TestToken.new()
    })

    it("correctly distributes ERC20", async () => {
      await initializeTokens(token, keep, accounts[0], erc20Value)

      const expectedBalances = addToBalances(
        await getERC20BalancesFromList(members, token),
        erc20Value / members.length
      )

      await keep.distributeERC20Reward(token.address, erc20Value)

      const newBalances = await getERC20BalancesFromList(members, token)

      assert.equal(newBalances.toString(), expectedBalances.toString())
    })

    it("emits an event", async () => {
      await initializeTokens(token, keep, accounts[0], erc20Value)

      const startBlock = await web3.eth.getBlockNumber()

      const res = await keep.distributeERC20Reward(token.address, erc20Value)
      truffleAssert.eventEmitted(res, "ERC20RewardDistributed", (event) => {
        return (
          token.address == event.token &&
          web3.utils.toBN(event.amount).eq(erc20Value)
        )
      })

      assert.lengthOf(
        await keep.getPastEvents("ERC20RewardDistributed", {
          fromBlock: startBlock,
          toBlock: "latest",
        }),
        1,
        "unexpected events emitted"
      )
    })

    it("correctly handles remainder", async () => {
      const expectedRemainder = new BN(members.length - 1)
      const valueWithRemainder = erc20Value.add(expectedRemainder)

      await initializeTokens(token, keep, accounts[0], valueWithRemainder)

      const expectedBalances = addToBalances(
        await getERC20BalancesFromList(members, token),
        erc20Value / members.length
      )

      const lastMemberIndex = members.length - 1
      expectedBalances[lastMemberIndex] = expectedBalances[lastMemberIndex].add(
        expectedRemainder
      )

      await keep.distributeERC20Reward(token.address, valueWithRemainder)

      const newBalances = await getERC20BalancesFromList(members, token)

      assert.equal(newBalances.toString(), expectedBalances.toString())

      expect(await token.balanceOf(keep.address)).to.eq.BN(
        0,
        "incorrect keep balance"
      )
    })

    it("fails with insufficient approval", async () => {
      await expectRevert(
        keep.distributeERC20Reward(token.address, erc20Value),
        "SafeERC20: low-level call failed"
      )
    })

    it("fails with zero value", async () => {
      await expectRevert(
        keep.distributeERC20Reward(token.address, 0),
        "Dividend value must be non-zero"
      )
    })

    it("reverts with zero dividend", async () => {
      const value = members.length - 1

      await initializeTokens(token, keep, accounts[0], value)

      await expectRevert(
        keep.distributeERC20Reward(token.address, value),
        "Dividend value must be non-zero"
      )
    })

    it("sends ERC20 to beneficiary", async () => {
      const valueWithRemainder = erc20Value.add(new BN(1))

      const member1 = accounts[2]
      const member2 = accounts[3]
      const beneficiary = accounts[4]

      const testMembers = [member1, member2]

      const accountsInTest = [member1, member2, beneficiary]
      const expectedBalances = [
        new BN(await token.balanceOf(member1)),
        new BN(await token.balanceOf(member2)),
        new BN(await token.balanceOf(beneficiary)).add(valueWithRemainder),
      ]

      keep = await newKeep(
        owner,
        testMembers,
        honestThreshold,
        memberStake,
        stakeLockDuration,
        tokenStaking.address,
        keepBonding.address,
        factoryStub.address
      )

      await initializeTokens(token, keep, accounts[0], valueWithRemainder)

      await tokenStaking.setBeneficiary(member1, beneficiary)
      await tokenStaking.setBeneficiary(member2, beneficiary)

      await keep.distributeERC20Reward(token.address, valueWithRemainder)

      // Check balances of all keep members' and beneficiary.
      const newBalances = await getERC20BalancesFromList(accountsInTest, token)
      assert.equal(newBalances.toString(), expectedBalances.toString())
    })

    async function initializeTokens(token, keep, account, amount) {
      await token.mint(account, amount)
      await token.approve(keep.address, amount)
    }
  })

  async function submitMembersPublicKeys(publicKey) {
    await keep.submitPublicKey(publicKey, {from: members[0]})
    await keep.submitPublicKey(publicKey, {from: members[1]})
    await keep.submitPublicKey(publicKey, {from: members[2]})
  }

  async function stakeOperators(members, stakeBalance) {
    for (let i = 0; i < members.length; i++) {
      await tokenStaking.setBalance(members[i], stakeBalance)
    }
  }

  async function authorizeBondCreator() {
    await tokenStaking.authorizeOperatorContract(members[0], bondCreator)
    await tokenStaking.authorizeOperatorContract(members[1], bondCreator)
    await tokenStaking.authorizeOperatorContract(members[2], bondCreator)
  }

  async function depositForBonding(member1Value, member2Value, member3Value) {
    await keepBonding.deposit(members[0], {value: member1Value})
    await keepBonding.deposit(members[1], {value: member2Value})
    await keepBonding.deposit(members[2], {value: member3Value})
  }

  async function createMembersBonds(keep, bond1, bond2, bond3) {
    await authorizeBondCreator()

    const bondValue1 = bond1 || new BN(100)
    const bondValue2 = bond2 || new BN(200)
    const bondValue3 = bond3 || new BN(300)

    const referenceID = web3.utils.toBN(web3.utils.padLeft(keep.address, 32))

    await depositForBonding(bondValue1, bondValue2, bondValue3)

    await keepBonding.createBond(
      members[0],
      keep.address,
      referenceID,
      bondValue1,
      signingPool
    )
    await keepBonding.createBond(
      members[1],
      keep.address,
      referenceID,
      bondValue2,
      signingPool
    )
    await keepBonding.createBond(
      members[2],
      keep.address,
      referenceID,
      bondValue3,
      signingPool
    )

    return bondValue1.add(bondValue2).add(bondValue3)
  }
})
