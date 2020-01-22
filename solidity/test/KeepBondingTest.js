import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const KeepBonding = artifacts.require('./KeepBondingStub.sol')

const { expectRevert } = require('openzeppelin-test-helpers');

const BN = web3.utils.BN

const chai = require('chai')
chai.use(require('bn-chai')(BN))
const expect = chai.expect

contract('KeepBonding', (accounts) => {
    let keepBonding

    before(async () => {
        keepBonding = await KeepBonding.new()
    })

    beforeEach(async () => {
        await createSnapshot()
    })

    afterEach(async () => {
        await restoreSnapshot()
    })

    describe('deposit', async () => {
        it('registers unbonded value', async () => {
            const operator = accounts[1]
            const value = new BN(100)

            const expectedUnbonded = value

            await keepBonding.deposit(operator, { value: value })

            const unbonded = await keepBonding.availableBondingValue(operator)

            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })
    })

    describe('withdraw', async () => {
        const operator = accounts[1]
        const destination = accounts[2]
        const value = new BN(100)

        beforeEach(async () => {
            await keepBonding.deposit(operator, { value: value })
        })

        it('transfers unbonded value', async () => {
            const expectedUnbonded = 0
            const expectedDestinationBalance = web3.utils.toBN(await web3.eth.getBalance(destination)).add(value)

            await keepBonding.withdraw(value, destination, { from: operator })

            const unbonded = await keepBonding.availableBondingValue(operator)
            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')

            const destinationBalance = await web3.eth.getBalance(destination)
            expect(destinationBalance).to.eq.BN(expectedDestinationBalance, 'invalid destination balance')
        })

        it('fails if insufficient unbonded value', async () => {
            const invalidValue = value.add(new BN(1))

            await expectRevert(
                keepBonding.withdraw(invalidValue, destination, { from: operator }),
                "Insufficient unbonded value"
            )
        })
    })

    describe('availableBondingValue', async () => {
        const operator = accounts[1]
        const value = new BN(100)

        beforeEach(async () => {
            await keepBonding.deposit(operator, { value: value })
        })

        it('returns zero for not deposited operator', async () => {
            const unbondedOperator = "0x0000000000000000000000000000000000000001"
            const expectedUnbonded = 0

            const unbondedValue = await keepBonding.availableBondingValue(unbondedOperator)

            expect(unbondedValue).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })

        it('returns value of operators deposit', async () => {
            const expectedUnbonded = value

            const unbonded = await keepBonding.availableBondingValue(operator)

            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })
    })

    describe('createBond', async () => {
        const operator = accounts[1]
        const holder = accounts[3]
        const value = new BN(100)

        beforeEach(async () => {
            await keepBonding.deposit(operator, { value: value })
        })

        it('creates bond', async () => {
            const reference = 888

            const expectedUnbonded = 0

            await keepBonding.createBond(operator, reference, value, { from: holder })

            const unbonded = await keepBonding.availableBondingValue(operator)
            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')

            const lockedBonds = await keepBonding.getLockedBonds(holder, operator, reference)
            expect(lockedBonds).to.eq.BN(value, 'invalid locked bonds')
        })

        it('creates two bonds with the same reference for different operators', async () => {
            const operator2 = accounts[2]
            const bondValue = new BN(10)
            const reference = 777

            const expectedUnbonded = value.sub(bondValue)

            await keepBonding.deposit(operator2, { value: value })

            await keepBonding.createBond(operator, reference, bondValue, { from: holder })
            await keepBonding.createBond(operator2, reference, bondValue, { from: holder })

            const unbonded1 = await keepBonding.availableBondingValue(operator)
            expect(unbonded1).to.eq.BN(expectedUnbonded, 'invalid unbonded value 1')

            const unbonded2 = await keepBonding.availableBondingValue(operator2)
            expect(unbonded2).to.eq.BN(expectedUnbonded, 'invalid unbonded value 2')

            const lockedBonds1 = await keepBonding.getLockedBonds(holder, operator, reference)
            expect(lockedBonds1).to.eq.BN(bondValue, 'invalid locked bonds')

            const lockedBonds2 = await keepBonding.getLockedBonds(holder, operator2, reference)
            expect(lockedBonds2).to.eq.BN(bondValue, 'invalid locked bonds')
        })

        it('fails to create two bonds with the same reference for the same operator', async () => {
            const bondValue = new BN(10)
            const reference = 777

            await keepBonding.createBond(operator, reference, bondValue, { from: holder })

            await expectRevert(
                keepBonding.createBond(operator, reference, bondValue, { from: holder }),
                "Reference ID not unique for holder and operator"
            )
        })

        it('fails if insufficient unbonded value', async () => {
            const bondValue = value.add(new BN(1))

            await expectRevert(
                keepBonding.createBond(operator, 0, bondValue),
                "Insufficient unbonded value"
            )
        })
    })

    describe('reassignBond', async () => {
        const operator = accounts[1]
        const holder = accounts[2]
        const newHolder = accounts[3]
        const bondValue = new BN(100)
        const reference = 777
        const newReference = 888

        beforeEach(async () => {
            await keepBonding.deposit(operator, { value: bondValue })
            await keepBonding.createBond(operator, reference, bondValue, { from: holder })
        })

        it('reassigns bond to a new holder and a new reference', async () => {
            await keepBonding.reassignBond(operator, reference, newHolder, newReference, { from: holder })

            let lockedBonds = await keepBonding.getLockedBonds(holder, operator, reference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await keepBonding.getLockedBonds(holder, operator, newReference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await keepBonding.getLockedBonds(newHolder, operator, reference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await keepBonding.getLockedBonds(newHolder, operator, newReference)
            expect(lockedBonds).to.eq.BN(bondValue, 'invalid locked bonds')
        })

        it('reassigns bond to the same holder and a new reference', async () => {
            await keepBonding.reassignBond(operator, reference, holder, newReference, { from: holder })

            let lockedBonds = await keepBonding.getLockedBonds(holder, operator, reference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await keepBonding.getLockedBonds(holder, operator, newReference)
            expect(lockedBonds).to.eq.BN(bondValue, 'invalid locked bonds')
        })

        it('reassigns bond to a new holder and the same reference', async () => {
            await keepBonding.reassignBond(operator, reference, newHolder, reference, { from: holder })

            let lockedBonds = await keepBonding.getLockedBonds(holder, operator, reference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await keepBonding.getLockedBonds(newHolder, operator, reference)
            expect(lockedBonds).to.eq.BN(bondValue, 'invalid locked bonds')
        })

        it('fails if sender is not the holder', async () => {
            await expectRevert(
                keepBonding.reassignBond(operator, reference, newHolder, newReference, { from: accounts[0] }),
                "Bond not found"
            )
        })

        it('fails if reassigned to the same holder and the same reference', async () => {
            await keepBonding.deposit(operator, { value: bondValue })
            await keepBonding.createBond(operator, newReference, bondValue, { from: holder })

            await expectRevert(
                keepBonding.reassignBond(operator, reference, holder, newReference, { from: holder }),
                "Reference ID not unique for holder and operator"
            )
        })
    })

    describe('freeBond', async () => {
        const operator = accounts[1]
        const holder = accounts[2]
        const bondValue = new BN(100)
        const reference = 777

        beforeEach(async () => {
            await keepBonding.deposit(operator, { value: bondValue })
            await keepBonding.createBond(operator, reference, bondValue, { from: holder })
        })

        it('releases bond amount to operator\'s available bonding value', async () => {
            await keepBonding.freeBond(operator, reference, { from: holder })

            const lockedBonds = await keepBonding.getLockedBonds(holder, operator, reference)
            expect(lockedBonds).to.eq.BN(0, 'unexpected remaining locked bonds')

            const unbondedValue = await keepBonding.availableBondingValue(operator)
            expect(unbondedValue).to.eq.BN(bondValue, 'unexpected unbonded value')
        })

        it('fails if sender is not the holder', async () => {
            await expectRevert(
                keepBonding.freeBond(operator, reference, { from: accounts[0] }),
                "Bond not found"
            )
        })
    })
})
