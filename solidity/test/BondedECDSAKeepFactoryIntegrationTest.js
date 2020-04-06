import {createSnapshot, restoreSnapshot} from "./helpers/snapshot"

const {time} = require("@openzeppelin/test-helpers")

import {mineBlocks} from "./helpers/mineBlocks"

const KeepToken = artifacts.require("KeepTokenIntegration")
const Registry = artifacts.require("Registry")
const BondedECDSAKeepFactoryStub = artifacts.require(
  "BondedECDSAKeepFactoryStub"
)
const KeepBonding = artifacts.require("KeepBonding")
const TokenStaking = artifacts.require("TokenStaking")
const BondedSortitionPool = artifacts.require("BondedSortitionPool")
const BondedSortitionPoolFactory = artifacts.require(
  "BondedSortitionPoolFactory"
)
const RandomBeaconStub = artifacts.require("RandomBeaconStub")
const BondedECDSAKeep = artifacts.require("BondedECDSAKeep")

const BN = web3.utils.BN

contract("BondedECDSAKeepFactory", async (accounts) => {
  let keepToken
  let tokenStaking
  let keepFactory
  let bondedSortitionPoolFactory
  let keepBonding
  let randomBeacon
  let signerPool

  const application = accounts[1]
  const members = [accounts[2], accounts[3], accounts[4]]
  const authorizers = [members[0], members[1], members[2]]

  const keepOwner = accounts[5]

  const groupSize = new BN(members.length)
  const threshold = groupSize

  const singleBond = new BN(1)
  const bond = singleBond.mul(groupSize)

  const initializationPeriod = new BN(100)
  const undelegationPeriod = 30

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("openKeep", async () => {
    let feeEstimate

    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()
      await registerMemberCandidates()

      feeEstimate = await keepFactory.openKeepFeeEstimate()
    })

    it("registers token staking delegated authority claim from keep", async () => {
      const keepAddress = await keepFactory.openKeep.call(
        groupSize,
        threshold,
        keepOwner,
        bond,
        {from: application, value: feeEstimate}
      )

      await keepFactory.openKeep(groupSize, threshold, keepOwner, bond, {
        from: application,
        value: feeEstimate,
      })

      assert.equal(
        await tokenStaking.getAuthoritySource(keepAddress),
        keepFactory.address,
        "invalid token staking authority source"
      )
    })
  })

  async function initializeNewFactory() {
    keepToken = await KeepToken.new()
    const registry = await Registry.new()

    bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
    tokenStaking = await TokenStaking.new(
      keepToken.address,
      registry.address,
      initializationPeriod,
      undelegationPeriod
    )

    keepBonding = await KeepBonding.new(registry.address, tokenStaking.address)
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
    const minimumStake = await keepFactory.minimumStake.call()
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
})
