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

    describe('deposit', async () => {
        it('registers unbonded value', async () => {
            const account = accounts[1]
            const value = new BN(100)

            const expectedUnbonded = (await keepBonding.availableForBonding(account)).add(value)

            await keepBonding.deposit(account, { value: value })

            const unbonded = await keepBonding.availableForBonding(account)

            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })
    })

    describe('withdraw', async () => {
        const account = accounts[1]
        const destination = accounts[2]

        before(async () => {
            const value = new BN(100)
            await keepBonding.deposit(account, { value: value })
        })

        it('transfers unbonded value', async () => {
            const value = (await keepBonding.availableForBonding(account))

            const expectedUnbonded = (await keepBonding.availableForBonding(account)).sub(value)
            const expectedDestinationBalance = web3.utils.toBN(await web3.eth.getBalance(destination)).add(value)

            await keepBonding.withdraw(value, destination, { from: account })

            const unbonded = await keepBonding.availableForBonding(account)
            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')

            const destinationBalance = await web3.eth.getBalance(destination)
            expect(destinationBalance).to.eq.BN(expectedDestinationBalance, 'invalid destination balance')
        })

        it('fails if insufficient unbonded value', async () => {
            const value = (await keepBonding.availableForBonding(account)).add(new BN(1))

            await expectRevert(
                keepBonding.withdraw(value, destination, { from: account }),
                "Insufficient unbonded value"
            )
        })
    })

    describe('availableForBonding', async () => {
        const account = accounts[1]
        const value = new BN(100)

        before(async () => {
            await keepBonding.deposit(account, { value: value })
        })

        it('returns zero for not deposited operator', async () => {
            const operator = "0x0000000000000000000000000000000000000001"
            const expectedUnbonded = 0

            const unbondedValue = await keepBonding.availableForBonding(operator)

            expect(unbondedValue).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })

        it('returns value of operators deposit', async () => {
            const operator = account
            const expectedUnbonded = value

            const unbonded = await keepBonding.availableForBonding(operator)

            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })
    })
})
