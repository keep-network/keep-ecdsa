import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const { expectRevert } = require('openzeppelin-test-helpers');

import { getETHBalancesFromList, addToBalances } from './helpers/listBalanceUtils'

const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');
const KeepBonding = artifacts.require('KeepBonding');
const TokenStakingStub = artifacts.require("TokenStakingStub")
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
            keepBonding = await KeepBonding.new()
            randomBeacon = await RandomBeaconStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon.address
            )

            const stakeBalance = await keepFactory.minimumStake.call()
            await tokenStaking.setBalance(stakeBalance);

            const bondingValue = new BN(100)
            await keepBonding.deposit(member1, { value: bondingValue })
            await keepBonding.deposit(member2, { value: bondingValue })
            await keepBonding.deposit(member3, { value: bondingValue })
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
            const minimumStake = await keepFactory.minimumStake.call()
            const minimumStakeMultiplier = new BN("10")
            await tokenStaking.setBalance(minimumStake.mul(minimumStakeMultiplier))

            await keepFactory.registerMemberCandidate(application, { from: member1 })

            const signerPoolAddress = await keepFactory.getSignerPool(application)
            const signerPool = await BondedSortitionPool.at(signerPoolAddress)

            const actualWeight = await signerPool.getPoolWeight.call(member1)
            const expectedWeight = minimumStakeMultiplier

            expect(actualWeight).to.eq.BN(expectedWeight, 'invalid staking weight')
        })

        it("inserts operators to the same pool", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })

            const signerPoolAddress = await keepFactory.getSignerPool(application)

            const signerPool = await BondedSortitionPool.at(signerPoolAddress)

            assert.isTrue(await signerPool.isOperatorInPool(member1), "operator 1 is not in the pool")
            assert.isTrue(await signerPool.isOperatorInPool(member2), "operator 2 is not in the pool")
        })

        it("does not add an operator to the pool if it is already there", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })
            const signerPoolAddress = await keepFactory.getSignerPool(application)

            const signerPool = await BondedSortitionPool.at(signerPoolAddress)

            assert.isTrue(await signerPool.isOperatorInPool(member1), "operator is not in the pool")

            await keepFactory.registerMemberCandidate(application, { from: member1 })

            assert.isTrue(await signerPool.isOperatorInPool(member1), "operator is not in the pool")
        })

        it("does not add an operator to the pool if it does not have a minimum stake", async() => {
            await tokenStaking.setBalance(new BN("1"))

            await expectRevert(
                keepFactory.registerMemberCandidate(application, { from: member1 }),
                "Operator not eligible"
            )
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
            keepBonding = await KeepBonding.new()
            randomBeacon = await RandomBeaconStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon.address
            )

            const stakeBalance = await keepFactory.minimumStake.call()
            await tokenStaking.setBalance(stakeBalance);

            const bondingValue = new BN(100)
            await keepBonding.deposit(member1, { value: bondingValue })
            await keepBonding.deposit(member2, { value: bondingValue })
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

    describe("isOperatorRegistered", async () => {
        before(async () => {
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            tokenStaking = await TokenStakingStub.new()
            keepBonding = await KeepBonding.new()
            randomBeacon = await RandomBeaconStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon.address
            )

            const stakeBalance = await keepFactory.minimumStake.call()
            await tokenStaking.setBalance(stakeBalance);

            const bondingValue = new BN(100)
            await keepBonding.deposit(member1, { value: bondingValue })
            await keepBonding.deposit(member2, { value: bondingValue })
            await keepBonding.deposit(member3, { value: bondingValue })
        })

        beforeEach(async () => {
            await createSnapshot()
        })

        afterEach(async () => {
            await restoreSnapshot()
        })

        it("returns true if the operator is registered for the application", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })

            assert.isTrue(await keepFactory.isOperatorRegistered(member1, application))
        })

        it("returns false if the operator is registered for another application", async () => {
            const application2 = '0x0000000000000000000000000000000000000002'

            await keepFactory.registerMemberCandidate(application, { from: member1 })

            assert.isFalse(await keepFactory.isOperatorRegistered(member1, application2))
        })

        it("returns false if the operator is not registered for any application", async () => {
            assert.isFalse(await keepFactory.isOperatorRegistered(member1, application))
        })
    })

    describe("isOperatorUpToDate", async () => {
        let minimumStake
        let minimumBondingValue

        before(async () => {
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            tokenStaking = await TokenStakingStub.new()
            keepBonding = await KeepBonding.new()
            randomBeacon = await RandomBeaconStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon.address
            )

            minimumStake = await keepFactory.minimumStake.call()
            await tokenStaking.setBalance(minimumStake);

            minimumBondingValue = await keepFactory.minimumBond.call()
            await keepBonding.deposit(member1, { value: minimumBondingValue })
        })

        beforeEach(async () => {
            await createSnapshot()
        })

        afterEach(async () => {
            await restoreSnapshot()
        })

        it("returns true if the operator is up to date for the application", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })

            assert.isTrue(await keepFactory.isOperatorUpToDate(member1, application))
        })

        it("returns false if the operator stake is below minimum", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })

            tokenStaking.setBalance(minimumStake.sub(new BN(1)))

            assert.isFalse(await keepFactory.isOperatorUpToDate(member1, application))
        })

        it("returns true if the operator stake changed insufficiently", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })

            // We multiply minimumStake as sortition pools expect multiplies of the
            // minimum stake to calculate stakers weight for eligibility.
            // We subtract 1 to get the same staking weight which is calculated as
            // `weight = floor(stakingBalance / minimumStake)`.
            tokenStaking.setBalance(minimumStake.mul(new BN(2)).sub(new BN(1)))

            assert.isTrue(await keepFactory.isOperatorUpToDate(member1, application))
        })

        it("returns false if the operator stake is above minimum", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })

            // We multiply minimumStake as sortition pools expect multiplies of the
            // minimum stake to calculate stakers weight for eligibility.
            tokenStaking.setBalance(minimumStake.mul(new BN(2)))

            assert.isFalse(await keepFactory.isOperatorUpToDate(member1, application))
        })

        it("returns false if the operator bonding value is below minimum", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })

            keepBonding.withdraw(new BN(1), member1, { from: member1 })

            assert.isFalse(await keepFactory.isOperatorUpToDate(member1, application))
        })

        it("returns true if the operator bonding value is above minimum", async () => {
            await keepFactory.registerMemberCandidate(application, { from: member1 })

            keepBonding.deposit(member1, { value: new BN(1) })

            assert.isTrue(await keepFactory.isOperatorUpToDate(member1, application))
        })


        it("reverts if the operator is not registered for the application", async () => {
            await expectRevert(
                keepFactory.isOperatorUpToDate(member2, application),
                "Operator not registered for the application"
            )
        })
    })

    describe("updateOperatorStatus", async () => {
        let minimumStake
        let minimumBondingValue

        before(async () => {
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            tokenStaking = await TokenStakingStub.new()
            keepBonding = await KeepBonding.new()
            randomBeacon = await RandomBeaconStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon.address
            )

            minimumStake = await keepFactory.minimumStake.call()
            await tokenStaking.setBalance(minimumStake);

            minimumBondingValue = await keepFactory.minimumBond.call()
            await keepBonding.deposit(member1, { value: minimumBondingValue })

            await keepFactory.registerMemberCandidate(application, { from: member1 })
        })

        beforeEach(async () => {
            await createSnapshot()
        })

        afterEach(async () => {
            await restoreSnapshot()
        })

        it("revers if operator is up to date", async () => {
            await expectRevert(
                keepFactory.updateOperatorStatus(member1, application),
                "Operator already up to date"
            )
        })

        it("removes operator if stake has changed below minimum", async () => {
            tokenStaking.setBalance(minimumStake.sub(new BN(1)))
            assert.isFalse(
                await keepFactory.isOperatorUpToDate(member1, application),
                "unexpected status of the operator after stake change"
            )

            await keepFactory.updateOperatorStatus(member1, application)

            await expectRevert(
                keepFactory.isOperatorUpToDate(member1, application),
                "Operator not registered for the application"
            )
        })

        it("updates operator if stake has changed above minimum", async () => {
            // We multiply minimumStake as sortition pools expect multiplies of the
            // minimum stake to calculate stakers weight for eligibility.
            tokenStaking.setBalance(minimumStake.mul(new BN(2)))
            assert.isFalse(
                await keepFactory.isOperatorUpToDate(member1, application),
                "unexpected status of the operator after stake change"
            )

            await keepFactory.updateOperatorStatus(member1, application)

            assert.isTrue(
                await keepFactory.isOperatorUpToDate(member1, application),
                "unexpected status of the operator after status update"
            )
        })

        it("removes operator if bonding value has changed below minimum", async () => {
            keepBonding.withdraw(new BN(1), member1, { from: member1 })
            assert.isFalse(
                await keepFactory.isOperatorUpToDate(member1, application),
                "unexpected status of the operator after bonding value change"
            )

            await keepFactory.updateOperatorStatus(member1, application)

            await expectRevert(
                keepFactory.isOperatorUpToDate(member1, application),
                "Operator not registered for the application"
            )
        })

        it("updates operator if bonding value has changed above minimum", async () => {
            keepBonding.deposit(member1, { value: new BN(1) })
            assert.isTrue(
                await keepFactory.isOperatorUpToDate(member1, application),
                "unexpected status of the operator after bonding value change"
            )

            await expectRevert(
                keepFactory.updateOperatorStatus(member1, application),
                "Operator already up to date"
            )
        })

        it("reverts if the operator is not registered for the application", async () => {
            await expectRevert(
                keepFactory.updateOperatorStatus(member2, application),
                "Operator not registered for the application"
            )
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
            keepBonding = await KeepBonding.new()
            randomBeacon = await RandomBeaconStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon.address
            )

            feeEstimate = await keepFactory.openKeepFeeEstimate()
        }

        beforeEach(async () => {
            await initializeNewFactory()

            const stakeBalance = await keepFactory.minimumStake.call()
            await tokenStaking.setBalance(stakeBalance)

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
            await expectRevert(
                keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { value: feeEstimate }
                ),
                "No signer pool for this application"
            )
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
                await keepBonding.bondAmount(member1, keepAddress, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member1')

            expect(
                await keepBonding.bondAmount(member2, keepAddress, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member2')

            expect(
                await keepBonding.bondAmount(member3, keepAddress, keepAddress)
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

            const stakeBalance = await keepFactory.minimumStake.call()
            await tokenStaking.setBalance(stakeBalance)

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
                await keepBonding.bondAmount(member1, keepAddress, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member1')

            expect(
                await keepBonding.bondAmount(member2, keepAddress, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member2')

            expect(
                await keepBonding.bondAmount(member3, keepAddress, keepAddress)
            ).to.eq.BN(singleBond, 'invalid bond value for member3')
        })

        it("reverts if not enough member candidates are registered", async () => {
            await initializeNewFactory()

            let groupSize = 2
            let threshold = 2

            const stakeBalance = await keepFactory.minimumStake.call()
            await tokenStaking.setBalance(stakeBalance)

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

            const stakeBalance = await keepFactory.minimumStake.call()
            await tokenStaking.setBalance(stakeBalance)

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

        it("splits subsidy pool between selected signers", async () => {
            const members = [ member1, member2, member3 ];
            const subsidyPool = 50; // [wei]

            // pump subsidy pool
            web3.eth.sendTransaction({
                value: subsidyPool, 
                from: accounts[0],
                to: keepFactory.address
            });

            const initialBalances = await getETHBalancesFromList(members)

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application, value: feeEstimate },
            )

            const newBalances = await getETHBalancesFromList(members)
            const expectedBalances = addToBalances(initialBalances, subsidyPool / members.length)

            assert.equal(newBalances.toString(), expectedBalances.toString())

            expect(await keepFactory.subsidyPool()).to.eq.BN(
                0,
                "subsidy pool should go down to 0"
            )
        })

        it("does not transfer more from subsidy pool than entry fee", async () => {
            const members = [ member1, member2, member3 ];
            const subsidyPool = feeEstimate * 10; // [wei]

            // pump subsidy pool
            web3.eth.sendTransaction({
                value: subsidyPool, 
                from: accounts[0],
                to: keepFactory.address
            });

            const initialBalances = await getETHBalancesFromList(members)

            await keepFactory.openKeep(
                groupSize,
                threshold,
                keepOwner,
                bond,
                { from: application, value: feeEstimate },
            )

            const newBalances = await getETHBalancesFromList(members)
            const expectedBalances = addToBalances(initialBalances, feeEstimate / members.length)

            assert.equal(newBalances.toString(), expectedBalances.toString()) 

            expect(await keepFactory.subsidyPool()).to.eq.BN(
                subsidyPool - feeEstimate,
                "unexpected subsidy pool balance"
            )
        })
    })

    describe("setGroupSelectionSeed", async () => {
        const newGroupSelectionSeed = new BN(2345675)

        before(async () => {
            bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
            tokenStaking = await TokenStakingStub.new()
            keepBonding = await KeepBonding.new()
            randomBeacon = accounts[1]
            keepFactory = await ECDSAKeepFactoryStub.new(
                bondedSortitionPoolFactory.address,
                tokenStaking.address,
                keepBonding.address,
                randomBeacon
            )
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
