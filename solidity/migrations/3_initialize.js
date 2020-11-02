const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1 = artifacts.require(
  "BondedECDSAKeepVendorImplV1"
)
const BondedECDSAKeepFactory = artifacts.require("BondedECDSAKeepFactory")
const FullyBackedECDSAKeepFactory = artifacts.require(
  "FullyBackedECDSAKeepFactory"
)
const KeepRegistry = artifacts.require("KeepRegistry")

const {RegistryAddress} = require("./external-contracts")

module.exports = async function (deployer) {
  let registry
  if (process.env.TEST) {
    registry = await KeepRegistry.deployed()
  } else {
    registry = await KeepRegistry.at(RegistryAddress)
  }

  const proxy = await BondedECDSAKeepVendor.deployed()
  const vendor = await BondedECDSAKeepVendorImplV1.at(proxy.address)

  // The vendor implementation is being initialized as part of the vendor proxy
  // deployment. Here we just want to sanity check if it is initialized.
  if (!(await vendor.initialized())) {
    throw Error("vendor contract not initialized")
  }

  const factoryAddress = await vendor.selectFactory()
  console.log(`current factory address: [${factoryAddress}]`)

  // Configure registry
  await registry.approveOperatorContract(BondedECDSAKeepFactory.address)
  console.log(
    `approved BondedECDSAKeepFactory operator contract [${BondedECDSAKeepFactory.address}] in registry`
  )

  await registry.approveOperatorContract(FullyBackedECDSAKeepFactory.address)
  console.log(
    `approved FullyBackedECDSAKeepFactory operator contract [${FullyBackedECDSAKeepFactory.address}] in registry`
  )

  // Set service contract owner as operator contract upgrader by default
  const operatorContractUpgrader = await proxy.admin()
  await registry.setOperatorContractUpgrader(
    vendor.address,
    operatorContractUpgrader
  )
  console.log(
    `set operator [${operatorContractUpgrader}] as [${vendor.address}] contract upgrader`
  )
}
