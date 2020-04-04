const BondedECDSAKeepVendor = artifacts.require("./BondedECDSAKeepVendor.sol")
const BondedECDSAKeepVendorImplV1 = artifacts.require(
  "./BondedECDSAKeepVendorImplV1.sol"
)
const BondedECDSAKeepFactory = artifacts.require("./BondedECDSAKeepFactory.sol")
const BondedECDSAKeep = artifacts.require("./BondedECDSAKeep.sol")

// how many iterations
const iterations = 10
// delay between the iterations [ms]
const delay = 30000
// number of keeps created at once
const stressLevel = 5

const groupSize = 3
const threshold = 3
const bond = 10

module.exports = async function () {
  const accounts = await web3.eth.getAccounts()

  const keepOwner = accounts[4]
  const application = accounts[5]

  let keepFactory

  try {
    const keepVendor = await BondedECDSAKeepVendorImplV1.at(
      (await BondedECDSAKeepVendor.deployed()).address
    )
    const keepFactoryAddress = await keepVendor.selectFactory()
    keepFactory = await BondedECDSAKeepFactory.at(keepFactoryAddress)
  } catch (err) {
    console.error(`failed to select a factory: [${err}]`)
    process.exit(1)
  }

  let generatedKeys = 0

  try {
    keepFactory.BondedECDSAKeepCreated(async (_, event) => {
      const keepAddress = event.returnValues.keepAddress
      keep = await BondedECDSAKeep.at(keepAddress)
      console.log(
        `new keep created: [${keepAddress}] at [${new Date().toLocaleString()}]`
      )

      const publicKeyPublishedEvent = await watchPublicKeyPublished(keep)
      keepPublicKey = publicKeyPublishedEvent.returnValues.publicKey
      console.log(
        `public key generated for keep [${keepAddress}] at [${new Date().toLocaleString()}]: [${keepPublicKey}]`
      )

      generatedKeys++
      console.log(`generated [${generatedKeys}] public keys so far`)
    })

    for (let i = 0; i < iterations; i++) {
      promises = []
      for (let j = 0; j < stressLevel; j++) {
        promises.push(openKeep(keepFactory, keepOwner, application))
      }

      await Promise.all(promises)
      await wait(delay)
    }
  } catch (err) {
    console.error(`unexpected failure: [${err}]`)
    process.exit(1)
  }
}

function openKeep(keepFactory, keepOwner, application) {
  return new Promise(async (resolve) => {
    const fee = await keepFactory.openKeepFeeEstimate.call()

    console.log(`opening a new keep at [${new Date().toLocaleString()}]...`)
    await keepFactory.openKeep(groupSize, threshold, keepOwner, bond, {
      from: application,
      value: fee,
    })
    resolve()
  })
}

function watchPublicKeyPublished(keep) {
  return new Promise(async (resolve) => {
    keep.PublicKeyPublished().on("data", (event) => {
      resolve(event)
    })
  })
}

function wait(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
