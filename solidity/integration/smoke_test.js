const BondedECDSAKeepVendor = artifacts.require('./BondedECDSAKeepVendor.sol')
const BondedECDSAKeepVendorImplV1 = artifacts.require('./BondedECDSAKeepVendorImplV1.sol')
const BondedECDSAKeepFactory = artifacts.require('./BondedECDSAKeepFactory.sol')
const BondedECDSAKeep = artifacts.require('./BondedECDSAKeep.sol')

const RandomBeaconService = artifacts.require('IRandomBeacon')
const { RandomBeaconAddress } = require('../migrations/external-contracts')

// This test validates integration between on-chain contracts and off-chain client.
// It also validates integration with the random beacon by verifying update of
// the group selection seed by random beacon with a callback.
// It requires contracts to be deployed before running the test and the client
// to be configured with `BondedECDSAKeepFactory` address. It also requires random
// beacon to be running.
// 
// To execute this smoke test run:
// truffle exec integration/smoke_test.js
module.exports = async function () {
    const accounts = await web3.eth.getAccounts();

    // accounts 1, 2 and 3 correspond to three members.
    const keepOwner = accounts[4]
    const application = accounts[5]

    let startBlockNumber
    let keep
    let keepPublicKey
    let relayEntryGeneratedWatcher

    const groupSize = 3
    const threshold = 3
    const bond = 10

    try {
        startBlockNumber = await web3.eth.getBlock('latest').number

        randomBeacon = await RandomBeaconService.at(RandomBeaconAddress)
    } catch (err) {
        console.error(`initialization failed: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('selecting a keep factory...');

        const keepVendor = await BondedECDSAKeepVendorImplV1.at(
            (await BondedECDSAKeepVendor.deployed()).address
        )
        const keepFactoryAddress = await keepVendor.selectFactory()
        keepFactory = await BondedECDSAKeepFactory.at(keepFactoryAddress)
    } catch (err) {
        console.error(`failed to select a factory: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('opening a new keep...');

        const fee = await keepFactory.openKeepFeeEstimate.call()
        console.log(`open new keep fee: [${fee}]`)

        // Initialize relay entry generated watcher.
        relayEntryGeneratedWatcher = watchRelayEntryGenerated(randomBeacon)

        await keepFactory.openKeep(
            groupSize,
            threshold,
            keepOwner,
            bond,
            {
                from: application,
                value: fee
            }
        )

        const eventList = await keepFactory.getPastEvents('BondedECDSAKeepCreated', {
            fromBlock: startBlockNumber,
            toBlock: 'latest',
        })

        const keepAddress = eventList[0].returnValues.keepAddress
        keep = await BondedECDSAKeep.at(keepAddress)

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

    try {
        console.log('wait for new relay entry generation...')

        const event = await relayEntryGeneratedWatcher
        const newRelayEntry = web3.utils.toBN(event.returnValues.entry)

        console.log('get current group selection seed...')
        const currentSeed = web3.utils.toBN(await keepFactory.groupSelectionSeed.call())

        if (currentSeed.cmp(newRelayEntry) != 0) {
            throw Error(
                'current seed does not equal new relay entry\n' +
                `actual:   ${currentSeed}\n` +
                `expected: ${newRelayEntry}`
            )
        }

        console.log('group selection seed was successfully updated by the random beacon')
    } catch (err) {
        console.error(`random beacon callback failed: [${err}]`)
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

function watchRelayEntryGenerated(randomBeacon) {
    return new Promise(async (resolve) => {
        randomBeacon.RelayEntryGenerated()
            .on('data', event => {
                resolve(event)
            })
    })
}

function publicKeyToAddress(publicKey) {
    const hash = web3.utils.keccak256(publicKey)
    return "0x" + hash.slice(24 + 2)
}
