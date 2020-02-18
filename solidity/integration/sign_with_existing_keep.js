const BondedECDSAKeep = artifacts.require('./BondedECDSAKeep.sol')

// This test validates integration between on-chain contracts and off-chain client.
// It requires contracts to be deployed before running the test. It requests
// signature for a specific keep contract provided as a <KEEP_ADDRESS> argument.
// 
// To execute this test run:
// truffle exec integration/sign_with_existing_keep.js <KEEP_ADDRESS>
module.exports = async function () {
    const keepAddress = process.argv[4]
    const keepOwnerArg = process.argv[5]

    let keepOwner
    let keep
    let keepPublicKey

    try {
        if (!keepOwnerArg) {
            const accounts = await web3.eth.getAccounts();
            keepOwner = accounts[1]
        }

        keep = await BondedECDSAKeep.at(keepAddress)

        startBlockNumber = await web3.eth.getBlock('latest').number
    } catch (err) {
        console.error(`initialization failed: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('get public key...')
        keepPublicKey = await keep.getPublicKey.call()

    } catch (err) {
        console.error(`failed to get keep public key: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('request signature...')
        const signatureSubmittedEvent = watchSignatureSubmittedEvent(keep)

        const digest = web3.eth.accounts.hashMessage("hello")

        await keep.sign(digest, { from: keepOwner })

        const signature = (await signatureSubmittedEvent).returnValues

        const v = web3.utils.toHex(27 + Number(signature.recoveryID))

        const recoveredAddress = web3.eth.accounts.recover(
            digest,
            v,
            signature.r,
            signature.s,
            true
        )

        const keepPublicKeyAddress = publicKeyToAddress(keepPublicKey)

        if (web3.utils.toChecksumAddress(recoveredAddress)
            != web3.utils.toChecksumAddress(keepPublicKeyAddress)) {
            console.error(
                'signature validation failed, recovered address doesn\'t match expected\n' +
                `expected:  [${keepPublicKeyAddress}]\n` +
                `recovered: [${recoveredAddress}]`
            )
        }

        console.log(
            'received valid signature:\n' +
            `r: [${signature.r}]\n` +
            `s: [${signature.s}]\n` +
            `recoveryID: [${signature.recoveryID}]\n`
        )
    } catch (err) {
        console.error(`signing failed: [${err}]`)
        process.exit(1)
    }

    process.exit()
}

function watchSignatureSubmittedEvent(keep) {
    return new Promise(async (resolve) => {
        keep.SignatureSubmitted()
            .on('data', event => {
                resolve(event)
            })
    })
}

function publicKeyToAddress(publicKey) {
    const hash = web3.utils.keccak256(publicKey)
    return "0x" + hash.slice(24 + 2)
}
