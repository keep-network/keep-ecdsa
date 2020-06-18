import {createSnapshot, restoreSnapshot} from "./helpers/snapshot"

const {expectRevert, time} = require("@openzeppelin/test-helpers")

import {mineBlocks} from "./helpers/mineBlocks"

const truffleAssert = require("truffle-assertions")

const KeepRegistry = artifacts.require("KeepRegistry")
const BondedECDSAKeepFactoryStub = artifacts.require(
  "BondedECDSAKeepFactoryStub"
)
const KeepBonding = artifacts.require("KeepBonding")
const TokenStakingStub = artifacts.require("TokenStakingStub")
const TokenGrantStub = artifacts.require("TokenGrantStub")
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
  let registry
  let tokenStaking
  let tokenGrant
  let keepFactory
  let bondedSortitionPoolFactory
  let keepBonding
  let randomBeacon
  let signerPool
  let minimumStake

  const application = accounts[1]
  const members = [accounts[2], accounts[3], accounts[4]]
  const authorizers = [members[0], members[1], members[2]]
  const notMember = accounts[5]

  const keepOwner = accounts[6]

  const groupSize = new BN(members.length)
  const threshold = new BN(groupSize - 1)

  const singleBond = new BN(1)
  const bond = singleBond.mul(groupSize)

  const stakeLockDuration = time.duration.days(180)

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("registerMemberCandidate", async () => {
    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()
    })

    it("inserts operator with the correct staking weight in the pool", async () => {
      const minimumStakeMultiplier = new BN("10")
      await stakeOperators(members, minimumStake.mul(minimumStakeMultiplier))

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      const pool = await BondedSortitionPool.at(signerPool)
      const actualWeight = await pool.getPoolWeight.call(members[0])

      // minimumStake * minimumStakeMultiplier / poolStakeWeightDivisor =
      // 200000 * 1e18 * 10 / 1e18 = 2000000
      const expectedWeight = new BN("2000000")

      expect(actualWeight).to.eq.BN(expectedWeight, "invalid staking weight")
    })

    it("inserts operators to the same pool", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })
      await keepFactory.registerMemberCandidate(application, {
        from: members[1],
      })

      const pool = await BondedSortitionPool.at(signerPool)
      assert.isTrue(
        await pool.isOperatorInPool(members[0]),
        "operator 1 is not in the pool"
      )
      assert.isTrue(
        await pool.isOperatorInPool(members[1]),
        "operator 2 is not in the pool"
      )
    })

    it("does not add an operator to the pool if it is already there", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      const pool = await BondedSortitionPool.at(signerPool)

      assert.isTrue(
        await pool.isOperatorInPool(members[0]),
        "operator is not in the pool"
      )

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      assert.isTrue(
        await pool.isOperatorInPool(members[0]),
        "operator is not in the pool"
      )
    })

    it("does not add an operator to the pool if it does not have a minimum stake", async () => {
      await stakeOperators(members, new BN("1"))

      await expectRevert(
        keepFactory.registerMemberCandidate(application, {from: members[0]}),
        "Operator not eligible"
      )
    })

    it("does not add an operator to the pool if it does not have a minimum bond", async () => {
      const minimumBond = await keepFactory.minimumBond.call()
      const availableUnbonded = await keepBonding.availableUnbondedValue(
        members[0],
        keepFactory.address,
        signerPool
      )
      const withdrawValue = availableUnbonded.sub(minimumBond).add(new BN(1))
      await keepBonding.withdraw(withdrawValue, members[0], {
        from: members[0],
      })

      await expectRevert(
        keepFactory.registerMemberCandidate(application, {from: members[0]}),
        "Operator not eligible"
      )
    })

    it("inserts operators to different pools", async () => {
      const application1 = "0x0000000000000000000000000000000000000001"
      const application2 = "0x0000000000000000000000000000000000000002"

      const signerPool1Address = await keepFactory.createSortitionPool.call(
        application1
      )
      await keepFactory.createSortitionPool(application1)
      const signerPool2Address = await keepFactory.createSortitionPool.call(
        application2
      )
      await keepFactory.createSortitionPool(application2)

      await keepBonding.authorizeSortitionPoolContract(
        members[0],
        signerPool1Address,
        {from: authorizers[0]}
      )
      await keepBonding.authorizeSortitionPoolContract(
        members[1],
        signerPool2Address,
        {from: authorizers[1]}
      )

      await keepFactory.registerMemberCandidate(application1, {
        from: members[0],
      })
      await keepFactory.registerMemberCandidate(application2, {
        from: members[1],
      })

      const signerPool1 = await BondedSortitionPool.at(signerPool1Address)

      assert.isTrue(
        await signerPool1.isOperatorInPool(members[0]),
        "operator 1 is not in the pool"
      )
      assert.isFalse(
        await signerPool1.isOperatorInPool(members[1]),
        "operator 2 is in the pool"
      )

      const signerPool2 = await BondedSortitionPool.at(signerPool2Address)

      assert.isFalse(
        await signerPool2.isOperatorInPool(members[0]),
        "operator 1 is in the pool"
      )
      assert.isTrue(
        await signerPool2.isOperatorInPool(members[1]),
        "operator 2 is not in the pool"
      )
    })
  })

  describe("createSortitionPool", async () => {
    before(async () => {
      await initializeNewFactory()
    })

    it("creates new sortition pool and emits an event", async () => {
      const sortitionPoolAddress = await keepFactory.createSortitionPool.call(
        application
      )

      const res = await keepFactory.createSortitionPool(application)
      truffleAssert.eventEmitted(res, "SortitionPoolCreated", {
        application: application,
        sortitionPool: sortitionPoolAddress,
      })
    })

    it("reverts when sortition pool already exists", async () => {
      await keepFactory.createSortitionPool(application)

      await expectRevert(
        keepFactory.createSortitionPool(application),
        "Sortition pool already exists"
      )
    })
  })

  describe("getSortitionPool", async () => {
    before(async () => {
      await initializeNewFactory()
    })

    it("returns address of sortition pool", async () => {
      const sortitionPoolAddress = await keepFactory.createSortitionPool.call(
        application
      )
      await keepFactory.createSortitionPool(application)

      const result = await keepFactory.getSortitionPool(application)
      assert.equal(
        result,
        sortitionPoolAddress,
        "incorrect sortition pool address"
      )
    })

    it("reverts if sortition pool does not exist", async () => {
      await expectRevert(
        keepFactory.getSortitionPool(application),
        "No pool found for the application"
      )
    })
  })

  describe("isOperatorRegistered", async () => {
    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()
    })

    it("returns true if the operator is registered for the application", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      assert.isTrue(
        await keepFactory.isOperatorRegistered(members[0], application)
      )
    })

    it("returns false if the operator is registered for another application", async () => {
      const application2 = "0x0000000000000000000000000000000000000002"

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      assert.isFalse(
        await keepFactory.isOperatorRegistered(members[0], application2)
      )
    })

    it("returns false if the operator is not registered for any application", async () => {
      assert.isFalse(
        await keepFactory.isOperatorRegistered(members[0], application)
      )
    })
  })

  describe("isOperatorUpToDate", async () => {
    const precision = new BN("1000000000000000000") // 1 KEEP = 10^18

    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()
    })

    it("returns true if the operator is up to date for the application", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      assert.isTrue(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns false if the operator stake dropped", async () => {
      const stake = minimumStake.muln(10) // 10 * 2000000 * 10^18
      await stakeOperators(members, stake)

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await stakeOperators(members, stake.subn(1)) // (10 * 2000000 * 10^18) - 1

      // Precision (pool weight divisor) is 10^18
      // ((10 * 200000 * 10^18)) / 10^18     =  2000000
      // ((10 * 200000 * 10^18) - 1) / 10^18 =~ 1999999.99
      // Ethereum uint256 division performs implicit floor:
      // The weight went down from 2,000,000 to 1,999,999
      assert.isFalse(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns false if the operator stake increased", async () => {
      const stake = minimumStake.muln(10) // 10 * 2000000 * 10^18
      await stakeOperators(members, stake)

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await stakeOperators(members, stake.add(precision)) // (10 * 2000000 * 10^18) + 10^18

      // Precision (pool weight divisor) is 10^18
      // ((10 * 200000 * 10^18)) / 10^18         = 2000000
      // ((10 * 200000 * 10^18) + 10^18) / 10^18 = 2000001
      // The weight went up from 2,000,000 to 2,000,001
      assert.isFalse(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns true if the operator stake increase is below weight precision", async () => {
      const stake = minimumStake.muln(10) // 10 * 2000000 * 10^18
      await stakeOperators(members, stake)

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await stakeOperators(members, stake.add(precision).subn(1)) // (10 * 2000000 * 10^18) + 10^18 - 1

      // Precision (pool weight divisor) is 10^18
      // ((10 * 200000 * 10^18)) / 10^18             = 2000000
      // ((10 * 200000 * 10^18) + 10^18 - 1) / 10^18 = 2000000.99
      // Ethereum uint256 division performs implicit floor:
      // The weight dit not change: 2,000,000 == floor(2,000,000.99)
      assert.isTrue(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns false if the operator stake dropped below minimum stake", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await stakeOperators(members, minimumStake.subn(1))

      assert.isFalse(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns false if the operator bonding value is below minimum", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await keepBonding.withdraw(new BN(1), members[0], {from: members[0]})

      assert.isFalse(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns true if the operator bonding value is above minimum", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await keepBonding.deposit(members[0], {value: new BN(1)})

      assert.isTrue(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("reverts if the operator is not registered for the application", async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()

      await expectRevert(
        keepFactory.isOperatorUpToDate(members[0], application),
        "Operator not registered for the application"
      )
    })
  })

  describe("updateOperatorStatus", async () => {
    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()
      await registerMemberCandidates()
    })

    it("revers if operator is up to date", async () => {
      await expectRevert(
        keepFactory.updateOperatorStatus(members[0], application),
        "Operator already up to date"
      )
    })

    it("removes operator if stake has changed below minimum", async () => {
      await stakeOperators(members, minimumStake.sub(new BN(1)))
      assert.isFalse(
        await keepFactory.isOperatorUpToDate(members[0], application),
        "unexpected status of the operator after stake change"
      )

      await keepFactory.updateOperatorStatus(members[0], application)

      await expectRevert(
        keepFactory.isOperatorUpToDate(members[0], application),
        "Operator not registered for the application"
      )
    })

    it("updates operator if stake has changed above minimum", async () => {
      // We multiply minimumStake as sortition pools expect multiplies of the
      // minimum stake to calculate stakers weight for eligibility.
      await stakeOperators(members, minimumStake.mul(new BN(2)))
      assert.isFalse(
        await keepFactory.isOperatorUpToDate(members[0], application),
        "unexpected status of the operator after stake change"
      )

      await keepFactory.updateOperatorStatus(members[0], application)

      assert.isTrue(
        await keepFactory.isOperatorUpToDate(members[0], application),
        "unexpected status of the operator after status update"
      )
    })

    it("removes operator if bonding value has changed below minimum", async () => {
      keepBonding.withdraw(new BN(1), members[0], {from: members[0]})
      assert.isFalse(
        await keepFactory.isOperatorUpToDate(members[0], application),
        "unexpected status of the operator after bonding value change"
      )

      await keepFactory.updateOperatorStatus(members[0], application)

      await expectRevert(
        keepFactory.isOperatorUpToDate(members[0], application),
        "Operator not registered for the application"
      )
    })

    it("updates operator if bonding value has changed above minimum", async () => {
      keepBonding.deposit(members[0], {value: new BN(1)})
      assert.isTrue(
        await keepFactory.isOperatorUpToDate(members[0], application),
        "unexpected status of the operator after bonding value change"
      )

      await expectRevert(
        keepFactory.updateOperatorStatus(members[0], application),
        "Operator already up to date"
      )
    })

    it("reverts if the operator is not registered for the application", async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()

      await expectRevert(
        keepFactory.updateOperatorStatus(members[0], application),
        "Operator not registered for the application"
      )
    })
  })

  describe("openKeep", async () => {
    let feeEstimate

    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()
      await registerMemberCandidates()

      feeEstimate = await keepFactory.openKeepFeeEstimate()
    })

    it("reverts if no member candidates are registered", async () => {
      await expectRevert(
        keepFactory.openKeep(
          groupSize,
          threshold,
          keepOwner,
          bond,
          stakeLockDuration,
          {
            value: feeEstimate,
          }
        ),
        "No signer pool for this application"
      )
    })

    it("reverts if bond equals zero", async () => {
      const bond = 0

      await expectRevert(
        keepFactory.openKeep(
          groupSize,
          threshold,
          keepOwner,
          bond,
          stakeLockDuration,
          {
            from: application,
            value: feeEstimate,
          }
        ),
        "Bond per member must be greater than zero"
      )
    })

    it("reverts if value is less than the required fee estimate", async () => {
      const insufficientFee = feeEstimate.sub(new BN(1))

      await expectRevert(
        keepFactory.openKeep(
          groupSize,
          threshold,
          keepOwner,
          bond,
          stakeLockDuration,
          {
            from: application,
            fee: insufficientFee,
          }
        ),
        "Insufficient payment for opening a new keep"
      )
    })

    it("opens keep with multiple members and emits event", async () => {
      const blockNumber = await web3.eth.getBlockNumber()

      const keepAddress = await keepFactory.openKeep.call(
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

      const eventList = await keepFactory.getPastEvents(
        "BondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      assert.equal(eventList.length, 1, "incorrect number of emitted events")

      const ev = eventList[0].returnValues

      assert.equal(ev.keepAddress, keepAddress, "incorrect keep address")
      assert.equal(ev.owner, keepOwner, "incorrect keep owner")
      assert.equal(ev.application, application, "incorrect application")

      assert.sameMembers(
        ev.members,
        [members[0], members[1], members[2]],
        "incorrect keep members"
      )

      expect(ev.honestThreshold).to.eq.BN(threshold, "incorrect threshold")
    })

    it("opens bonds for keep", async () => {
      const blockNumber = await web3.eth.getBlockNumber()

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

      const eventList = await keepFactory.getPastEvents(
        "BondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      const keepAddress = eventList[0].returnValues.keepAddress

      expect(
        await keepBonding.bondAmount(members[0], keepAddress, keepAddress)
      ).to.eq.BN(singleBond, "invalid bond value for members[0]")

      expect(
        await keepBonding.bondAmount(members[1], keepAddress, keepAddress)
      ).to.eq.BN(singleBond, "invalid bond value for members[1]")

      expect(
        await keepBonding.bondAmount(members[2], keepAddress, keepAddress)
      ).to.eq.BN(singleBond, "invalid bond value for members[2]")
    })

    it("rounds up members bonds", async () => {
      const requestedBond = bond.add(new BN(1))
      const unbondedAmount = singleBond.add(new BN(1))
      const expectedMemberBond = singleBond.add(new BN(1))

      await depositMemberCandidates(unbondedAmount)

      const blockNumber = await web3.eth.getBlockNumber()
      await keepFactory.openKeep(
        groupSize,
        threshold,
        keepOwner,
        requestedBond,
        stakeLockDuration,
        {from: application, value: feeEstimate}
      )

      const eventList = await keepFactory.getPastEvents(
        "BondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      const keepAddress = eventList[0].returnValues.keepAddress

      expect(
        await keepBonding.bondAmount(members[0], keepAddress, keepAddress),
        "invalid bond value for members[0]"
      ).to.eq.BN(expectedMemberBond)

      expect(
        await keepBonding.bondAmount(members[1], keepAddress, keepAddress),
        "invalid bond value for members[1]"
      ).to.eq.BN(expectedMemberBond)

      expect(
        await keepBonding.bondAmount(members[2], keepAddress, keepAddress),
        "invalid bond value for members[2]"
      ).to.eq.BN(expectedMemberBond)
    })

    it("rounds up members bonds when calculated bond per member equals zero", async () => {
      const requestedBond = new BN(groupSize).sub(new BN(1))
      const unbondedAmount = new BN(1)
      const expectedMemberBond = new BN(1)

      await depositMemberCandidates(unbondedAmount)

      const blockNumber = await web3.eth.getBlockNumber()
      await keepFactory.openKeep(
        groupSize,
        threshold,
        keepOwner,
        requestedBond,
        stakeLockDuration,
        {from: application, value: feeEstimate}
      )

      const eventList = await keepFactory.getPastEvents(
        "BondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      const keepAddress = eventList[0].returnValues.keepAddress

      expect(
        await keepBonding.bondAmount(members[0], keepAddress, keepAddress),
        "invalid bond value for members[0]"
      ).to.eq.BN(expectedMemberBond)

      expect(
        await keepBonding.bondAmount(members[1], keepAddress, keepAddress),
        "invalid bond value for members[1]"
      ).to.eq.BN(expectedMemberBond)

      expect(
        await keepBonding.bondAmount(members[2], keepAddress, keepAddress),
        "invalid bond value for members[2]"
      ).to.eq.BN(expectedMemberBond)
    })

    it("reverts if not enough member candidates are registered", async () => {
      const requestedGroupSize = groupSize.addn(1)

      await expectRevert(
        keepFactory.openKeep(
          requestedGroupSize,
          threshold,
          keepOwner,
          bond,
          stakeLockDuration,
          {
            from: application,
            value: feeEstimate,
          }
        ),
        "Not enough operators in pool"
      )
    })

    it("reverts if one member has insufficient unbonded value", async () => {
      const minimumBond = await keepFactory.minimumBond.call()
      const availableUnbonded = await keepBonding.availableUnbondedValue(
        members[2],
        keepFactory.address,
        signerPool
      )
      const withdrawValue = availableUnbonded.sub(minimumBond).add(new BN(1))
      await keepBonding.withdraw(withdrawValue, members[2], {
        from: members[2],
      })

      await expectRevert(
        keepFactory.openKeep(
          groupSize,
          threshold,
          keepOwner,
          bond,
          stakeLockDuration,
          {
            from: application,
            value: feeEstimate,
          }
        ),
        "Not enough operators in pool"
      )
    })

    it("opens keep with multiple members and emits an event", async () => {
      const blockNumber = await web3.eth.getBlockNumber()

      const keep = await openKeep()

      const eventList = await keepFactory.getPastEvents(
        "BondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      assert.isTrue(
        web3.utils.isAddress(keep.address),
        `keep address ${keep.address} is not a valid address`
      )

      assert.equal(eventList.length, 1, "incorrect number of emitted events")

      assert.equal(
        eventList[0].returnValues.keepAddress,
        keep.address,
        "incorrect keep address in emitted event"
      )

      assert.sameMembers(
        eventList[0].returnValues.members,
        [members[0], members[1], members[2]],
        "incorrect keep member in emitted event"
      )

      assert.equal(
        eventList[0].returnValues.owner,
        keepOwner,
        "incorrect keep owner in emitted event"
      )
    })

    it("requests new random group selection seed from random beacon", async () => {
      const expectedNewEntry = new BN(789)

      await randomBeacon.setEntry(expectedNewEntry)

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
        await randomBeacon.requestCount.call(),
        1,
        "incorrect number of beacon calls"
      )

      expect(await keepFactory.getGroupSelectionSeed()).to.eq.BN(
        expectedNewEntry,
        "incorrect new group selection seed"
      )
    })

    it("calculates new group selection seed", async () => {
      // Set entry to `0` so the beacon stub won't execute the callback.
      await randomBeacon.setEntry(0)

      const groupSelectionSeed = new BN(12)
      await keepFactory.initialGroupSelectionSeed(groupSelectionSeed)

      const expectedNewGroupSelectionSeed = web3.utils.toBN(
        web3.utils.soliditySha3(groupSelectionSeed, keepFactory.address)
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

      expect(await keepFactory.getGroupSelectionSeed()).to.eq.BN(
        expectedNewGroupSelectionSeed,
        "incorrect new group selection seed"
      )
    })

    it("ignores beacon request relay entry failure", async () => {
      await randomBeacon.setShouldFail(true)

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

      // TODO: Add verification of what we will do in case of the failure.
    })

    it("forwards payment to random beacon", async () => {
      const value = new BN(150)

      await keepFactory.openKeep(
        groupSize,
        threshold,
        keepOwner,
        bond,
        stakeLockDuration,
        {
          from: application,
          value: value,
        }
      )

      expect(await web3.eth.getBalance(randomBeacon.address)).to.eq.BN(
        value,
        "incorrect random beacon balance"
      )
    })

    it("reverts when honest threshold is greater than the group size", async () => {
      const honestThreshold = 4
      const groupSize = 3

      await expectRevert(
        keepFactory.openKeep(
          groupSize,
          honestThreshold,
          keepOwner,
          bond,
          stakeLockDuration,
          {
            from: application,
            value: feeEstimate,
          }
        ),
        "Honest threshold must be less or equal the group size"
      )
    })

    it("reverts when honest threshold is 0", async () => {
      const honestThreshold = 0

      await expectRevert(
        keepFactory.openKeep(
          groupSize,
          honestThreshold,
          keepOwner,
          bond,
          stakeLockDuration,
          {
            from: application,
            value: feeEstimate,
          }
        ),
        "Honest threshold must be greater than 0"
      )
    })

    it("works when honest threshold is equal to the group size", async () => {
      const honestThreshold = 3
      const groupSize = honestThreshold

      const blockNumber = await web3.eth.getBlockNumber()

      await keepFactory.openKeep(
        groupSize,
        honestThreshold,
        keepOwner,
        bond,
        stakeLockDuration,
        {
          from: application,
          value: feeEstimate,
        }
      )

      const eventList = await keepFactory.getPastEvents(
        "BondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      assert.equal(eventList.length, 1, "incorrect number of emitted events")
    })

    it("records the keep address and opening time", async () => {
      const preKeepCount = await keepFactory.getKeepCount()

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
      const recordedKeepAddress = await keepFactory.getKeepAtIndex(preKeepCount)
      const keep = await BondedECDSAKeep.at(keepAddress)
      const keepOpenedTime = await keep.getOpenedTimestamp()
      const factoryKeepOpenedTime = await keepFactory.getKeepOpenedTimestamp(
        keepAddress
      )

      assert.equal(
        recordedKeepAddress,
        keepAddress,
        "address recorded in factory differs from returned keep address"
      )

      expect(factoryKeepOpenedTime).to.eq.BN(
        keepOpenedTime,
        "opened time in factory differs from opened time in keep"
      )
    })

    it("produces active keeps", async () => {
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

      const keep = await BondedECDSAKeep.at(keepAddress)

      assert.isTrue(await keep.isActive(), "keep should be active")
    })

    it("allows to use a group of 16 signers", async () => {
      const groupSize = 16

      // create and authorize enough operators to perform the test;
      // we need more than the default 10 accounts
      await createDepositAndRegisterMembers(groupSize, singleBond)

      const blockNumber = await web3.eth.getBlockNumber()

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

      const eventList = await keepFactory.getPastEvents(
        "BondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      assert.equal(eventList.length, 1, "incorrect number of emitted events")
      assert.equal(
        eventList[0].returnValues.members.length,
        groupSize,
        "incorrect number of members"
      )
    })

    it("reverts when trying to use a group of 17 signers", async () => {
      const groupSize = 17

      await expectRevert(
        keepFactory.openKeep(
          groupSize,
          threshold,
          keepOwner,
          bond,
          stakeLockDuration,
          {
            from: application,
            value: feeEstimate,
          }
        ),
        "Maximum signing group size is 16"
      )
    })

    it("reverts when trying to use a group of 0 signers", async () => {
      const groupSize = 0

      await expectRevert(
        keepFactory.openKeep(
          groupSize,
          threshold,
          keepOwner,
          bond,
          stakeLockDuration,
          {
            from: application,
            value: feeEstimate,
          }
        ),
        "Minimum signing group size is 1"
      )
    })

    async function createDepositAndRegisterMembers(
      memberCount,
      unbondedAmount
    ) {
      const stakeBalance = await tokenStaking.minimumStake.call()

      for (let i = 0; i < memberCount; i++) {
        const operator = await web3.eth.personal.newAccount("pass")
        await web3.eth.personal.unlockAccount(operator, "pass", 5000) // 5 sec unlock

        web3.eth.sendTransaction({
          from: accounts[0],
          to: operator,
          value: web3.utils.toWei("1", "ether"),
        })

        await tokenStaking.setBalance(operator, stakeBalance)
        await tokenStaking.authorizeOperatorContract(
          operator,
          keepFactory.address
        )
        await keepBonding.authorizeSortitionPoolContract(operator, signerPool, {
          from: operator,
        })
        await keepBonding.deposit(operator, {value: unbondedAmount})
        await keepFactory.registerMemberCandidate(application, {
          from: operator,
        })
      }
    }
  })

  describe("__beaconCallback", async () => {
    const newRelayEntry = new BN(2345675)

    before(async () => {
      registry = await KeepRegistry.new()
      bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
      tokenStaking = await TokenStakingStub.new()
      tokenGrant = await TokenGrantStub.new()
      keepBonding = await KeepBonding.new(
        registry.address,
        tokenStaking.address,
        tokenGrant.address
      )
      randomBeacon = accounts[1]
      const bondedECDSAKeepMasterContract = await BondedECDSAKeep.new()
      keepFactory = await BondedECDSAKeepFactoryStub.new(
        bondedECDSAKeepMasterContract.address,
        bondedSortitionPoolFactory.address,
        tokenStaking.address,
        keepBonding.address,
        randomBeacon
      )
    })

    it("sets group selection seed", async () => {
      await keepFactory.__beaconCallback(newRelayEntry, {
        from: randomBeacon,
      })

      expect(await keepFactory.getGroupSelectionSeed()).to.eq.BN(
        newRelayEntry,
        "incorrect new group selection seed"
      )
    })

    it("reverts if called not by the random beacon", async () => {
      await expectRevert(
        keepFactory.__beaconCallback(newRelayEntry, {
          from: accounts[2],
        }),
        "Caller is not the random beacon"
      )
    })
  })

  describe("newGroupSelectionSeedFee", async () => {
    let newEntryFee

    before(async () => {
      await initializeNewFactory()

      const callbackGas = await keepFactory.callbackGas()
      newEntryFee = await randomBeacon.entryFeeEstimate(callbackGas)
    })

    it("evaluates reseed fee for empty pool", async () => {
      const reseedFee = await keepFactory.newGroupSelectionSeedFee()
      expect(reseedFee).to.eq.BN(
        newEntryFee,
        "reseed fee should equal new entry fee"
      )
    })

    it("evaluates reseed fee for non-empty pool", async () => {
      const poolValue = new BN(15)
      web3.eth.sendTransaction({
        from: accounts[0],
        to: keepFactory.address,
        value: poolValue,
      })

      const reseedFee = await keepFactory.newGroupSelectionSeedFee()
      expect(reseedFee).to.eq.BN(
        newEntryFee.sub(poolValue),
        "reseed fee should equal new entry fee minus pool value"
      )
    })

    it("should reseed for free if has enough funds in the pool", async () => {
      web3.eth.sendTransaction({
        from: accounts[0],
        to: keepFactory.address,
        value: newEntryFee,
      })

      const reseedFee = await keepFactory.newGroupSelectionSeedFee()
      expect(reseedFee).to.eq.BN(0, "reseed fee should be zero")
    })

    it("should reseed for free if has more than needed funds in the pool", async () => {
      web3.eth.sendTransaction({
        from: accounts[0],
        to: keepFactory.address,
        value: newEntryFee.addn(1),
      })

      const reseedFee = await keepFactory.newGroupSelectionSeedFee()
      expect(reseedFee).to.eq.BN(0, "reseed fee should be zero")
    })
  })

  describe("requestNewGroupSelectionSeed", async () => {
    let newEntryFee

    before(async () => {
      await initializeNewFactory()
      const callbackGas = await keepFactory.callbackGas()
      newEntryFee = await randomBeacon.entryFeeEstimate(callbackGas)
    })

    it("requests new relay entry from the beacon and reseeds factory", async () => {
      const expectedNewEntry = new BN(1337)
      await randomBeacon.setEntry(expectedNewEntry)

      const reseedFee = await keepFactory.newGroupSelectionSeedFee()
      await keepFactory.requestNewGroupSelectionSeed({value: reseedFee})

      assert.equal(
        await randomBeacon.requestCount.call(),
        1,
        "incorrect number of beacon calls"
      )

      expect(await keepFactory.getGroupSelectionSeed()).to.eq.BN(
        expectedNewEntry,
        "incorrect new group selection seed"
      )
    })

    it("allows to reseed for free if the pool is full", async () => {
      const expectedNewEntry = new BN(997)
      await randomBeacon.setEntry(expectedNewEntry)

      const poolValue = newEntryFee
      web3.eth.sendTransaction({
        from: accounts[0],
        to: keepFactory.address,
        value: poolValue,
      })

      await keepFactory.requestNewGroupSelectionSeed({value: 0})

      assert.equal(
        await randomBeacon.requestCount.call(),
        1,
        "incorrect number of beacon calls"
      )

      expect(await keepFactory.getGroupSelectionSeed()).to.eq.BN(
        expectedNewEntry,
        "incorrect new group selection seed"
      )
    })

    it("updates pool after reseeding", async () => {
      await randomBeacon.setEntry(new BN(1337))

      const poolValue = newEntryFee.muln(15)
      web3.eth.sendTransaction({
        from: accounts[0],
        to: keepFactory.address,
        value: poolValue,
      })

      await keepFactory.requestNewGroupSelectionSeed({value: 0})

      const expectedPoolValue = poolValue.sub(newEntryFee)
      expect(await keepFactory.reseedPool()).to.eq.BN(
        expectedPoolValue,
        "unexpected reseed pool value"
      )
    })

    it("updates pool after reseeding with value", async () => {
      await randomBeacon.setEntry(new BN(1337))

      const poolValue = newEntryFee.muln(15)
      web3.eth.sendTransaction({
        from: accounts[0],
        to: keepFactory.address,
        value: poolValue,
      })

      const valueSent = new BN(10)
      await keepFactory.requestNewGroupSelectionSeed({value: 10})

      const expectedPoolValue = poolValue.sub(newEntryFee).add(valueSent)
      expect(await keepFactory.reseedPool()).to.eq.BN(
        expectedPoolValue,
        "unexpected reseed pool value"
      )
    })

    it("reverts if the provided payment is not sufficient", async () => {
      const poolValue = newEntryFee.subn(2)
      web3.eth.sendTransaction({
        from: accounts[0],
        to: keepFactory.address,
        value: poolValue,
      })

      await expectRevert(
        keepFactory.requestNewGroupSelectionSeed({value: 1}),
        "Not enough funds to trigger reseed"
      )
    })

    it("reverts if beacon is busy", async () => {
      await randomBeacon.setShouldFail(true)

      const reseedFee = await keepFactory.newGroupSelectionSeedFee()
      await expectRevert(
        keepFactory.requestNewGroupSelectionSeed({value: reseedFee}),
        "request relay entry failed"
      )
    })
  })

  describe("getKeepAtIndex", async () => {
    before(async () => {
      await initializeNewFactory()
      const minimumBond = await keepFactory.minimumBond.call()
      const memberBond = minimumBond.muln(2) // want to be able to open 2 keeps
      await initializeMemberCandidates(memberBond)
      await registerMemberCandidates()
    })

    it("reverts when there are no keeps", async () => {
      await expectRevert(keepFactory.getKeepAtIndex(0), "Out of bounds")
    })

    it("reverts for out of bond index", async () => {
      await openKeep()

      await expectRevert(keepFactory.getKeepAtIndex(1), "Out of bounds")
    })

    it("returns keep at index", async () => {
      const keep0 = await openKeep()
      const keep1 = await openKeep()

      const atIndex0 = await keepFactory.getKeepAtIndex(0)
      const atIndex1 = await keepFactory.getKeepAtIndex(1)

      assert.equal(
        keep0.address,
        atIndex0,
        "incorrect keep address returned for index 0"
      )
      assert.equal(
        keep1.address,
        atIndex1,
        "incorrect keep address returned for index 1"
      )
    })
  })

  describe("isOperatorAuthorized", async () => {
    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()
    })

    it("returns true if operator is authorized for the factory", async () => {
      assert.isTrue(
        await keepFactory.isOperatorAuthorized(members[0]),
        "the operator is authorized for the factory"
      )
    })

    it("returns false if operator is not authorized for the factory", async () => {
      assert.isFalse(
        await keepFactory.isOperatorAuthorized(notMember),
        "the operator is not authorized for the factory"
      )
    })
  })

  describe("getSortitionPoolWeight", async () => {
    before(async () => {
      await initializeNewFactory()
    })

    it("returns pool weight if pool exists for application", async () => {
      await initializeMemberCandidates()
      await registerMemberCandidates()

      const poolWeight = await keepFactory.getSortitionPoolWeight(application)

      const expectedPoolWeight = new BN(600000)
      expect(poolWeight).to.eq.BN(
        expectedPoolWeight,
        "incorrect sortition pool weight"
      )
    })

    it("reverts when pool doesn't exist for application", async () => {
      await expectRevert(
        keepFactory.getSortitionPoolWeight(application),
        "No pool found for the application"
      )
    })
  })

  async function initializeNewFactory() {
    registry = await KeepRegistry.new()
    bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
    tokenStaking = await TokenStakingStub.new()
    tokenGrant = await TokenGrantStub.new()
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

    minimumStake = await tokenStaking.minimumStake.call()
    await stakeOperators(members, minimumStake)
  }

  async function stakeOperators(members, stakeBalance) {
    for (let i = 0; i < members.length; i++) {
      await tokenStaking.setBalance(members[i], stakeBalance)
    }
  }

  async function initializeMemberCandidates(unbondedValue) {
    const minimumBond = await keepFactory.minimumBond.call()

    signerPool = await keepFactory.createSortitionPool.call(application)
    await keepFactory.createSortitionPool(application)

    for (let i = 0; i < members.length; i++) {
      await tokenStaking.authorizeOperatorContract(
        members[i],
        keepFactory.address
      )
      await keepBonding.authorizeSortitionPoolContract(members[i], signerPool, {
        from: authorizers[i],
      })
    }

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
    const feeEstimate = await keepFactory.openKeepFeeEstimate()

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
})
