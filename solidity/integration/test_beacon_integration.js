const ECDSAKeepFactoryExposed = artifacts.require('ECDSAKeepFactoryExposed')
const RandomBeaconService = artifacts.require('KeepRandomBeaconServiceImplV1')
const { RandomBeaconServiceAddress } = require('../migrations/externals')

// This scripts validates integration with the random beacon service.
// It checks that the beacon executes the callback function and sets a new
// group selection seed on the factory contract.
//
// It requires usage of `ECDSAKeepFactoryExposed` contract which exposes function
// to get internal `groupSelectionSeed`. To do that the contract has to be deployed
// instead of regular `ECDSAKeepFactory` contract.
// 
// To execute this smoke test run:
// truffle exec integration/test_beacon_integration.js
module.exports = async function () {
    const application = "0x2AA420Af8CB62888ACBD8C7fAd6B4DdcDD89BC82"

    let keepOwner
    let randomBeacon

    const groupSize = 3
    const threshold = 3
    const bond = 10

    try {
        const accounts = await web3.eth.getAccounts();
        keepOwner = accounts[1]
    } catch (err) {
        console.error(`initialization failed: [${err}]`)
        process.exit(1)
    }


    try {
        keepFactory = await ECDSAKeepFactoryExposed.deployed()
        randomBeacon = await RandomBeaconService.at(RandomBeaconServiceAddress)
    } catch (err) {
        console.error(`factory deployment failed: [${err}]`)
        process.exit(1)
    }

    try {
        console.log('opening a new keep...');

        const fee = await keepFactory.openKeepFeeEstimate.call()
        console.log(`open new keep fee: [${fee}]`)

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
    } catch (err) {
        console.error(`failed to open new keep: [${err}]`)
        process.exit(1)
    }

    let newRelayEntry
    try {
        console.log('wait for new relay entry...')
        const event = await watchRelayEntryGenerated(randomBeacon)

        newRelayEntry = web3.utils.toBN(event.returnValues.entry)

        console.log(`new relay entry: [${newRelayEntry}]`)
    } catch (err) {
        console.error(`failed to get relay entry: [${err}]`)
        process.exit(1)
    }

    let currentSeed
    try {
        console.log('get current group selection seed...')
        currentSeed = web3.utils.toBN(await keepFactory.getGroupSelectionSeed())

        console.log(`current seed: [${currentSeed}]`)
    } catch (err) {
        console.error(`failed to get current seed: [${err}]`)
        process.exit(1)
    }

    try {
        if (currentSeed.cmp(newRelayEntry) != 0) {
            console.error(
                `current seed is not equal new relay entry\nactual:   ${currentSeed}\nexpected: ${newRelayEntry}`
            )
        }
    } catch (err) {
        console.error(`failed to compare results: [${err}]`)
        process.exit(1)
    }

    process.exit()
}

function watchRelayEntryGenerated(randomBeacon) {
    return new Promise(async (resolve) => {
        randomBeacon.RelayEntryGenerated()
            .on('data', event => {
                resolve(event)
            })
    })
}
