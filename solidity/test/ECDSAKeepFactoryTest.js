var ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory');

contract("ECDSAKeepFactory", async accounts => {

    it("emits ECDSAKeepCreated event upon keep creation", async () => {
        let blockNumber = await web3.eth.getBlockNumber()

        let keepFactory = await ECDSAKeepFactory.deployed();

        let keepAddress = await keepFactory.createNewKeep.call(
            10, // _groupSize,
            5, // _honestThreshold,
            "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
        ).catch((err) => {
            assert.fail(`ecdsa keep creation failed: ${err}`);
        });

        await keepFactory.createNewKeep(
            10, // _groupSize,
            5, // _honestThreshold,
            "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
        ).catch((err) => {
            assert.fail(`ecdsa keep creation failed: ${err}`);
        });

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
    });
});
