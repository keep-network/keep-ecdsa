var ECDSAKeep = artifacts.require('ECDSAKeep');

contract("ECDSAKeep test", async accounts => {
    let expectedPublicKey = web3.utils.hexToBytes("0x67656e657261746564207075626c6963206b6579")
    let owner = "0xbc4862697a1099074168d54A555c4A60169c18BD";
    let members = ["0x774700a36A96037936B8666dCFdd3Fb6687b08cb"];
    let honestThreshold = 5;

    it("get public key before it is set", async () => {
        let keep = await ECDSAKeep.new(owner, members, honestThreshold);

        let publicKey = await keep.getPublicKey.call().catch((err) => {
            assert.fail(`ecdsa keep creation failed: ${err}`);
        });

        assert.equal(publicKey, undefined, "incorrect public key")
    });

    it("set public key and get it", async () => {
        let keep = await ECDSAKeep.new(owner, members, honestThreshold);

        await keep.setPublicKey(expectedPublicKey).catch((err) => {
            assert.fail(`ecdsa keep creation failed: ${err}`);
        });

        publicKey = await keep.getPublicKey.call().catch((err) => {
            assert.fail(`cannot get public key: ${err}`);
        });

        assert.equal(
            publicKey,
            web3.utils.bytesToHex(expectedPublicKey),
            "incorrect public key"
        )
    });
});
