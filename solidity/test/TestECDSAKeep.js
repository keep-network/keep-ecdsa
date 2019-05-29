const ECDSAKeepFactory = artifacts.require('./ECDSAKeepFactory.sol');
const ECDSAKeep = artifacts.require('./ECDSAKeep.sol');

const truffleAssert = require('truffle-assertions');

contract('ECDSAKeep', function(accounts) {
  describe("#constructor", async function() {
    it('succeeds', async () => {
        let owner = accounts[0]
        let members = [ owner ];
        let threshold = 1;

        let instance = await ECDSAKeep.new(
            owner,
            members,
            threshold
        );

        expect(instance.address).to.be.not.empty;
    })
  });

  describe('#sign', async function () {
    let instance;

    before(async () => {
        let owner = accounts[0]
        let members = [ owner ];
        let threshold = 1;

        instance = await ECDSAKeep.new(
            owner,
            members,
            threshold
        );
    });

    it('emits event', async () => {
        let res = await instance.sign('0x00')
        truffleAssert.eventEmitted(res, 'SignatureRequest', (ev) => {
            return ev.digest == '0x00'
        });
    })
  })

});