const ECDSAKeepFactory = artifacts.require('./ECDSAKeepFactory.sol');
const ECDSAKeep = artifacts.require('./ECDSAKeep.sol');

const truffleAssert = require('truffle-assertions');

// ECDSAKeep
// signature request
// creates a new keep and calls .sign
module.exports = async function () {
    let accounts = await web3.eth.getAccounts();

    let factoryInstance = await ECDSAKeepFactory.deployed();

    // Deploy Keep
    let groupSize = 1;
    let honestThreshold = 1;
    let owner = accounts[0];

    let createKeepTx = await factoryInstance.openKeep(
        groupSize,
        honestThreshold,
        owner
    )

    // Get the Keep's address
    let instanceAddress;
    truffleAssert.eventEmitted(createKeepTx, 'ECDSAKeepCreated', (ev) => {
        instanceAddress = ev.keepAddress;
        return true;
    });

    expect(instanceAddress.length).to.eq(42);

    let instance = await ECDSAKeep.at(instanceAddress)

    let signTx = await instance.sign('0x00')
    truffleAssert.eventEmitted(signTx, 'SignatureRequested', (ev) => {
        return ev.digest == '0x00'
    });

    process.exit(0)
}
