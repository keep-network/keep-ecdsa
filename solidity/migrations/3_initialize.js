const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1 = artifacts.require("BondedECDSAKeepVendorImplV1")
const ECDSAKeepFactory = artifacts.require("ECDSAKeepFactory")
const Registry = artifacts.require("Registry")

let { RegistryAddress } = require('./external-contracts')

module.exports = async function (deployer) {
    const ecdsaKeepFactory = await ECDSAKeepFactory.deployed()

    let registry
    if (process.env.TEST) {
        registry = await Registry.deployed()
    } else {
        registry = await Registry.at(RegistryAddress)
    }

    const vendor = await BondedECDSAKeepVendorImplV1.at(BondedECDSAKeepVendor.address)
    await vendor.initialize(registry.address)

    // Configure registry
    await registry.approveOperatorContract(ECDSAKeepFactory.address)
    console.log(`approved operator contract [${ECDSAKeepFactory.address}] in registry`)

    // Set service contract owner as operator contract upgrader by default
    const operatorContractUpgrader = await vendor.owner()
    await registry.setOperatorContractUpgrader(vendor.address, operatorContractUpgrader)

    // Register keep factory
    await vendor.registerFactory(ECDSAKeepFactory.address)
}
