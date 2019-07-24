var TECDSAKeepFactory = artifacts.require('TECDSAKeepFactory');

contract("TECDSAKeepFactory", async accounts => {

    it("emits TECDSAKeepCreated event upon keep creation", async () => {
        let blockNumber = await web3.eth.getBlockNumber()

        let keepFactory = await TECDSAKeepFactory.deployed();

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

        let eventList = await keepFactory.getPastEvents('TECDSAKeepCreated', {
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
