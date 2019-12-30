
const KeepBond = artifacts.require('./KeepBondStub.sol')

const { expectRevert } = require('openzeppelin-test-helpers');
const BN = web3.utils.BN

contract('KeepBond', (accounts) => {
    let keepBond

    before(async () => {
        keepBond = await KeepBond.new()
    })

    describe('deposit', async () => {
        it('registers pot', async () => {
            const account = accounts[1]
            const value = new BN(100)

            const expectedPot = (await keepBond.getPot(account)).add(value)

            await keepBond.deposit({ from: account, value: value })

            const pot = await keepBond.getPot(account)

            assert.equal(pot.toString(), expectedPot.toString(), 'invalid pot')
        })
    })

    describe('withdraw', async () => {
        const account = accounts[1]
        const destination = accounts[2]

        beforeEach(async () => {
            const value = new BN(100)
            await keepBond.deposit({ from: account, value: value })
        })

        it('transfers value', async () => {
            const value = (await keepBond.getPot(account))

            const expectedPot = (await keepBond.getPot(account)).sub(value)
            const expectedDestinationBalance = web3.utils.toBN(await web3.eth.getBalance(destination)).add(value)

            await keepBond.withdraw(value, destination, { from: account })

            const pot = await keepBond.getPot(account)
            assert.equal(pot, expectedPot.toNumber(), 'invalid pot')

            const destinationBalance = await web3.eth.getBalance(destination)
            assert.equal(destinationBalance, expectedDestinationBalance.toString(), 'invalid destination balance')
        })

        it('fails if insufficient pot', async () => {
            const value = (await keepBond.getPot(account)).add(new BN(1))

            await expectRevert(
                keepBond.withdraw(value, destination, { from: account }),
                "Insufficient pot"
            )
        })
    })
})
