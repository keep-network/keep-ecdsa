const ECDSAKeepFactory = artifacts.require('./ECDSAKeepFactory.sol');
const ECDSAKeep = artifacts.require('./ECDSAKeep.sol');

// ECDSAKeep signing.
// Creates a new keep, requests signature for a digest and gets the signature
// submitted to the chain.
module.exports = async function () {
    let accounts = await web3.eth.getAccounts();
    let factory = await ECDSAKeepFactory.deployed();

    // Create new keep
    let groupSize = 1;
    let honestThreshold = 1;
    let owner = accounts[1];

    const startBlockNumber = await web3.eth.getBlock('latest').number

    let openKeepTx = await factory.openKeep(
        groupSize,
        honestThreshold,
        owner
    ).catch(err => {
        console.error(`failed keep creation: [${err}]`)
        process.exit(1)
    })

    const keepAddress = openKeepTx.logs[0].args.keepAddress

    console.log("New keep created with address:", keepAddress)

    const keep = await ECDSAKeep.at(keepAddress)

    // Sign digest
    const digest = '0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da'
    console.log('Sign digest:', digest)

    let signTx = await keep.sign(digest, { from: owner })
        .catch(err => {
            console.error(`failed signing: [${err}]`)
            process.exit(1)
        })

    // Give off-chain client some time to calculate and submit a signature.
    await sleep(1000);

    // Get signature submitted by off-chain keep client
    eventList = await keep.getPastEvents('SignatureSubmitted', {
        fromBlock: startBlockNumber,
        toBlock: 'latest',
    })

    if (eventList[0].returnValues.digest != digest) {
        console.error(`unexpected digest: ${eventList[0].returnValues.digest}`)
        process.exit(1)
    }
    // TODO: Validate signature.

    console.log(`Received signature:\nR: ${eventList[0].returnValues.r}\nS: ${eventList[0].returnValues.s}`)

    process.exit(0)
}

function sleep(ms) {
    return new Promise(resolve => {
        setTimeout(resolve, ms)
    })
}
