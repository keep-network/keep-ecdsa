const ECDSAKeepFactory = artifacts.require('./ECDSAKeepFactory.sol');
const ECDSAKeep = artifacts.require('./ECDSAKeep.sol');

const truffleAssert = require('truffle-assertions');

require('chai')
    .expect();

contract('ECDSAKeep', function(accounts) {
    describe("signature request", async function() {
        it('creates a new keep and calls .sign', async () => {
            let factoryInstance = await ECDSAKeepFactory.deployed();

            // Deploy Keep
            let groupSize = 1;
            let honestThreshold = 1;
            let owner = accounts[0];

            let createKeepTx = await factoryInstance.createNewKeep(
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

            expect(instanceAddress).to.not.be.empty;

            let instance = await ECDSAKeep.at(instanceAddress)

            let signTx = await instance.sign('0x00')
            truffleAssert.eventEmitted(signTx, 'SignatureRequested', (ev) => {
                return ev.digest == '0x00'
            });
        })
    });
});