import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";
import { mineBlocks } from "./helpers/mineBlocks";

const Registry = artifacts.require('./Registry.sol')
const TokenStaking = artifacts.require('./TokenStakingStub.sol')
const EthBonding = artifacts.require('./EthBonding.sol')
const TestEtherReceiver = artifacts.require('./TestEtherReceiver.sol')

const { expectRevert } = require('openzeppelin-test-helpers');

const BN = web3.utils.BN

const chai = require('chai')
chai.use(require('bn-chai')(BN))
const expect = chai.expect

contract('EthBonding', (accounts) => {
    let registry
    let tokenStaking
    let ethBonding
    let etherReceiver

    let operator
    let authorizer
    let notOperator
    let bondCreator
    let sortitionPool

    const initializationPeriod = 10

    before(async () => {
        operator = accounts[1]
        authorizer = operator
        notOperator = accounts[3]
        bondCreator = accounts[4]
        sortitionPool = accounts[5]

        registry = await Registry.new()
        ethBonding = await EthBonding.new(registry.address, initializationPeriod)
        etherReceiver = await TestEtherReceiver.new()

        await registry.approveOperatorContract(bondCreator)
        await ethBonding.authorizeSortitionPoolContract(
            operator,
            sortitionPool,
            { from: operator }
        )

        await ethBonding.authorizeOperatorContract(
            operator,
            bondCreator,
            { from: operator }
        )
    })

    beforeEach(async () => {
        await createSnapshot()
    })

    afterEach(async () => {
        await restoreSnapshot()
    })

    describe('deposit', async () => {
        it('registers unbonded value', async () => {
            const value = new BN(100)

            const expectedUnbonded = value

            await ethBonding.deposit(operator, { value: value })
            await mineBlocks(initializationPeriod + 1)

            const unbonded = await ethBonding.availableUnbondedValue(operator, bondCreator, sortitionPool)

            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })
    })

    describe('withdraw', async () => {
        const value = new BN(1000)

        beforeEach(async () => {
            await ethBonding.deposit(operator, { value: value })
            await mineBlocks(initializationPeriod + 1)
        })

        // it('transfers unbonded value to operator', async () => {
        //     const expectedUnbonded = 0
        //     const expectedOperatorBalance = web3.utils.toBN(
        //         await web3.eth.getBalance(operator)
        //     ).add(value)

        //     await ethBonding.withdraw(value, operator, { from: operator })

        //     const unbonded = await ethBonding.availableUnbondedValue(
        //         operator,
        //         bondCreator,
        //         sortitionPool
        //     )
        //     expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')

        //     const actualOperatorBalance = await web3.eth.getBalance(operator)
        //     expect(actualOperatorBalance).to.eq.BN(
        //         expectedOperatorBalance,
        //         'invalid operator balance'
        //     )
        // })

        it('reverts if insufficient unbonded value', async () => {
            const invalidValue = value.add(new BN(1))

            await expectRevert(
                ethBonding.withdraw(invalidValue, operator, { from: operator }),
                "Insufficient unbonded value"
            )
        })

        it('reverts if transfer fails', async () => {
            await ethBonding.deposit(etherReceiver.address, { value: value })
            await mineBlocks(initializationPeriod + 1)

            await etherReceiver.setShouldFail(true)

            await expectRevert(
                ethBonding.withdraw(
                    value,
                    etherReceiver.address,
                    { from: etherReceiver.address }
                ),
                "Transfer failed"
            )
        })

        it('reverts if someone else is trying to withdraw bond', async () => {
            await expectRevert(
                ethBonding.withdraw(value, operator, { from: notOperator }),
                "Not authorized to withdraw bond"
            )
        })
    })

    describe('availableUnbondedValue', async () => {
        const value = new BN(100)

        beforeEach(async () => {
            await ethBonding.deposit(operator, { value: value })
            await mineBlocks(initializationPeriod + 1)
        })

        it('returns zero for operator with no deposit', async () => {
            const unbondedOperator = "0x0000000000000000000000000000000000000001"
            const expectedUnbonded = 0

            const unbondedValue = await ethBonding.availableUnbondedValue(
                unbondedOperator,
                bondCreator,
                sortitionPool
            )
            expect(unbondedValue).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })

        it('return zero when bond creator is not approved by operator', async () => {
            const notApprovedBondCreator = "0x0000000000000000000000000000000000000001"
            const expectedUnbonded = 0

            const unbondedValue = await ethBonding.availableUnbondedValue(
                operator,
                notApprovedBondCreator,
                sortitionPool
            )
            expect(unbondedValue).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })

        it('returns zero when sortition pool is not authorized', async () => {
            const notAuthorizedSortitionPool = "0x0000000000000000000000000000000000000001"
            const expectedUnbonded = 0

            const unbondedValue = await ethBonding.availableUnbondedValue(
                operator,
                bondCreator,
                notAuthorizedSortitionPool
            )
            expect(unbondedValue).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })

        it('returns value of operators deposit', async () => {
            const expectedUnbonded = value

            const unbonded = await ethBonding.availableUnbondedValue(operator, bondCreator, sortitionPool)

            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })

        it('returns 0 when the initialization period is not over', async () => {
            await ethBonding.deposit(notOperator, { value: value })
            await mineBlocks(initializationPeriod)
            const expectedUnbonded = 0

            const unbonded = await ethBonding.availableUnbondedValue(
                notOperator,
                bondCreator,
                sortitionPool
            )

            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')
        })
    })

    describe('createBond', async () => {
        const holder = accounts[3]
        const value = new BN(100)

        beforeEach(async () => {
            await ethBonding.deposit(operator, { value: value })
            await mineBlocks(initializationPeriod + 1)
        })

        it('creates bond', async () => {
            const reference = 888

            const expectedUnbonded = 0

            await ethBonding.createBond(
                operator,
                holder,
                reference,
                value,
                sortitionPool,
                { from: bondCreator }
            )

            const unbonded = await ethBonding.availableUnbondedValue(
                operator,
                bondCreator,
                sortitionPool
            )
            expect(unbonded).to.eq.BN(expectedUnbonded, 'invalid unbonded value')

            const lockedBonds = await ethBonding.bondAmount(
                operator,
                holder,
                reference
            )
            expect(lockedBonds).to.eq.BN(value, 'unexpected bond value')
        })

        it('creates two bonds with the same reference for different operators', async () => {
            const operator2 = accounts[2]
            const bondValue = new BN(10)
            const reference = 777

            const expectedUnbonded = value.sub(bondValue)

            await ethBonding.deposit(operator2, { value: value })
            await mineBlocks(initializationPeriod + 1)

            await ethBonding.authorizeOperatorContract(
                operator2,
                bondCreator,
                { from: operator2 }
            )
            await ethBonding.authorizeSortitionPoolContract(
                operator2,
                sortitionPool,
                { from: operator2 }
            )
            await ethBonding.createBond(
                operator,
                holder,
                reference,
                bondValue,
                sortitionPool,
                { from: bondCreator }
            )
            await ethBonding.createBond(
                operator2,
                holder,
                reference,
                bondValue,
                sortitionPool,
                { from: bondCreator }
            )

            const unbonded1 = await ethBonding.availableUnbondedValue(
                operator, bondCreator, sortitionPool
            )
            expect(unbonded1).to.eq.BN(expectedUnbonded, 'invalid unbonded value 1')

            const unbonded2 = await ethBonding.availableUnbondedValue(
                operator2, bondCreator, sortitionPool
            )
            expect(unbonded2).to.eq.BN(expectedUnbonded, 'invalid unbonded value 2')

            const lockedBonds1 = await ethBonding.bondAmount(
                operator, holder, reference
            )
            expect(lockedBonds1).to.eq.BN(bondValue, 'unexpected bond value 1')

            const lockedBonds2 = await ethBonding.bondAmount(
                operator2, holder, reference
            )
            expect(lockedBonds2).to.eq.BN(bondValue, 'unexpected bond value 2')
        })

        it('fails to create two bonds with the same reference for the same operator', async () => {
            const bondValue = new BN(10)
            const reference = 777

            await ethBonding.createBond(
                operator, holder,
                reference, bondValue,
                sortitionPool,
                { from: bondCreator }
            )

            await expectRevert(
                ethBonding.createBond(
                    operator, holder,
                    reference, bondValue,
                    sortitionPool,
                    { from: bondCreator }
                ),
                "Reference ID not unique for holder and operator"
            )
        })

        it('fails if insufficient unbonded value', async () => {
            const bondValue = value.add(new BN(1))

            await expectRevert(
                ethBonding.createBond(operator, holder, 0, bondValue, sortitionPool, { from: bondCreator }),
                "Insufficient unbonded value"
            )
        })

        it('fails if not initialized', async () => {
            const bondValue = value.add(new BN(1))
            await ethBonding.deposit(notOperator, { value: value })
            await ethBonding.authorizeSortitionPoolContract(
                notOperator,
                sortitionPool,
                { from: notOperator }
            )

            await ethBonding.authorizeOperatorContract(
                notOperator,
                bondCreator,
                { from: notOperator }
            )

            await expectRevert(
                ethBonding.createBond(notOperator, holder, 0, bondValue, sortitionPool, { from: bondCreator }),
                "Insufficient unbonded value"
            )
        })
    })

    describe('reassignBond', async () => {
        const holder = accounts[2]
        const newHolder = accounts[3]
        const bondValue = new BN(100)
        const reference = 777
        const newReference = 888

        beforeEach(async () => {
            await ethBonding.deposit(operator, { value: bondValue })
            await mineBlocks(initializationPeriod + 1)
            await ethBonding.createBond(operator, holder, reference, bondValue, sortitionPool, { from: bondCreator })
        })

        it('reassigns bond to a new holder and a new reference', async () => {
            await ethBonding.reassignBond(operator, reference, newHolder, newReference, { from: holder })

            let lockedBonds = await ethBonding.bondAmount(operator, holder, reference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await ethBonding.bondAmount(operator, holder, newReference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await ethBonding.bondAmount(operator, newHolder, reference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await ethBonding.bondAmount(operator, newHolder, newReference)
            expect(lockedBonds).to.eq.BN(bondValue, 'invalid locked bonds')
        })

        it('reassigns bond to the same holder and a new reference', async () => {
            await ethBonding.reassignBond(operator, reference, holder, newReference, { from: holder })

            let lockedBonds = await ethBonding.bondAmount(operator, holder, reference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await ethBonding.bondAmount(operator, holder, newReference)
            expect(lockedBonds).to.eq.BN(bondValue, 'invalid locked bonds')
        })

        it('reassigns bond to a new holder and the same reference', async () => {
            await ethBonding.reassignBond(operator, reference, newHolder, reference, { from: holder })

            let lockedBonds = await ethBonding.bondAmount(operator, holder, reference)
            expect(lockedBonds).to.eq.BN(0, 'invalid locked bonds')

            lockedBonds = await ethBonding.bondAmount(operator, newHolder, reference)
            expect(lockedBonds).to.eq.BN(bondValue, 'invalid locked bonds')
        })

        it('fails if sender is not the holder', async () => {
            await expectRevert(
                ethBonding.reassignBond(operator, reference, newHolder, newReference, { from: accounts[0] }),
                "Bond not found"
            )
        })

        it('fails if reassigned to the same holder and the same reference', async () => {
            await expectRevert(
                ethBonding.reassignBond(operator, reference, holder, reference, { from: holder }),
                "Reference ID not unique for holder and operator"
            )
        })
    })

    describe('freeBond', async () => {
        const holder = accounts[2]
        const initialUnboundedValue = new BN(500)
        const bondValue = new BN(100)
        const reference = 777

        beforeEach(async () => {
            await ethBonding.deposit(operator, { value: initialUnboundedValue })
            await mineBlocks(initializationPeriod + 1)
            await ethBonding.createBond(operator, holder, reference, bondValue, sortitionPool, { from: bondCreator })
        })

        it('releases bond amount to operator\'s available bonding value', async () => {
            await ethBonding.freeBond(operator, reference, { from: holder })

            const lockedBonds = await ethBonding.bondAmount(operator, holder, reference)
            expect(lockedBonds).to.eq.BN(0, 'unexpected remaining locked bonds')

            const unbondedValue = await ethBonding.availableUnbondedValue(operator, bondCreator, sortitionPool)
            expect(unbondedValue).to.eq.BN(initialUnboundedValue, 'unexpected unbonded value')
        })

        it('fails if sender is not the holder', async () => {
            await expectRevert(
                ethBonding.freeBond(operator, reference, { from: accounts[0] }),
                "Bond not found"
            )
        })
    })

    describe('seizeBond', async () => {
        const holder = accounts[2]
        const destination = accounts[3]
        const bondValue = new BN(1000)
        const reference = 777

        beforeEach(async () => {
            await ethBonding.deposit(operator, { value: bondValue })
            await mineBlocks(initializationPeriod + 1)
            await ethBonding.createBond(operator, holder, reference, bondValue, sortitionPool, { from: bondCreator })
        })

        it('transfers whole bond amount to destination account', async () => {
            const amount = bondValue
            let expectedBalance = web3.utils.toBN(await web3.eth.getBalance(destination)).add(amount)

            await ethBonding.seizeBond(operator, reference, amount, destination, { from: holder })

            const actualBalance = await web3.eth.getBalance(destination)
            expect(actualBalance).to.eq.BN(expectedBalance, 'invalid destination account balance')

            const lockedBonds = await ethBonding.bondAmount(operator, holder, reference)
            expect(lockedBonds).to.eq.BN(0, 'unexpected remaining bond value')
        })

        it('transfers less than bond amount to destination account', async () => {
            const remainingBond = new BN(1)
            const amount = bondValue.sub(remainingBond)
            let expectedBalance = web3.utils.toBN(await web3.eth.getBalance(destination)).add(amount)

            await ethBonding.seizeBond(operator, reference, amount, destination, { from: holder })

            const actualBalance = await web3.eth.getBalance(destination)
            expect(actualBalance).to.eq.BN(expectedBalance, 'invalid destination account balance')

            const lockedBonds = await ethBonding.bondAmount(operator, holder, reference)
            expect(lockedBonds).to.eq.BN(remainingBond, 'unexpected remaining bond value')
        })

        it('reverts if seized amount equals zero', async () => {
            const amount = new BN(0)
            await expectRevert(
                ethBonding.seizeBond(operator, reference, amount, destination, { from: holder }),
                "Requested amount should be greater than zero"
            )
        })

        it('reverts if seized amount is greater than bond value', async () => {
            const amount = bondValue.add(new BN(1))
            await expectRevert(
                ethBonding.seizeBond(operator, reference, amount, destination, { from: holder }),
                "Requested amount is greater than the bond"
            )
        })

        it('reverts if transfer fails', async () => {
            await etherReceiver.setShouldFail(true)
            const destination = etherReceiver.address

            await expectRevert(
                ethBonding.seizeBond(operator, reference, bondValue, destination, { from: holder }),
                "Transfer failed"
            )

            const destinationBalance = await web3.eth.getBalance(destination)
            expect(destinationBalance).to.eq.BN(0, 'invalid destination account balance')

            const lockedBonds = await ethBonding.bondAmount(operator, holder, reference)
            expect(lockedBonds).to.eq.BN(bondValue, 'unexpected bond value')
        })
    })

    describe("authorizeSortitionPoolContract", async () => {
        it("reverts when operator is not an authorizer", async () => {
            let authorizer1 = accounts[2]

            await expectRevert(
                ethBonding.authorizeSortitionPoolContract(operator, sortitionPool, { from: authorizer1 }),
                'Not authorized'
            )
        })

        it("should authorize sortition pool for provided operator", async () => {
            await ethBonding.authorizeSortitionPoolContract(
                operator,
                sortitionPool,
                { from: operator }
            )

            assert.isTrue(
                await ethBonding.hasSecondaryAuthorization(operator, sortitionPool),
                "Sortition pool has not beeen authorized for provided operator"
            )
        })
    })
})
