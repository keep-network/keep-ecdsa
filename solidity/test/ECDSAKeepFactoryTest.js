const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory');
const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');

contract("ECDSAKeepFactory", async accounts => {
    let keepFactory

    describe.only("registerMemberCandidate", async () => {
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
        before(async () => {
            keepFactory = await ECDSAKeepFactory.new()
        })

        it("reverts if no member candidates are registered", async () => {
            keepFactory = await ECDSAKeepFactory.new()

            try {
                await keepFactory.openKeep.call(
                    10, // _groupSize
                    5, // _honestThreshold
                    "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
                )

                assert(false, 'Test call did not error as expected')
            } catch (e) {
                assert.include(e.message, "candidates list is empty")
            }
        })

        it("emits ECDSAKeepCreated event upon keep creation", async () => {
            const member = accounts[1]

            await keepFactory.registerMemberCandidate({ from: member })

            let blockNumber = await web3.eth.getBlockNumber()

            let keepAddress = await keepFactory.openKeep.call(
                10, // _groupSize
                5, // _honestThreshold
                "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
            )

            await keepFactory.openKeep(
                10, // _groupSize
                5, // _honestThreshold
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
