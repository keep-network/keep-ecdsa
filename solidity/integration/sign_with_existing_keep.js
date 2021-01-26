const BondedECDSAKeep = artifacts.require("./BondedECDSAKeep.sol")

/*
This test validates integration between on-chain contracts and off-chain client.
It requires contracts to be deployed before running the test. It requests
signature for a specific keep contract provided as a KEEP_ADDRESS environment
variable using keep owner (privileged application) provided as a KEEP_OWNER
environment variable.

To execute this test run:
  KEEP_ADDRESS=<KEEP_ADDRESS> \
  KEEP_OWNER=<KEEP_OWNER> \
  truffle exec integration/sign_with_existing_keep.js --network local
*/

module.exports = async function () {
  const keepAddress = process.env.KEEP_ADDRESS
  const keepOwnerArg = process.env.KEEP_OWNER

  let keepOwner
  let keep
  let keepPublicKey

  try {
    if (!keepOwnerArg) {
      const accounts = await web3.eth.getAccounts()
      keepOwner = accounts[5] // keep owner address from smoke_test.js
    } else {
      keepOwner = keepOwnerArg
    }

    keep = await BondedECDSAKeep.at(keepAddress)

    startBlockNumber = await web3.eth.getBlock("latest").number
  } catch (err) {
    console.error(`initialization failed: [${err}]`)
    process.exit(1)
  }

  try {
    console.log("get public key...")
    keepPublicKey = await keep.getPublicKey.call()
  } catch (err) {
    console.error(`failed to get keep public key: [${err}]`)
    process.exit(1)
  }

  try {
    console.log(`requesting signature with keep owner: ${keepOwner}`)
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

    if (
      web3.utils.toChecksumAddress(recoveredAddress) !=
      web3.utils.toChecksumAddress(keepPublicKeyAddress)
    ) {
      console.error(
        "signature validation failed, recovered address doesn't match expected\n" +
          `expected:  [${keepPublicKeyAddress}]\n` +
          `recovered: [${recoveredAddress}]`
      )
    }

    console.log(
      "received valid signature:\n" +
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
    keep.SignatureSubmitted().on("data", (event) => {
      resolve(event)
    })
  })
}

function publicKeyToAddress(publicKey) {
  const hash = web3.utils.keccak256(publicKey)
  return "0x" + hash.slice(24 + 2)
}
