var ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory');

contract("ECDSAKeepFactory test", async accounts => {
    it("check event", async () => {
        let blockNumber = await web3.eth.getBlock("latest").number

        let keepFactory = await ECDSAKeepFactory.deployed();

        console.log('Call createNewKeep');
        let res = await keepFactory.createNewKeep.call(
            10, // uint256 _groupSize,
            5, // uint256 _honestThreshold,
            "0xbc4862697a1099074168d54A555c4A60169c18BD" // address _owner
        );

        let eventList = await keepFactory.getPastEvents('ECDSAKeepCreated', {
            fromBlock: blockNumber,
            toBlock: 'latest'
        })

        // Just for debug print list of events
        console.log(eventList)

        assert.equal(eventList.length, 1, "incorrect number of emitted events")
        assert.equal(eventList[0].returnValues.keepAddress, res, "incorrect keep address in emitted event")
    });
});
