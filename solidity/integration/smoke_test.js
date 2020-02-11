const BondedECDSAKeepVendor = artifacts.require('./BondedECDSAKeepVendor.sol')
const BondedECDSAKeepVendorImplV1 = artifacts.require('./BondedECDSAKeepVendorImplV1.sol')
const ECDSAKeepFactory = artifacts.require('./ECDSAKeepFactory.sol')
const ECDSAKeep = artifacts.require('./ECDSAKeep.sol')

// This test validates integration between on-chain contracts and off-chain client.
// It requires contracts to be deployed before running the test and the client
// to be configured with `ECDSAKeepFactory` address.
// 
// To execute this smoke test run:
// truffle exec integration/smoke_test.js
module.exports = async function () {
    let keepOwner
    let application
    let startBlockNumber
    let keep
    let keepPublicKey

    const groupSize = 3
    const threshold = 3
    const bond = 10

    try {
        const accounts = await web3.eth.getAccounts();
        keepOwner = accounts[1]
        application = "0x72e81c70670F0F89c1e3E8a29409157BC321B107"

        startBlockNumber = await web3.eth.getBlock('latest').number
    } catch (err) {
        console.error(`initialization failed: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('opening a new keep...');

        const keepVendor = await BondedECDSAKeepVendorImplV1.at(
            (await BondedECDSAKeepVendor.deployed()).address
        )
        const keepFactoryAddress = await keepVendor.selectFactory()
        keepFactory = await ECDSAKeepFactory.at(keepFactoryAddress)
        await keepFactory.openKeep(
            groupSize,
            threshold,
            keepOwner,
            bond,
        )

        const eventList = await keepFactory.getPastEvents('ECDSAKeepCreated', {
            fromBlock: startBlockNumber,
            toBlock: 'latest',
        })

        const keepAddress = eventList[0].returnValues.keepAddress
        keep = await ECDSAKeep.at(keepAddress)

        console.log(`new keep opened with address: [${keepAddress}] and members: [${eventList[0].returnValues.members}]`)
    } catch (err) {
        console.error(`failed to open new keep: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('get public key...')
        const publicKeyPublishedEvent = await watchPublicKeyPublished(keep)

        keepPublicKey = publicKeyPublishedEvent.returnValues.publicKey

        console.log(`public key generated for keep: [${keepPublicKey}]`)
    } catch (err) {
        console.error(`failed to get keep public key: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('request signature...')
        const digest = web3.eth.accounts.hashMessage("hello")
        const signatureSubmittedEvent = watchSignatureSubmittedEvent(keep)

        setTimeout(
            async () => {
                await keep.sign(digest, { from: keepOwner })
            },
            2000
        )

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

function watchPublicKeyPublished(keep) {
    return new Promise(async (resolve) => {
        keep.PublicKeyPublished()
            .on('data', event => {
                resolve(event)
            })
    })
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
