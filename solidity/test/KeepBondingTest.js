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

            const bondAssignments = await keepBonding.getBondAssignments(holder, operator)
            expect(bondAssignments).to.eq.BN(reference, 'invalid bond assignments')
        })

        it('creates two bonds with the same reference', async () => {
            const bondValue = new BN(10)
            const reference = 777

            const expectedUnbonded = value.sub(bondValue.mul(new BN(2)))

            await keepBonding.createBond(operator, reference, bondValue, { from: holder })
            await keepBonding.createBond(operator, reference, bondValue, { from: holder })

            const unbonded = await keepBonding.availableBondingValue(operator)
            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')

            const lockedBonds = await keepBonding.getLockedBonds(holder, operator, reference)
            expect(lockedBonds).to.eq.BN(bondValue.mul(new BN(2)), 'invalid locked bonds')

            const bondAssignments = await keepBonding.getBondAssignments(holder, operator)
            expect(bondAssignments).to.have.lengthOf(2, 'invalid bond assignment length')
            expect(bondAssignments[0]).to.eq.BN(reference, 'invalid bond assignment 1')
            expect(bondAssignments[1]).to.eq.BN(reference, 'invalid bond assignment 2')
        })

        it('fails if insufficient unbonded value', async () => {
            const bondValue = value.add(new BN(1))

            await expectRevert(
                keepBonding.createBond(operator, 0, bondValue),
                "Insufficient unbonded value"
            )
        })
    })
})
