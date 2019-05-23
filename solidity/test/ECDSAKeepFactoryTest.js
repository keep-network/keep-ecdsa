var ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory');

contract("ECDSAKeepFactory test", async accounts => {

    it("ECDSAKeepCreated event emission", async () => {
        let blockNumber = await web3.eth.getBlock("latest").number

        let keepFactory = await ECDSAKeepFactory.deployed();

        let keepAddress = await keepFactory.createNewKeep.call(
            10, // _groupSize,
            5, // _honestThreshold,
            "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
        ).catch((err) => {
            console.log(`ecdsa keep creation for failed: ${err}`);
        });

        await keepFactory.createNewKeep(
            10, // _groupSize,
            5, // _honestThreshold,
            "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
        ).catch((err) => {
            console.log(`ecdsa keep creation failed: ${err}`);
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
