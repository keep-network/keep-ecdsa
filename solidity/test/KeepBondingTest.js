import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const KeepBonding = artifacts.require('./KeepBonding.sol')

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

    xdescribe('deposit', async () => {
        it('registers unbonded value', async () => {
            const operator = accounts[1]
            const value = new BN(100)

            const expectedUnbonded = value

            await keepBonding.deposit(operator, { value: value })

            const unbonded = await keepBonding.availableBondingValue(operator)

            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })
    })

    xdescribe('withdraw', async () => {
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

    xdescribe('availableBondingValue', async () => {
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
})