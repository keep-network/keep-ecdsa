const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory');
const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');

contract("ECDSAKeepFactory", async accounts => {
    let keepFactory

    describe("registerMemberCandidate", async () => {
        before(async () => {
            keepFactory = await ECDSAKeepFactoryStub.new()
        })

        it("reverts if no member candidates are registered", async () => {
            const member = accounts[1]

            await keepFactory.registerMemberCandidate({ from: member })

            const memberCandidates1 = await keepFactory.getMemberCandidates()

            assert.equal(
                memberCandidates1,
                member,
                "incorrect registered member candidates list",
            )

            await keepFactory.registerMemberCandidate({ from: member })

            const memberCandidates2 = await keepFactory.getMemberCandidates()

            assert.equal(
                memberCandidates2,
                member,
                "incorrect registered member candidates list after re-registration",
            )
        })
    })

    describe("openKeep", async () => {
        beforeEach(async () => {
            keepFactory = await ECDSAKeepFactory.new()
        })

        // it("reverts if not enough member candidates are registered", async () => {
        //     const member1 = accounts[1]
        //     await keepFactory.registerMemberCandidate({ from: member1 })

        //     try {
        //         await keepFactory.openKeep(
        //             2, // _groupSize
        //             2, // _honestThreshold
        //             "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
        //         )

        //         assert(false, 'Test call did not error as expected')
        //     } catch (e) {
        //         assert.include(e.message, "not enough member candidates registered to form a group")
        //     }
        // })

        it("opens keep with multiple members", async () => {
            const member1 = accounts[1]
            const member2 = accounts[2]
            const member3 = accounts[3]

            await keepFactory.registerMemberCandidate({ from: member1 })
            await keepFactory.registerMemberCandidate({ from: member2 })
            await keepFactory.registerMemberCandidate({ from: member3 })

            let blockNumber = await web3.eth.getBlockNumber()

            await keepFactory.openKeep(
                3, // _groupSize
                3, // _honestThreshold
                "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
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
            const member = accounts[1]

            await keepFactory.registerMemberCandidate({ from: member })

            let blockNumber = await web3.eth.getBlockNumber()

            let keepAddress = await keepFactory.openKeep.call(
                1, // _groupSize
                1, // _honestThreshold
                "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
            )

            await keepFactory.openKeep(
                1, // _groupSize
                1, // _honestThreshold
                "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
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

            assert.equal(
                eventList[0].returnValues.members,
                member,
                "incorrect keep member in emitted event",
            )
        })
    })
})
