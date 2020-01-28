const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory');
const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');
const SortitionPoolFactoryStub = artifacts.require('SortitionPoolFactoryStub');
const SortitionPoolStub = artifacts.require('SortitionPoolStub');

contract("ECDSAKeepFactory", async accounts => {
    let keepFactory
    let sortitionPoolFactory

    const application = '0x0000000000000000000000000000000000000001'

    describe("registerMemberCandidate", async () => {
        beforeEach(async () => {
            sortitionPoolFactory = await SortitionPoolFactoryStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(sortitionPoolFactory.address)
        })

        it("creates a signer pool", async () => {
            const member = accounts[1]

            await keepFactory.registerMemberCandidate(application, { from: member })

            const signerPoolAddress = await keepFactory.getSignerPool(application)

            assert.notEqual(
                signerPoolAddress,
                "0x0000000000000000000000000000000000000000",
                "incorrect registered signer pool",
            )
        })

        it("inserts operators to the same pool", async () => {
            const member1 = accounts[1]
            const member2 = accounts[2]

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

        it("inserts operators to different pools", async () => {
            const member1 = accounts[1]
            const member2 = accounts[2]

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

        beforeEach(async () => {
            sortitionPoolFactory = await SortitionPoolFactoryStub.new()
            keepFactory = await ECDSAKeepFactory.new(sortitionPoolFactory.address)
        })

        it("reverts if no member candidates are registered", async () => {
            keepFactory = await ECDSAKeepFactory.new(sortitionPoolFactory.address)

            try {
                await keepFactory.openKeep(
                    10, // _groupSize
                    5, // _honestThreshold
                    keepOwner // _owner
                )

                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, "No signer pool for this application")
            }
        })

        it("opens keep with multiple members", async () => {
            const application = accounts[1]
            const member1 = accounts[2]
            const member2 = accounts[3]
            const member3 = accounts[4]

            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })
            await keepFactory.registerMemberCandidate(application, { from: member3 })

            let blockNumber = await web3.eth.getBlockNumber()

            await keepFactory.openKeep(
                3, // _groupSize
                3, // _honestThreshold
                "0xbc4862697a1099074168d54A555c4A60169c18BD", // _owner,
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

        it("emits ECDSAKeepCreated event upon keep creation", async () => {
            const application = accounts[1]
            const member1 = accounts[2]
            const member2 = accounts[3]

            await keepFactory.registerMemberCandidate(application, { from: member1 })
            await keepFactory.registerMemberCandidate(application, { from: member2 })

            let blockNumber = await web3.eth.getBlockNumber()

            let keepAddress = await keepFactory.openKeep.call(
                3, // _groupSize
                2, // _honestThreshold
                keepOwner, // _owner
                { from: application }
            )

            await keepFactory.openKeep(
                3, // _groupSize
                2, // _honestThreshold
                keepOwner, // _owner
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

            assert.deepEqual(
                eventList[0].returnValues.members,
                [member1, member2, member1],
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
