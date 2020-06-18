import {createSnapshot, restoreSnapshot} from "./helpers/snapshot"

const {time} = require("@openzeppelin/test-helpers")

import {mineBlocks} from "./helpers/mineBlocks"

const KeepToken = artifacts.require("KeepTokenIntegration")
const KeepRegistry = artifacts.require("KeepRegistry")
const BondedECDSAKeepFactoryStub = artifacts.require(
  "BondedECDSAKeepFactoryStub"
)
const KeepBonding = artifacts.require("KeepBonding")
const TokenStaking = artifacts.require("TokenStaking")
const TokenGrant = artifacts.require("TokenGrant")
const BondedSortitionPool = artifacts.require("BondedSortitionPool")
const BondedSortitionPoolFactory = artifacts.require(
  "BondedSortitionPoolFactory"
)
const RandomBeaconStub = artifacts.require("RandomBeaconStub")
const BondedECDSAKeep = artifacts.require("BondedECDSAKeep")

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

contract("BondedECDSAKeepFactory", async (accounts) => {
  let keepToken
  let tokenStaking
  let tokenGrant
  let keepFactory
  let bondedSortitionPoolFactory
  let keepBonding
  let randomBeacon
  let signerPool
  let feeEstimate

  const application = accounts[1]
  const members = [accounts[2], accounts[3], accounts[4]]
  const authorizers = [members[0], members[1], members[2]]

  const keepOwner = accounts[5]

  const groupSize = new BN(members.length)
  const threshold = groupSize

  const singleBond = new BN(1)
  const bond = singleBond.mul(groupSize)

  const stakeLockDuration = time.duration.days(180)

  const initializationPeriod = new BN(100)
  const undelegationPeriod = 30

  before(async () => {
    await initializeNewFactory()
    await initializeMemberCandidates()
    await registerMemberCandidates()

    feeEstimate = await keepFactory.openKeepFeeEstimate()
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("openKeep", async () => {
    it("registers token staking delegated authority claim from keep", async () => {
      const keepAddress = await keepFactory.openKeep.call(
        groupSize,
        threshold,
        keepOwner,
        bond,
        stakeLockDuration,
        {from: application, value: feeEstimate}
      )

      await keepFactory.openKeep(
        groupSize,
        threshold,
        keepOwner,
        bond,
        stakeLockDuration,
        {
          from: application,
          value: feeEstimate,
        }
      )

      assert.equal(
        await tokenStaking.getAuthoritySource(keepAddress),
        keepFactory.address,
        "invalid token staking authority source"
      )
    })

    it("locks member stakes", async () => {
      const tx = await keepFactory.openKeep(
        groupSize,
        threshold,
        keepOwner,
        bond,
        stakeLockDuration,
        {
          from: application,
          value: feeEstimate,
        }
      )
      const keepAddress = tx.logs[0].args.keepAddress

      const expectedExpirationTime = (await time.latest()).add(
        stakeLockDuration
      )

      for (let i = 0; i < members.length; i++) {
        const {creators, expirations} = await tokenStaking.getLocks(members[i])

        assert.deepEqual(
          creators,
          [keepAddress],
          "incorrect token lock creator"
        )

        expect(expirations[0], "incorrect token lock expiration time").to.eq.BN(
          expectedExpirationTime
        )
      }
    })
  })

  describe("closeKeep", async () => {
    it("releases locks on member stakes", async () => {
      const keep = await openKeep({from: keepOwner})

      await keep.closeKeep({from: keepOwner})

      for (let i = 0; i < members.length; i++) {
        const {creators, expirations} = await tokenStaking.getLocks(members[i])

        assert.isEmpty(creators, "incorrect token lock creator")

        assert.isEmpty(expirations, "incorrect token lock expiration time")
      }
    })
  })

  describe("seizeSignerBonds", async () => {
    it("releases locks on member stakes", async () => {
      const keep = await openKeep({from: keepOwner})

      await keep.seizeSignerBonds({from: keepOwner})

      for (let i = 0; i < members.length; i++) {
        const {creators, expirations} = await tokenStaking.getLocks(members[i])

        assert.isEmpty(creators, "incorrect token lock creator")

        assert.isEmpty(expirations, "incorrect token lock expiration time")
      }
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
      const keep = await openKeep()

      const initialStakes = []
      for (let i = 0; i < members.length; i++) {
        initialStakes[i] = web3.utils.toBN(
          await tokenStaking.eligibleStake(members[i], keepFactory.address)
        )
      }

      await submitMembersPublicKeys(keep, publicKey1)

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
        const expectedStake = initialStakes[i].sub(
          await tokenStaking.minimumStake.call()
        )

        const actualStake = await tokenStaking.eligibleStake(
          members[i],
          keepFactory.address
        )

        expect(actualStake).to.eq.BN(
          expectedStake,
          `incorrect stake for member ${i}`
        )
      }
    })
  })

  async function initializeNewFactory() {
    keepToken = await KeepToken.new()
    const registry = await KeepRegistry.new()

    bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
    tokenStaking = await TokenStaking.new(
      keepToken.address,
      registry.address,
      initializationPeriod,
      undelegationPeriod
    )
    tokenGrant = await TokenGrant.new(keepToken.address)

    keepBonding = await KeepBonding.new(
      registry.address,
      tokenStaking.address,
      tokenGrant.address
    )
    randomBeacon = await RandomBeaconStub.new()
    const bondedECDSAKeepMasterContract = await BondedECDSAKeep.new()
    keepFactory = await BondedECDSAKeepFactoryStub.new(
      bondedECDSAKeepMasterContract.address,
      bondedSortitionPoolFactory.address,
      tokenStaking.address,
      keepBonding.address,
      randomBeacon.address
    )

    await registry.approveOperatorContract(keepFactory.address)
  }

  async function stakeOperators(stakeBalance) {
    const tokenOwner = accounts[0]

    for (let i = 0; i < members.length; i++) {
      const beneficiary = tokenOwner
      const operator = members[i]
      const authorizer = authorizers[i]

      const delegation = Buffer.concat([
        Buffer.from(web3.utils.hexToBytes(beneficiary)),
        Buffer.from(web3.utils.hexToBytes(operator)),
        Buffer.from(web3.utils.hexToBytes(authorizer)),
      ])

      await keepToken.approveAndCall(
        tokenStaking.address,
        stakeBalance,
        delegation,
        {from: tokenOwner}
      )

      await time.increase(initializationPeriod.addn(1))
    }
  }

  async function initializeMemberCandidates(unbondedValue) {
    const minimumStake = await tokenStaking.minimumStake.call()
    await stakeOperators(minimumStake)

    signerPool = await keepFactory.createSortitionPool.call(application)
    await keepFactory.createSortitionPool(application)

    for (let i = 0; i < members.length; i++) {
      await tokenStaking.authorizeOperatorContract(
        members[i],
        keepFactory.address,
        {from: authorizers[i]}
      )
      await keepBonding.authorizeSortitionPoolContract(members[i], signerPool, {
        from: authorizers[i],
      })
    }

    const minimumBond = await keepFactory.minimumBond.call()
    const unbondedAmount = unbondedValue || minimumBond

    await depositMemberCandidates(unbondedAmount)
  }

  async function depositMemberCandidates(unbondedAmount) {
    for (let i = 0; i < members.length; i++) {
      await keepBonding.deposit(members[i], {value: unbondedAmount})
    }
  }

  async function registerMemberCandidates() {
    for (let i = 0; i < members.length; i++) {
      await keepFactory.registerMemberCandidate(application, {
        from: members[i],
      })
    }

    const pool = await BondedSortitionPool.at(signerPool)
    const initBlocks = await pool.operatorInitBlocks()
    await mineBlocks(initBlocks.add(new BN(1)))
  }

  async function openKeep() {
    const keepAddress = await keepFactory.openKeep.call(
      groupSize,
      threshold,
      keepOwner,
      bond,
      stakeLockDuration,
      {from: application, value: feeEstimate}
    )

    await keepFactory.openKeep(
      groupSize,
      threshold,
      keepOwner,
      bond,
      stakeLockDuration,
      {
        from: application,
        value: feeEstimate,
      }
    )

    return await BondedECDSAKeep.at(keepAddress)
  }

  async function submitMembersPublicKeys(keep, publicKey) {
    await keep.submitPublicKey(publicKey, {from: members[0]})
    await keep.submitPublicKey(publicKey, {from: members[1]})
    await keep.submitPublicKey(publicKey, {from: members[2]})
  }
})
