import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const { expectRevert } = require('openzeppelin-test-helpers');

const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory');
const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');
const KeepBondingStub = artifacts.require('KeepBondingStub');
const SortitionPoolFactoryStub = artifacts.require('SortitionPoolFactoryStub');
const SortitionPoolStub = artifacts.require('SortitionPoolStub');
const SortitionPoolFactory = artifacts.require('SortitionPoolFactory');

const BN = web3.utils.BN

const chai = require('chai')
chai.use(require('bn-chai')(BN))
const expect = chai.expect

contract("ECDSAKeepFactory", async accounts => {
    let keepFactory
    let sortitionPoolFactory
    let keepBonding

    const application = accounts[1]
    const member1 = accounts[2]
    const member2 = accounts[3]
    const member3 = accounts[4]

    describe("registerMemberCandidate", async () => {
        before(async () => {
            sortitionPoolFactory = await SortitionPoolFactoryStub.new()
            keepBonding = await KeepBondingStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(sortitionPoolFactory.address, keepBonding.address)
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

        it("registers transferred value in bonding contract", async () => {
            const value = new BN(100)

            await keepFactory.registerMemberCandidate(application, { from: member1, value: value })

            expect(
                await keepBonding.availableBondingValue(member1)
            ).to.eq.BN(value, 'invalid available bonding value')
        })

        it("inserts operators to the same pool", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })

            const signerPoolAddress = await keepFactory.getSignerPool(application)

            const signerPool = await SortitionPoolStub.at(signerPoolAddress)

            const operators = await signerPool.getOperators.call()

            assert.deepEqual(
                operators,
                [member1, member2],
                "incorrect registered operators",
            )
        })

        it("does not add an operator to the pool if he is already there", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })
            const signerPoolAddress = await keepFactory.getSignerPool(application)

            const signerPool = await SortitionPoolStub.at(signerPoolAddress)

            assert.deepEqual(
                await signerPool.getOperators(),
                [member1],
                "incorrect registered operators",
            )

            await keepFactory.registerMemberCandidate(application, { from: member1 })

            assert.deepEqual(
                await signerPool.getOperators(),
                [member1],
                "incorrect registered operators after re-registration",
            )
        })

        it("inserts operators to different pools", async () => {
            const application1 = '0x0000000000000000000000000000000000000001'
            const application2 = '0x0000000000000000000000000000000000000002'

            await keepFactory.registerMemberCandidate(application1, { from: member1 })
            await keepFactory.registerMemberCandidate(application2, { from: member2 })

            const signerPool1Address = await keepFactory.getSignerPool(application1)
            const signerPool1 = await SortitionPoolStub.at(signerPool1Address)
            const operators1 = await signerPool1.getOperators.call()

            assert.deepEqual(
                operators1,
                [member1],
                "incorrect registered operators for application 1",
            )

            const signerPool2Address = await keepFactory.getSignerPool(application2)
            const signerPool2 = await SortitionPoolStub.at(signerPool2Address)
            const operators2 = await signerPool2.getOperators.call()

            assert.deepEqual(
                operators2,
                [member2],
                "incorrect registered operators for application 2",
            )
        })
    })

    describe("openKeep", async () => {
        const keepOwner = "0xbc4862697a1099074168d54A555c4A60169c18BD"
        const groupSize = new BN(3)
        const threshold = new BN(3)

        const singleBond = new BN(1)
        const bond = singleBond.mul(groupSize)

        before(async () => {
            // Tests are executed with real implementation of sortition pools.
            // We don't use stub to ensure that keep members selection works correctly.
            sortitionPoolFactory = await SortitionPoolFactory.new()
            keepBonding = await KeepBondingStub.new()
            keepFactory = await ECDSAKeepFactory.new(sortitionPoolFactory.address, keepBonding.address)
        })

        beforeEach(async () => {
            await createSnapshot()
        })

        afterEach(async () => {
            await restoreSnapshot()
        })

        it("reverts if no member candidates are registered", async () => {
            keepFactory = await ECDSAKeepFactory.new(sortitionPoolFactory.address)

            try {
                await keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                )

                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, "No signer pool for this application")
            }
        })

        it("opens keep with multiple members", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })
            await keepFactory.registerMemberCandidate(application, { from: member3 })

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond
                ),
                "No signer pool for this application"
            )
        })

        it("reverts if bond equals zero", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member2, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member3, value: singleBond })

            let bond = 0

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application },
                ),
                "Bond per member equals zero"
            )
        })

        it("reverts if bond per member equals zero", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member2, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member3, value: singleBond })

            let bond = new BN(2)

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application },
                ),
                "Bond per member equals zero"
            )
        })

        it("opens keep with multiple members", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member2, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member3, value: singleBond })

            let blockNumber = await web3.eth.getBlockNumber()

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application },
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
            await keepFactory.registerMemberCandidate(application, { from: member1, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member2, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member3, value: singleBond })

            let blockNumber = await web3.eth.getBlockNumber()

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application },
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
        // size has a reminder the reminder is not bonded, e.g.:
        // requested bond = 11 & group size = 3 => bond per member ≈ 3,66
        // but `11.div(3) = 3` so in current implementation we bond only 9 and
        // the rest remains unbonded.
        // TODO: Check if such case is acceptable.
        it("forgets about the reminder", async () => {
            const groupSize = 3
            const singleBond = new BN(3)
            const bond = new BN(11)

            await keepFactory.registerMemberCandidate(application, { from: member1, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member2, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member3, value: singleBond })

            let blockNumber = await web3.eth.getBlockNumber()

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application },
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
            let groupSize = 2
            let threshold = 2

            await keepFactory.registerMemberCandidate(application, { from: member1, value: singleBond })

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application }
                ),
                "Not enough operators in pool"
            )
        })

        // TODO: This is temporary, we don't expect a group to be formed if a member
        // doesn't have sufficient unbonded value.
        it("reverts if one member has insufficient unbonded value", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member2, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member3, value: singleBond.sub(new BN(1)) })

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application }
                ),
                "Insufficient unbonded value"
            )
        })

        it("opens keep with multiple members and emits an event", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member2, value: singleBond })
            await keepFactory.registerMemberCandidate(application, { from: member3, value: singleBond })

            let blockNumber = await web3.eth.getBlockNumber()

            let keepAddress = await keepFactory.openKeep.call(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application }
            )

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application }
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
    })
})
