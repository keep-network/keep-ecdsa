import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const { expectRevert } = require('openzeppelin-test-helpers');

const Registry = artifacts.require('Registry');
const TokenStaking = artifacts.require('TokenStakingStub');
const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory');
const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');
const KeepBondingStub = artifacts.require('KeepBondingStub');
const BondedSortitionPool = artifacts.require('BondedSortitionPool');
const BondedSortitionPoolFactory = artifacts.require('BondedSortitionPoolFactory');

const BN = web3.utils.BN

const chai = require('chai')
chai.use(require('bn-chai')(BN))
const expect = chai.expect

contract("ECDSAKeepFactory", async accounts => {
    let registry
    let tokenStaking
    let keepFactory
    let bondedSortitionPoolFactory
    let keepBonding

    const application = accounts[1]
    const member1 = accounts[2]
    const member2 = accounts[3]
    const member3 = accounts[4]

    describe("registerMemberCandidate", async () => {
        before(async () => {
            registry = await Registry.new()
            tokenStaking = await TokenStaking.new()
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            keepBonding = await KeepBondingStub.new(registry.address, tokenStaking.address)
            keepFactory = await ECDSAKeepFactoryStub.new(bondedSortitionPoolFactory.address, keepBonding.address)
            await registry.approveOperatorContract(keepFactory.address)
            await registry.approveOperatorContract(keepBonding.address)
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

    describe("openKeep", async () => {
        const keepOwner = "0xbc4862697a1099074168d54A555c4A60169c18BD"
        const groupSize = new BN(3)
        const threshold = new BN(3)

        const singleBond = new BN(1)
        const bond = singleBond.mul(groupSize)

        async function initializeNewFactory() {
            // Tests are executed with real implementation of sortition pools.
            // We don't use stub to ensure that keep members selection works correctly.
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            keepBonding = await KeepBondingStub.new(registry.address, tokenStaking.address)
            keepFactory = await ECDSAKeepFactory.new(bondedSortitionPoolFactory.address, keepBonding.address)
            await registry.approveOperatorContract(keepFactory.address)
            await registry.approveOperatorContract(keepBonding.address)
        }

        before(async () => {
            await initializeNewFactory()

            await keepBonding.deposit(member1, { value: singleBond })
            await keepBonding.deposit(member2, { value: singleBond })
            await keepBonding.deposit(member3, { value: singleBond })

            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })
            await keepFactory.registerMemberCandidate(application, { from: member3 })
        })

        beforeEach(async () => {
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
                    { from: application },
                ),
                "Bond per member must be greater than zero"
            )
        })

        it("reverts if Bond per member must be greater than zero", async () => {
            let bond = new BN(2)

            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application },
                ),
                "Bond per member must be greater than zero"
            )
        })

        it("opens keep with multiple members", async () => {
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
                    { from: application }
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
                    { from: application }
                ),
                "Insufficient unbonded value"
            )
        })

        it("opens keep with multiple members and emits an event", async () => {
            await initializeNewFactory()

            await keepBonding.deposit(member1, { value: singleBond })
            await keepBonding.deposit(member2, { value: singleBond })
            await keepBonding.deposit(member3, { value: singleBond })

            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })
            await keepFactory.registerMemberCandidate(application, { from: member3 })

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
