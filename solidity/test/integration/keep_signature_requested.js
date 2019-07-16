const ECDSAKeepFactory = artifacts.require('./ECDSAKeepFactory.sol')
const ECDSAKeep = artifacts.require('./ECDSAKeep.sol')

const truffleAssert = require('truffle-assertions')

// ECDSAKeep
// signature request
// creates a new keep and calls .sign
module.exports = async function() {
  const accounts = await web3.eth.getAccounts()

  const factoryInstance = await ECDSAKeepFactory.deployed()

  // Deploy Keep
  const groupSize = 1
  const honestThreshold = 1
  const owner = accounts[0]

  const createKeepTx = await factoryInstance.createNewKeep(
    groupSize,
    honestThreshold,
    owner
  )

  // Get the Keep's address
  let instanceAddress
  truffleAssert.eventEmitted(createKeepTx, 'ECDSAKeepCreated', (ev) => {
    instanceAddress = ev.keepAddress
    return true
  })

  expect(instanceAddress.length).to.eq(42)

  const instance = await ECDSAKeep.at(instanceAddress)

  const signTx = await instance.sign('0x00')
  truffleAssert.eventEmitted(signTx, 'SignatureRequested', (ev) => {
    return ev.digest == '0x00'
  })

  process.exit(0)
}
