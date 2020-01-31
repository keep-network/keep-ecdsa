const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory');
const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');
const KeepBondingStub = artifacts.require('KeepBondingStub');
const SortitionPoolFactoryStub = artifacts.require('SortitionPoolFactoryStub');
const SortitionPoolStub = artifacts.require('SortitionPoolStub');
const SortitionPoolFactory = artifacts.require('SortitionPoolFactory');

contract("ECDSAKeepFactory", async accounts => {
    let keepFactory
    let sortitionPoolFactory
    let keepBonding

    const application = accounts[1]
    const member1 = accounts[2]
    const member2 = accounts[3]
    const member3 = accounts[4]

    describe("registerMemberCandidate", async () => {
        beforeEach(async () => {
            sortitionPoolFactory = await SortitionPoolFactoryStub.new()
            keepBonding = await KeepBondingStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(sortitionPoolFactory.address, keepBonding.address)
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

            assert.include(
                eventList[0].returnValues.members,
                member1,
                "array doesn't include member1",
            )
            assert.include(
                eventList[0].returnValues.members,
                member2,
                "array doesn't include member2",
            )
            assert.include(
                eventList[0].returnValues.members,
                member3,
                "array doesn't include member3",
            )
        })

        it("reverts if not enough member candidates are registered", async () => {
            let groupSize = 2
            let threshold = 2

            await keepFactory.registerMemberCandidate(application, { from: member1 })

            try {
                await keepFactory.openKeep(
                    groupSize,
                    threshold,
                    keepOwner,
                    bond,
                    { from: application }
                )

                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, "Not enough operators in pool")
            }
        })

        it("opens keep with multiple members and emits an event", async () => {
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
