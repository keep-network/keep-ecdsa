const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory')
const RandomBeaconService = artifacts.require('IRandomBeaconService')

const { RandomBeaconServiceAddress } = require('../migrations/externals')

const BN = web3.utils.BN
const maxUINT256 = new BN('2').pow(new BN('256')).sub(new BN('1'))

// This scripts checks gas required to call `setGroupSelectionSeed` function.
// The function will be used by the random beacon in a callback setting the new
// random entry.
// To run this script successfully we need to tweak the `onlyRandomBeacon` modifier
// to require specific account address (`entrySubmitterAddress`) instead of the
// random beacon contract address.
// 
// To execute this smoke test run:
// truffle exec integration/estimate_callback_gas.js
module.exports = async function () {
    let randomBeaconService
    let accounts

    try {
        accounts = await web3.eth.getAccounts();

        randomBeaconService = await RandomBeaconService.at(RandomBeaconServiceAddress)
        keepFactory = await ECDSAKeepFactory.deployed()
    } catch (err) {
        console.error(`initialization failed: [${err}]`)
        process.exit(1)
    }

    let gasEstimate
    try {
        // Update contract to use this address in `onlyRandomBeacon` modifier require
        // statement.
        const entrySubmitterAddress = accounts[1]
        console.log("entrySubmitterAddress", entrySubmitterAddress)

        // Set to min value
        await keepFactory.setGroupSelectionSeed(0, { from: entrySubmitterAddress })

        // Update to max value
        tx = await keepFactory.setGroupSelectionSeed(maxUINT256, { from: entrySubmitterAddress })
        gasEstimate = tx.receipt.gasUsed
        console.log("gasEstimate", tx.receipt.gasUsed)
    } catch (err) {
        console.error(`failed to estimate gas for callback [${err}]`)
        process.exit(1)
    }

    try {
        const feeEstimate = await randomBeaconService.entryFeeEstimate(gasEstimate)
        console.log("feeEstimate", feeEstimate.toString())
    } catch (err) {
        console.error(`failed to estimate fee [${err}]`)
        process.exit(1)
    }

    // Received result:
    // gasEstimate 41830
    // feeEstimate 11516250000000000

    process.exit(0)
}
