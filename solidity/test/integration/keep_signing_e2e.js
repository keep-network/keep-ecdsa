const { contract, web3 } = require("@openzeppelin/test-environment")

const BondedECDSAKeepFactory = contract.fromArtifact("BondedECDSAKeepFactory")
const BondedECDSAKeep = contract.fromArtifact("BondedECDSAKeep")

// Creates a new keep, requests signature for a digest and gets the signature
// submitted to the chain.
module.exports = async function () {
  try {
    const accounts = await web3.eth.getAccounts()
    const factory = await BondedECDSAKeepFactory.deployed()

    // Create new keep.
    const groupSize = 1
    const honestThreshold = 1
    const keepOwner = accounts[1]
    const bond = 1

    const openKeepTx = await factory
      .openKeep(groupSize, honestThreshold, keepOwner, bond)
      .catch((err) => {
        console.error(`failed keep creation: [${err}]`)
        process.exit(1)
      })

    const keepAddress = openKeepTx.logs[0].args.keepAddress

    console.log("new keep created at address:", keepAddress)

    const keep = await BondedECDSAKeep.at(keepAddress)

    // Sign digest.
    const digest =
      "0xca071ca92644f1f2c4ae1bf71b6032e5eff4f78f3aa632b27cbc5f84104a32da"

    await requestSignature(keep, keepOwner, digest).catch((err) => {
      console.error(`failed to sign: [${err}]`)
      process.exit(1)
    })
    process.exit(0)
  } catch (err) {
    console.error(`unexpected error: [${err}]`)
    process.exit(1)
  }
}

async function requestSignature(keep, keepOwner, digest) {
  // Register event listener to wait for an event emitted after signature is
  // submitted by an off-chain keep client.
  const eventPromise = waitForEvent(keep.SignatureSubmitted()).catch((err) => {
    throw new Error(`event watching failed: [${err}]`)
  })

  console.log("signing digest:", digest)

  await keep.sign(digest, { from: keepOwner }).catch((err) => {
    throw new Error(`failed signing: [${err}]`)
  })

  const receivedSignatureEvent = await eventPromise

  if (receivedSignatureEvent.returnValues.digest != digest) {
    throw new Error(
      `unexpected digest: ${receivedSignatureEvent.returnValues.digest}`
    )
  }

  console.log(
    `received signature:\nR: ${receivedSignatureEvent.returnValues.r}\nS: ${receivedSignatureEvent.returnValues.s}`
  )
}
