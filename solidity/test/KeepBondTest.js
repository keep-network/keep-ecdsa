const KeepBond = artifacts.require('./KeepBondStub.sol')

const { expectRevert } = require('openzeppelin-test-helpers');

const BN = web3.utils.BN

const chai = require('chai')
chai.use(require('bn-chai')(BN))
const expect = chai.expect

contract('KeepBond', (accounts) => {
    let keepBond

    before(async () => {
        keepBond = await KeepBond.new()
    })

    describe('deposit', async () => {
        it('registers pot', async () => {
            const account = accounts[1]
            const value = new BN(100)

            const expectedPot = (await keepBond.availableForBonding(account)).add(value)

            await keepBond.deposit({ from: account, value: value })

            const pot = await keepBond.availableForBonding(account)

            expect(pot).to.eq.BN(expectedPot, 'invalid pot')
        })
    })

    describe('withdraw', async () => {
        const account = accounts[1]
        const destination = accounts[2]

        before(async () => {
            const value = new BN(100)
            await keepBond.deposit({ from: account, value: value })
        })

        it('transfers value', async () => {
            const value = (await keepBond.availableForBonding(account))

            const expectedPot = (await keepBond.availableForBonding(account)).sub(value)
            const expectedDestinationBalance = web3.utils.toBN(await web3.eth.getBalance(destination)).add(value)

            await keepBond.withdraw(value, destination, { from: account })

            const pot = await keepBond.availableForBonding(account)
            expect(pot).to.eq.BN(expectedPot, 'invalid pot')

            const destinationBalance = await web3.eth.getBalance(destination)
            expect(destinationBalance).to.eq.BN(expectedDestinationBalance, 'invalid destination balance')
        })

        it('fails if insufficient pot', async () => {
            const value = (await keepBond.availableForBonding(account)).add(new BN(1))

            await expectRevert(
                keepBond.withdraw(value, destination, { from: account }),
                "Insufficient pot"
            )
        })
    })

    describe('availableForBonding', async () => {
        const account = accounts[1]
        const value = new BN(100)

        before(async () => {
            await keepBond.deposit({ from: account, value: value })
        })

        it('returns zero for not deposited operator', async () => {
            const operator = "0x0000000000000000000000000000000000000001"
            const expectedPot = 0

            const pot = await keepBond.availableForBonding(operator)

            expect(pot).to.eq.BN(expectedPot, 'invalid pot')
        })

        it('returns value of operators deposit', async () => {
            const operator = account
            const expectedPot = value

            const pot = await keepBond.availableForBonding(operator)

            expect(pot).to.eq.BN(expectedPot, 'invalid pot')
        })
    })

    describe('createBond', async () => {
        const operator = accounts[1]
        const holder = accounts[3]

        beforeEach(async () => {
            keepBond = await KeepBond.new()

            const value = new BN(100)
            await keepBond.deposit({ from: operator, value: value })
        })

        it('creates bond', async () => {
            const value = (await keepBond.availableForBonding(operator))
            const reference = 888

            const expectedPot = (await keepBond.availableForBonding(operator)).sub(value)

            await keepBond.createBond(operator, reference, value, { from: holder })

            const pot = await keepBond.availableForBonding(operator)
            expect(pot).to.eq.BN(expectedPot, 'invalid pot')

            const lockedBonds = await keepBond.getLockedBonds(holder, operator, reference)
            expect(lockedBonds).to.eq.BN(value, 'invalid locked bonds')

            const bondAssignments = await keepBond.getBondAssignments(holder, operator)
            expect(bondAssignments).to.eq.BN(reference, 'invalid bond assignments')
        })

        it('creates two bonds with the same reference', async () => {
            const operator = accounts[2]
            await keepBond.deposit({ from: operator, value: 20 })

            const value = new BN(10)
            const reference = 777

            const expectedPot = (await keepBond.availableForBonding(operator)).sub(value.mul(new BN(2)))

            await keepBond.createBond(operator, reference, value, { from: holder })
            await keepBond.createBond(operator, reference, value, { from: holder })

            const pot = await keepBond.availableForBonding(operator)
            expect(pot).to.eq.BN(expectedPot, 'invalid pot')

            const lockedBonds = await keepBond.getLockedBonds(holder, operator, reference)
            expect(lockedBonds).to.eq.BN(value.mul(new BN(2)), 'invalid locked bonds')

            const bondAssignments = await keepBond.getBondAssignments(holder, operator)
            expect(bondAssignments).to.have.lengthOf(2, 'invalid bond assignment length')
            expect(bondAssignments[0]).to.eq.BN(reference, 'invalid bond assignment 1')
            expect(bondAssignments[1]).to.eq.BN(reference, 'invalid bond assignment 2')
        })

        it('fails if insufficient pot', async () => {
            const value = (await keepBond.availableForBonding(operator)).add(new BN(1))

            await expectRevert(
                keepBond.createBond(operator, 0, value),
                "Insufficient pot"
            )
        })
    })
})
