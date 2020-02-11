import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const { expectRevert } = require('openzeppelin-test-helpers');

const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactoryExposed');
const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');
const TokenStakingStub = artifacts.require("TokenStakingStub")
const KeepBondingStub = artifacts.require('KeepBondingStub');
const BondedSortitionPool = artifacts.require('BondedSortitionPool');
const BondedSortitionPoolFactory = artifacts.require('BondedSortitionPoolFactory');
const RandomBeaconStub = artifacts.require('RandomBeaconStub')

const BN = web3.utils.BN

const chai = require('chai')
chai.use(require('bn-chai')(BN))
const expect = chai.expect

contract("ECDSAKeepFactory", async accounts => {
    let keepFactory
    let bondedSortitionPoolFactory
    let tokenStaking
    let keepBonding
    let randomBeacon

    const application = accounts[1]
    const member1 = accounts[2]
    const member2 = accounts[3]
    const member3 = accounts[4]

    describe("registerMemberCandidate", async () => {
        before(async () => {
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            tokenStaking = await TokenStakingStub.new()
            keepBonding = await KeepBondingStub.new()
            randomBeacon = await RandomBeaconStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon.address
            )
        })

        beforeEach(async () => {
            await createSnapshot()
        })

        afterEach(async () => {
            await restoreSnapshot()
        })

        it("creates a signer pool", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })

            const signerPoolAddress = await keepFactory.getSignerPool(application)

            assert.notEqual(
                signerPoolAddress,
                "0x0000000000000000000000000000000000000000",
                "incorrect registered signer pool",
            )
        })

        it("inserts operator with the correct staking weight in the pool", async () => {
            const stakingWeight = new BN(199)
            await tokenStaking.setBalance(stakingWeight)

            await keepFactory.registerMemberCandidate(application, { from: member1 })

            const signerPoolAddress = await keepFactory.getSignerPool(application)
            const signerPool = await BondedSortitionPool.at(signerPoolAddress)

            // TODO: Update result verification when sortition-pools interfaces
            // and implementation are ready.
            // expect(
            //     await signerPool.getPoolWeight.call(member1)
            // ).to.eq.BN(stakingWeight, 'invalid staking weight')
        })

        it("inserts operators to the same pool", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })

            const signerPoolAddress = await keepFactory.getSignerPool(application)

            const signerPool = await BondedSortitionPool.at(signerPoolAddress)

            assert.isTrue(await signerPool.isOperatorInPool(member1), "operator 1 is not in the pool")
            assert.isTrue(await signerPool.isOperatorInPool(member2), "operator 2 is not in the pool")
        })

        it("does not add an operator to the pool if he is already there", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })
            const signerPoolAddress = await keepFactory.getSignerPool(application)

            const signerPool = await BondedSortitionPool.at(signerPoolAddress)

            assert.isTrue(await signerPool.isOperatorInPool(member1), "operator is not in the pool")

            await keepFactory.registerMemberCandidate(application, { from: member1 })

            assert.isTrue(await signerPool.isOperatorInPool(member1), "operator is not in the pool")
        })

        it("inserts operators to different pools", async () => {
            const application1 = '0x0000000000000000000000000000000000000001'
            const application2 = '0x0000000000000000000000000000000000000002'

            await keepFactory.registerMemberCandidate(application1, { from: member1 })
            await keepFactory.registerMemberCandidate(application2, { from: member2 })

            const signerPool1Address = await keepFactory.getSignerPool(application1)
            const signerPool1 = await BondedSortitionPool.at(signerPool1Address)

            assert.isTrue(await signerPool1.isOperatorInPool(member1), "operator 1 is not in the pool")
            assert.isFalse(await signerPool1.isOperatorInPool(member2), "operator 2 is in the pool")

            const signerPool2Address = await keepFactory.getSignerPool(application2)
            const signerPool2 = await BondedSortitionPool.at(signerPool2Address)

            assert.isFalse(await signerPool2.isOperatorInPool(member1), "operator 1 is in the pool")
            assert.isTrue(await signerPool2.isOperatorInPool(member2), "operator 2 is not in the pool")
        })
    })

    describe("registerMemberCandidate", async () => {
        before(async () => {
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            tokenStaking = await TokenStakingStub.new()
            keepBonding = await KeepBondingStub.new()
            randomBeacon = await RandomBeaconStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon.address
            )
        })

        beforeEach(async () => {
            await createSnapshot()
        })

        afterEach(async () => {
            await restoreSnapshot()
        })

        it("inserts operators to different pools", async () => {
            const application1 = '0x0000000000000000000000000000000000000001'
            const application2 = '0x0000000000000000000000000000000000000002'

            await keepFactory.registerMemberCandidate(application1, { from: member1 })
            await keepFactory.registerMemberCandidate(application2, { from: member2 })

            const signerPool1Address = await keepFactory.getSignerPool(application1)
            const signerPool1 = await BondedSortitionPool.at(signerPool1Address)

            assert.isTrue(await signerPool1.isOperatorInPool(member1), "operator 1 is not in the pool")
            assert.isFalse(await signerPool1.isOperatorInPool(member2), "operator 2 is in the pool")

            const signerPool2Address = await keepFactory.getSignerPool(application2)
            const signerPool2 = await BondedSortitionPool.at(signerPool2Address)

            assert.isFalse(await signerPool2.isOperatorInPool(member1), "operator 1 is in the pool")
            assert.isTrue(await signerPool2.isOperatorInPool(member2), "operator 2 is not in the pool")
        })
    })

    describe("openKeep", async () => {
        const keepOwner = "0xbc4862697a1099074168d54A555c4A60169c18BD"
        const groupSize = new BN(3)
        const threshold = new BN(3)

        const singleBond = new BN(1)
        const bond = singleBond.mul(groupSize)

        let feeEstimate

        async function initializeNewFactory() {
            // Tests are executed with real implementation of sortition pools.
            // We don't use stub to ensure that keep members selection works correctly.
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            tokenStaking = await TokenStakingStub.new()
            keepBonding = await KeepBondingStub.new()
            randomBeacon = await RandomBeaconStub.new()
            keepFactory = await ECDSAKeepFactory.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon.address
            )

            feeEstimate = await keepFactory.openKeepFeeEstimate()
        }

        beforeEach(async () => {
            await initializeNewFactory()

            await keepBonding.deposit(member1, { value: singleBond })
            await keepBonding.deposit(member2, { value: singleBond })
            await keepBonding.deposit(member3, { value: singleBond })

            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })
            await keepFactory.registerMemberCandidate(application, { from: member3 })

            await createSnapshot()
        })

        afterEach(async () => {
            await restoreSnapshot()
        })

        it("reverts if no member candidates are registered", async () => {
            try {
                await keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { value: feeEstimate }
                )

                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, "No signer pool for this application")
            }
        })

        it("reverts if bond equals zero", async () => {
            let bond = 0

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application, value: feeEstimate },
                ),
                "Bond per member must be greater than zero"
            )
        })

        it("reverts if bond per member equals zero", async () => {
            let bond = new BN(2)

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application, value: feeEstimate },
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
                    { from: application, fee: insufficientFee },
                ),
                "Insufficient payment for opening a new keep"
            )
        })

        it("opens keep with multiple members", async () => {
            let blockNumber = await web3.eth.getBlockNumber()

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application, value: feeEstimate },
            )

            let eventList = await keepFactory.getPastEvents('ECDSAKeepCreated', {
                fromBlock: blockNumber,
                toBlock: 'latest'
            })

            assert.equal(eventList.length, 1, "incorrect number of emitted events")

            assert.sameMembers(
                eventList[0].returnValues.members,
                [member1, member2, member3],
                "incorrect keep member in emitted event",
            )
        })

        it("opens bonds for keep", async () => {
            let blockNumber = await web3.eth.getBlockNumber()

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application, value: feeEstimate },
            )

            let eventList = await keepFactory.getPastEvents('ECDSAKeepCreated', {
                fromBlock: blockNumber,
                toBlock: 'latest'
            })

            const keepAddress = eventList[0].returnValues.keepAddress

            expect(
                await keepBonding.getLockedBonds(keepAddress, member1, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member1')

            expect(
                await keepBonding.getLockedBonds(keepAddress, member2, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member2')

            expect(
                await keepBonding.getLockedBonds(keepAddress, member3, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member3')
        })

        // This test checks that if the requested bond value divided by the group
        // size has a remainder which is not bonded, e.g.:
        // requested bond = 11 & group size = 3 => bond per member ≈ 3,66
        // but `11.div(3) = 3` so in current implementation we bond only 9 and
        // the rest remains unbonded.
        // TODO: Check if such case is acceptable.
        it("forgets about the remainder", async () => {
            await initializeNewFactory()

            const groupSize = 3
            const singleBond = new BN(3)
            const bond = new BN(11)

            await keepBonding.deposit(member1, { value: singleBond })
            await keepBonding.deposit(member2, { value: singleBond })
            await keepBonding.deposit(member3, { value: singleBond })

            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })
            await keepFactory.registerMemberCandidate(application, { from: member3 })

            let blockNumber = await web3.eth.getBlockNumber()

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application, value: feeEstimate },
            )

            let eventList = await keepFactory.getPastEvents('ECDSAKeepCreated', {
                fromBlock: blockNumber,
                toBlock: 'latest'
            })

            const keepAddress = eventList[0].returnValues.keepAddress

            expect(
                await keepBonding.getLockedBonds(keepAddress, member1, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member1')

            expect(
                await keepBonding.getLockedBonds(keepAddress, member2, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member2')

            expect(
                await keepBonding.getLockedBonds(keepAddress, member3, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member3')
        })

        it("reverts if not enough member candidates are registered", async () => {
            await initializeNewFactory()

            let groupSize = 2
            let threshold = 2

            await keepBonding.deposit(member1, { value: singleBond })

            await keepFactory.registerMemberCandidate(application, { from: member1 })

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application, value: feeEstimate }
                ),
                "Not enough operators in pool"
            )
        })

        // TODO: This is temporary, we don't expect a group to be formed if a member
        // doesn't have sufficient unbonded value.
        it("reverts if one member has insufficient unbonded value", async () => {
            await initializeNewFactory()

            await keepBonding.deposit(member1, { value: singleBond })
            await keepBonding.deposit(member2, { value: singleBond })
            await keepBonding.deposit(member3, { value: singleBond.sub(new BN(1)) })

            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })
            await keepFactory.registerMemberCandidate(application, { from: member3 })

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application, value: feeEstimate }
                ),
                "Insufficient unbonded value"
            )
        })

        it("opens keep with multiple members and emits an event", async () => {
            let blockNumber = await web3.eth.getBlockNumber()

            let keepAddress = await keepFactory.openKeep.call(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application, value: feeEstimate }
            )

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application, value: feeEstimate }
            )

            let eventList = await keepFactory.getPastEvents('ECDSAKeepCreated', {
                fromBlock: blockNumber,
                toBlock: 'latest'
            })

            assert.isTrue(
                web3.utils.isAddress(keepAddress),
                `keep address ${keepAddress} is not a valid address`,
            );

            assert.equal(eventList.length, 1, "incorrect number of emitted events")

            assert.equal(
                eventList[0].returnValues.keepAddress,
                keepAddress,
                "incorrect keep address in emitted event",
            )

            assert.sameMembers(
                eventList[0].returnValues.members,
                [member1, member2, member3],
                "incorrect keep member in emitted event",
            )

            assert.equal(
                eventList[0].returnValues.owner,
                keepOwner,
                "incorrect keep owner in emitted event",
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
                { from: application, value: feeEstimate }
            )

            assert.equal(
                await randomBeacon.requestCount.call(),
                1,
                "incorrect number of beacon calls",
            )

            expect(
                await keepFactory.getGroupSelectionSeed()
            ).to.eq.BN(expectedNewEntry, "incorrect new group selection seed")
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
                { from: application, value: feeEstimate }
            )

            expect(
                await keepFactory.getGroupSelectionSeed()
            ).to.eq.BN(
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
                { from: application, value: feeEstimate }
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
                { from: application, value: value }
            )

            expect(
                await web3.eth.getBalance(randomBeacon.address)
            ).to.eq.BN(
                value,
                "incorrect random beacon balance"
            )
        })
    })

    describe("setGroupSelectionSeed", async () => {
        const newGroupSelectionSeed = new BN(2345675)

        before(async () => {
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            tokenStaking = await TokenStakingStub.new()
            keepBonding = await KeepBondingStub.new()
            randomBeacon = accounts[1]
            keepFactory = await ECDSAKeepFactory.new(bondedSortitionPoolFactory.address, tokenStaking.address, keepBonding.address, randomBeacon)
        })

        beforeEach(async () => {
            await createSnapshot()
        })

        afterEach(async () => {
            await restoreSnapshot()
        })

        it("sets group selection seed", async () => {
            await keepFactory.setGroupSelectionSeed(newGroupSelectionSeed, { from: randomBeacon })

            expect(
                await keepFactory.getGroupSelectionSeed()
            ).to.eq.BN(
                newGroupSelectionSeed,
                "incorrect new group selection seed"
            )
        })

        it("reverts if called not by the random beacon", async () => {
            await expectRevert(
                keepFactory.setGroupSelectionSeed(newGroupSelectionSeed, { from: accounts[2] }),
                "Caller is not the random beacon"
            )
        })
    })
})
