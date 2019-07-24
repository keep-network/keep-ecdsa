const TECDSAKeepFactory = artifacts.require('./TECDSAKeepFactory.sol');
const TECDSAKeep = artifacts.require('./TECDSAKeep.sol');

const truffleAssert = require('truffle-assertions');

// TECDSAKeep
// signature request
// creates a new keep and calls .sign
module.exports = async function () {
    let accounts = await web3.eth.getAccounts();

    let factoryInstance = await TECDSAKeepFactory.deployed();

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
    truffleAssert.eventEmitted(createKeepTx, 'TECDSAKeepCreated', (ev) => {
        instanceAddress = ev.keepAddress;
        return true;
    });

    expect(instanceAddress.length).to.eq(42);

    let instance = await TECDSAKeep.at(instanceAddress)

    let signTx = await instance.sign('0x00')
    truffleAssert.eventEmitted(signTx, 'SignatureRequested', (ev) => {
        return ev.digest == '0x00'
    });

    process.exit(0)
}
