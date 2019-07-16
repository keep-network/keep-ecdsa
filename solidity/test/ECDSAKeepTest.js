const ECDSAKeepFactory = artifacts.require('./ECDSAKeepFactory.sol');
const ECDSAKeep = artifacts.require('./ECDSAKeep.sol');

const truffleAssert = require('truffle-assertions');

contract('ECDSAKeep', function (accounts) {
    describe("#constructor", async function () {
        it('succeeds', async () => {
            let owner = accounts[0]
            let members = [owner];
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
            let members = [owner];
            let threshold = 1;

            instance = await ECDSAKeep.new(
                owner,
                members,
                threshold
            );
        });

        it('emits event', async () => {
            let res = await instance.sign('0x00')
            truffleAssert.eventEmitted(res, 'SignatureRequested', (ev) => {
                return ev.digest == '0x00'
            });
        })
    })

    describe("public key", () => {
        let expectedPublicKey = web3.utils.hexToBytes("0x67656e657261746564207075626c6963206b6579")
        let owner = "0xbc4862697a1099074168d54A555c4A60169c18BD";
        let members = ["0x774700a36A96037936B8666dCFdd3Fb6687b08cb"];
        let honestThreshold = 5;

        it("get public key before it is set", async () => {
            let keep = await ECDSAKeep.new(owner, members, honestThreshold);

            let publicKey = await keep.getPublicKey.call()

            assert.equal(publicKey, undefined, "incorrect public key")
        });

        it("set public key and get it", async () => {
            let keep = await ECDSAKeep.new(owner, members, honestThreshold);

            await keep.setPublicKey(expectedPublicKey)

            let publicKey = await keep.getPublicKey.call()

            assert.equal(
                publicKey,
                web3.utils.bytesToHex(expectedPublicKey),
                "incorrect public key"
            )
        });
    })
});
