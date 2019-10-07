const KeepRegistry = artifacts.require('./KeepRegistry.sol')
const ECDSAKeepVendor = artifacts.require('./ECDSAKeepVendor.sol')
const ECDSAKeepFactory = artifacts.require('./ECDSAKeepFactory.sol')
const ECDSAKeep = artifacts.require('./ECDSAKeep.sol')

const waitForEvent = require('../test/helpers/waitForEvent')

// This test validates integration between on-chain contracts and off-chain client.
// It requires contracts to be deployed before running the test and the client
// to be configured with `ECDSAKeepFactory` address.
// 
// To execute this smoke test run:
// truffle exec integration/smoke_test.js
module.exports = async function () {
    let keepRegistry
    let keepFactory
    let keepOwner
    let keep
    let keepPublicKey

    try {
        keepRegistry = await KeepRegistry.deployed()
        keepFactory = await ECDSAKeepFactory.deployed()

        const accounts = await web3.eth.getAccounts();
        keepOwner = accounts[1]

        startBlockNumber = await web3.eth.getBlock('latest').number
    } catch (err) {
        console.error(`initialization failed: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('open new keep...')
        const keepVendorAddress = await keepRegistry.getVendor.call("ECDSAKeep")
        const keepVendor = await ECDSAKeepVendor.at(keepVendorAddress)


        const keepCreatedEvent = watchKeepCreatedEvent(keepFactory)

        await keepVendor.openKeep(
            10,
            5,
            keepOwner
        )

        const keepCreated = (await keepCreatedEvent).returnValues

        const keepAddress = keepCreated.keepAddress
        keep = await ECDSAKeep.at(keepAddress)

        console.log(`new keep opened with address: [${keepAddress}]`)
    } catch (err) {
        console.error(`failed to open new keep: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('get public key...')
        await watchPublicKeyPublishedEvent(keep)
            .then(
                // onFulfilled - get public key from emitted event
                async (event) => {
                    keepPublicKey = event.returnValues.publicKey
                },
                // onRejected - as fallback check if keep already has public key
                async () => {
                    keepPublicKey = await keep.getPublicKey()
                }
            )

        console.log(`public key generated for keep: [${keepPublicKey}]`)
    } catch (err) {
        console.error(`failed to get keep public key: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('request signature...')
        const digest = web3.eth.accounts.hashMessage("hello")
        const signatureSubmittedEvent = watchSignatureSubmittedEvent(keep)

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

function watchKeepCreatedEvent(factory, timeout = 1000) {
    return waitForEvent(factory.ECDSAKeepCreated(), timeout)
}

function watchPublicKeyPublishedEvent(keep, timeout = 1000) {
    return waitForEvent(keep.PublicKeyPublished(), timeout)
}

function watchSignatureSubmittedEvent(keep, timeout = 1000) {
    return waitForEvent(keep.SignatureSubmitted(), timeout)
}

function publicKeyToAddress(publicKey) {
    const hash = web3.utils.keccak256(publicKey)
    return "0x" + hash.slice(24 + 2)
}
