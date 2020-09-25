const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("./helpers/snapshot")

const {expectRevert, time} = require("@openzeppelin/test-helpers")

const {mineBlocks} = require("./helpers/mineBlocks")

const truffleAssert = require("truffle-assertions")

const StackLib = contract.fromArtifact("StackLib")
const KeepRegistry = contract.fromArtifact("KeepRegistry")
const FullyBackedBondedECDSAKeepFactoryStub = contract.fromArtifact(
  "FullyBackedBondedECDSAKeepFactoryStub"
)
const FullyBackedBonding = contract.fromArtifact("FullyBackedBonding")
const FullyBackedSortitionPool = contract.fromArtifact(
  "FullyBackedSortitionPool"
)
const FullyBackedSortitionPoolFactory = contract.fromArtifact(
  "FullyBackedSortitionPoolFactory"
)
const RandomBeaconStub = contract.fromArtifact("RandomBeaconStub")
const FullyBackedBondedECDSAKeep = contract.fromArtifact(
  "FullyBackedBondedECDSAKeep"
)

const BN = web3.utils.BN

const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect
const assert = chai.assert

describe("FullyBackedBondedECDSAKeepFactory", function () {
  let registry
  let keepFactory
  let sortitionPoolFactory
  let bonding
  let randomBeacon
  let signerPool
  let minimumDelegationDeposit

  const application = accounts[1]
  const members = [accounts[2], accounts[3], accounts[4]]
  const authorizers = [members[0], members[1], members[2]]
  const notMember = accounts[5]

  const keepOwner = accounts[6]
  const beneficiary = accounts[7]

  const groupSize = new BN(members.length)
  const threshold = new BN(groupSize - 1)

  const singleBond = web3.utils.toWei(new BN(20))
  const bond = singleBond.mul(groupSize)

  const stakeLockDuration = 0 // parameter is ignored by FullyBackedBondedECDSAKeepFactory implementation
  const delegationInitPeriod = time.duration.hours(12)

  before(async () => {
    await FullyBackedSortitionPoolFactory.detectNetwork()
    await FullyBackedSortitionPoolFactory.link(
      "StackLib",
      (await StackLib.new()).address
    )
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("registerMemberCandidate", async () => {
    let minimumBondableValue
    let bondWeightDivisor
    let pool

    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()

      pool = await FullyBackedSortitionPool.at(signerPool)
      minimumBondableValue = await pool.getMinimumBondableValue()

      bondWeightDivisor = await keepFactory.bondWeightDivisor.call()
    })

    it("reverts for unknown application", async () => {
      const unknownApplication = "0xCfd27747D1583feb1eCbD7c4e66C848Db0aA82FB"
      await expectRevert(
        keepFactory.registerMemberCandidate(unknownApplication, {
          from: members[0],
        }),
        "No pool found for the application"
      )
    })

    it("inserts operator with the correct unbonded value available", async () => {
      const unbondedValue = await bonding.unbondedValue(members[0])

      const pool = await FullyBackedSortitionPool.at(signerPool)

      const expectedWeight = unbondedValue.div(bondWeightDivisor)

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      expect(await pool.getPoolWeight(members[0])).to.eq.BN(
        expectedWeight,
        "invalid staking weight"
      )
    })

    it("inserts operators to the same pool", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })
      await keepFactory.registerMemberCandidate(application, {
        from: members[1],
      })

      const pool = await FullyBackedSortitionPool.at(signerPool)
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

      const pool = await FullyBackedSortitionPool.at(signerPool)

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

    it("does not add an operator to the pool if it does not have a minimum bond", async () => {
      await setUnbondedValue(members[0], minimumBondableValue.sub(new BN(1)))

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

      await bonding.authorizeSortitionPoolContract(
        members[0],
        signerPool1Address,
        {from: authorizers[0]}
      )
      await bonding.authorizeSortitionPoolContract(
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

      const signerPool1 = await FullyBackedSortitionPool.at(signerPool1Address)

      assert.isTrue(
        await signerPool1.isOperatorInPool(members[0]),
        "operator 1 is not in the pool"
      )
      assert.isFalse(
        await signerPool1.isOperatorInPool(members[1]),
        "operator 2 is in the pool"
      )

      const signerPool2 = await FullyBackedSortitionPool.at(signerPool2Address)

      assert.isFalse(
        await signerPool2.isOperatorInPool(members[0]),
        "operator 1 is in the pool"
      )
      assert.isTrue(
        await signerPool2.isOperatorInPool(members[1]),
        "operator 2 is not in the pool"
      )
    })

    it("reverts if delegation initialization period has not passed", async () => {
      await initializeNewFactory()

      signerPool = await keepFactory.createSortitionPool.call(application)
      await keepFactory.createSortitionPool(application)
      const pool = await FullyBackedSortitionPool.at(signerPool)

      await delegate(
        members[0],
        members[0],
        authorizers[0],
        minimumBondableValue
      )

      await time.increase(delegationInitPeriod.subn(1))

      await expectRevert(
        keepFactory.registerMemberCandidate(application, {
          from: members[0],
        }),
        "Operator not eligible"
      )

      await time.increase(2)

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      assert.isTrue(
        await pool.isOperatorInPool(members[0]),
        "operator is not in the pool after initialization period"
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
    const precision = new BN("1000000000000000000") // 1 ETH

    let minimumBondableValue

    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()

      const pool = await FullyBackedSortitionPool.at(signerPool)
      minimumBondableValue = await pool.getMinimumBondableValue()
    })

    // TODO: Add a case to check if operator is initialized
    // github.com/keep-network/sortition-pools/blob/413abd66536032ee757f190a29fda2705c7b70cf/contracts/FullyBackedSortitionPool.sol#L155

    it("returns true if the operator is up to date for the application", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      assert.isTrue(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns false if the operator bonding value dropped", async () => {
      const unbondedValue = minimumBondableValue.muln(10) // (20 * 10^18) * 10
      await setUnbondedValue(members[0], unbondedValue)

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await setUnbondedValue(members[0], unbondedValue.subn(1)) // 200 - 1 = 199

      // Precision (bond weight divisor) is 10^18
      // (20 * 10^18 * 10) / 10^18       =  200
      // ((20 * 10^18 * 10) - 1) / 10^18 =~ 199.99
      // Ethereum uint256 division performs implicit floor:
      // The weight went down from 200 to 199
      assert.isFalse(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns false if the operator bonding value increased", async () => {
      const unbondedValue = minimumBondableValue.muln(10) // (20 * 10^18) * 10
      await setUnbondedValue(members[0], unbondedValue)

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await setUnbondedValue(members[0], unbondedValue.add(precision)) // ((20 * 10^18) * 10) + 10^18

      // Precision (bond weight divisor) is 10^18
      // (20 * 10^18 * 10) / 10^18           = 200
      // ((20 * 10^18 * 10) + 10^18) / 10^18 = 201
      // The weight went up from 200 to 201
      assert.isFalse(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns true if the operator bonding value increase is below weight precision", async () => {
      const unbondedValue = minimumBondableValue.muln(10) // (20 * 10^18) * 10
      await setUnbondedValue(members[0], unbondedValue)

      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await setUnbondedValue(members[0], unbondedValue.add(precision).subn(1)) // ((20 * 10^18) * 10) + 10^18 - 1

      // Precision (pool weight divisor) is 10^18
      // (20 * 10^18 * 10) / 10^18                 =  200
      // ((20 * 10^18 * 10) + (10^18 - 1)) / 10^18 =~ 200.99
      // Ethereum uint256 division performs implicit floor:
      // The weight dit not change: 200 == floor(200.99)
      assert.isTrue(
        await keepFactory.isOperatorUpToDate(members[0], application)
      )
    })

    it("returns false if the operator stake dropped below minimum", async () => {
      await keepFactory.registerMemberCandidate(application, {
        from: members[0],
      })

      await setUnbondedValue(members[0], minimumBondableValue.subn(1))

      assert.isFalse(
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

    it("removes operator if bonding value has changed below minimum", async () => {
      bonding.withdraw(new BN(1), members[0], {from: members[0]})
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
      bonding.deposit(members[0], {value: new BN(1)})
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
    let minimumBondableValue

    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()
      await registerMemberCandidates()

      feeEstimate = await keepFactory.openKeepFeeEstimate()

      const pool = await FullyBackedSortitionPool.at(signerPool)
      minimumBondableValue = await pool.getMinimumBondableValue()
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
        "No pool found for the application"
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
      const insufficientFee = feeEstimate.subn(1)

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
        "FullyBackedBondedECDSAKeepCreated",
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
        "FullyBackedBondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      const keepAddress = eventList[0].returnValues.keepAddress

      expect(
        await bonding.bondAmount(members[0], keepAddress, keepAddress)
      ).to.eq.BN(singleBond, "invalid bond value for members[0]")

      expect(
        await bonding.bondAmount(members[1], keepAddress, keepAddress)
      ).to.eq.BN(singleBond, "invalid bond value for members[1]")

      expect(
        await bonding.bondAmount(members[2], keepAddress, keepAddress)
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
        "FullyBackedBondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      const keepAddress = eventList[0].returnValues.keepAddress

      expect(
        await bonding.bondAmount(members[0], keepAddress, keepAddress),
        "invalid bond value for members[0]"
      ).to.eq.BN(expectedMemberBond)

      expect(
        await bonding.bondAmount(members[1], keepAddress, keepAddress),
        "invalid bond value for members[1]"
      ).to.eq.BN(expectedMemberBond)

      expect(
        await bonding.bondAmount(members[2], keepAddress, keepAddress),
        "invalid bond value for members[2]"
      ).to.eq.BN(expectedMemberBond)
    })

    it("rounds up members bonds when calculated bond per member equals zero", async () => {
      const requestedBond = new BN(groupSize).sub(new BN(1))
      const expectedMemberBond = new BN(1)

      await depositMemberCandidates(minimumBondableValue)

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
        "FullyBackedBondedECDSAKeepCreated",
        {
          fromBlock: blockNumber,
          toBlock: "latest",
        }
      )

      const keepAddress = eventList[0].returnValues.keepAddress

      expect(
        await bonding.bondAmount(members[0], keepAddress, keepAddress),
        "invalid bond value for members[0]"
      ).to.eq.BN(expectedMemberBond)

      expect(
        await bonding.bondAmount(members[1], keepAddress, keepAddress),
        "invalid bond value for members[1]"
      ).to.eq.BN(expectedMemberBond)

      expect(
        await bonding.bondAmount(members[2], keepAddress, keepAddress),
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
      const availableUnbonded = await bonding.availableUnbondedValue(
        members[2],
        keepFactory.address,
        signerPool
      )
      const withdrawValue = availableUnbonded.sub(minimumBond).add(new BN(1))
      await bonding.withdraw(withdrawValue, members[2], {
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
        "FullyBackedBondedECDSAKeepCreated",
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
        "FullyBackedBondedECDSAKeepCreated",
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
      const keep = await FullyBackedBondedECDSAKeep.at(keepAddress)
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

      const keep = await FullyBackedBondedECDSAKeep.at(keepAddress)

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
        "FullyBackedBondedECDSAKeepCreated",
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
      const operators = []

      for (let i = 0; i < memberCount; i++) {
        const operator = await web3.eth.personal.newAccount("pass")
        const authorizer = operator
        await web3.eth.personal.unlockAccount(operator, "pass", 5000) // 5 sec unlock

        web3.eth.sendTransaction({
          from: accounts[0],
          to: operator,
          value: web3.utils.toWei("21", "ether"),
        })

        await delegate(operator, beneficiary, authorizer, unbondedAmount)

        operators[i] = operator
      }

      await time.increase(delegationInitPeriod.addn(1))

      for (let i = 0; i < operators.length; i++) {
        await keepFactory.registerMemberCandidate(application, {
          from: operators[i],
        })
      }

      const pool = await FullyBackedSortitionPool.at(signerPool)
      await mineBlocks((await pool.operatorInitBlocks()).add(new BN(1)))
    }
  })

  describe("__beaconCallback", async () => {
    const newRelayEntry = new BN(2345675)

    before(async () => {
      registry = await KeepRegistry.new()
      sortitionPoolFactory = await FullyBackedSortitionPoolFactory.new()
      bonding = await FullyBackedBonding.new(
        registry.address,
        delegationInitPeriod
      )
      randomBeacon = accounts[1]
      const keepMasterContract = await FullyBackedBondedECDSAKeep.new()
      keepFactory = await FullyBackedBondedECDSAKeepFactoryStub.new(
        keepMasterContract.address,
        sortitionPoolFactory.address,
        bonding.address,
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

      const poolAddress = await keepFactory.getSortitionPool(application)
      const pool = await FullyBackedSortitionPool.at(poolAddress)

      const poolWeight = await keepFactory.getSortitionPoolWeight(application)

      const expectedPoolWeight = (await pool.getMinimumBondableValue())
        .div(await keepFactory.bondWeightDivisor.call())
        .muln(3)
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

  describe("setMinimumBondableValue", async () => {
    before(async () => {
      await initializeNewFactory()
      await initializeMemberCandidates()
    })

    it("reverts for unknown application", async () => {
      await expectRevert(
        keepFactory.setMinimumBondableValue(10, 3, 3),
        "No pool found for the application"
      )
    })

    it("sets the minimum bond value for the application", async () => {
      await keepFactory.setMinimumBondableValue(12, 3, 3, {from: application})
      const poolAddress = await keepFactory.getSortitionPool(application)
      const pool = await FullyBackedSortitionPool.at(poolAddress)
      expect(await pool.getMinimumBondableValue()).to.eq.BN(4)
    })

    it("rounds up member bonds", async () => {
      await keepFactory.setMinimumBondableValue(10, 3, 3, {from: application})
      const poolAddress = await keepFactory.getSortitionPool(application)
      const pool = await FullyBackedSortitionPool.at(poolAddress)
      expect(await pool.getMinimumBondableValue()).to.eq.BN(4)
    })

    it("rounds up members bonds when calculated bond per member equals zero", async () => {
      await keepFactory.setMinimumBondableValue(2, 3, 3, {from: application})
      const poolAddress = await keepFactory.getSortitionPool(application)
      const pool = await FullyBackedSortitionPool.at(poolAddress)
      expect(await pool.getMinimumBondableValue()).to.eq.BN(1)
    })
  })

  async function initializeNewFactory() {
    registry = await KeepRegistry.new()
    sortitionPoolFactory = await FullyBackedSortitionPoolFactory.new()
    bonding = await FullyBackedBonding.new(
      registry.address,
      delegationInitPeriod
    )
    randomBeacon = await RandomBeaconStub.new()
    const keepMasterContract = await FullyBackedBondedECDSAKeep.new()
    keepFactory = await FullyBackedBondedECDSAKeepFactoryStub.new(
      keepMasterContract.address,
      sortitionPoolFactory.address,
      bonding.address,
      randomBeacon.address
    )

    minimumDelegationDeposit = await bonding.MINIMUM_DELEGATION_DEPOSIT.call()

    await registry.approveOperatorContract(keepFactory.address)
  }

  async function initializeMemberCandidates(unbondedValue) {
    const minimumBond = await keepFactory.minimumBond.call()

    signerPool = await keepFactory.createSortitionPool.call(application)
    await keepFactory.createSortitionPool(application)

    for (let i = 0; i < members.length; i++) {
      await delegate(
        members[i],
        members[i],
        authorizers[i],
        unbondedValue || minimumBond
      )
    }

    await time.increase(delegationInitPeriod.addn(1))
  }

  async function delegate(operator, beneficiary, authorizer, unbondedValue) {
    await bonding.delegate(operator, beneficiary, authorizer, {
      value: unbondedValue || minimumDelegationDeposit,
    })

    await bonding.authorizeOperatorContract(operator, keepFactory.address, {
      from: authorizer,
    })
    await bonding.authorizeSortitionPoolContract(operator, signerPool, {
      from: authorizer,
    })
  }

  async function setUnbondedValue(operator, unbondedValue) {
    const initialUnbondedValue = await bonding.unbondedValue(operator)

    if (initialUnbondedValue.eq(unbondedValue)) {
      return
    } else if (initialUnbondedValue.gt(unbondedValue)) {
      await bonding.withdraw(initialUnbondedValue.sub(unbondedValue), operator)
    } else {
      await bonding.deposit(operator, {
        value: unbondedValue.sub(initialUnbondedValue),
      })
    }
  }

  async function depositMemberCandidates(unbondedValue) {
    for (let i = 0; i < members.length; i++) {
      await setUnbondedValue(members[i], unbondedValue)
    }
  }

  async function registerMemberCandidates() {
    for (let i = 0; i < members.length; i++) {
      await keepFactory.registerMemberCandidate(application, {
        from: members[i],
      })
    }

    const pool = await FullyBackedSortitionPool.at(signerPool)
    const poolInitBlocks = await pool.operatorInitBlocks()
    await mineBlocks(poolInitBlocks.add(new BN(1)))
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

    return await FullyBackedBondedECDSAKeep.at(keepAddress)
  }
})
