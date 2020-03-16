const BondedECDSAKeepVendor = artifacts.require("BondedECDSAKeepVendor")
const BondedECDSAKeepVendorImplV1 = artifacts.require("BondedECDSAKeepVendorImplV1")
const BondedECDSAKeepFactory = artifacts.require("BondedECDSAKeepFactory")
const Registry = artifacts.require("Registry")

let { RegistryAddress } = require('./external-contracts')

module.exports = async function (deployer) {
    await BondedECDSAKeepFactory.deployed()

    let registry
    if (process.env.TEST) {
        registry = await Registry.deployed()
    } else {
        registry = await Registry.at(RegistryAddress)
    }

    const proxy = await BondedECDSAKeepVendor.deployed()
    const vendor = await BondedECDSAKeepVendorImplV1.at(proxy.address)

    // Initialize vendor contract
    if (!(await vendor.initialized())) {
        throw Error("vendor contract not initialized")
    }

    const factoryAddress = await vendor.selectFactory()
    console.log(`current factory address: [${factoryAddress}]`)

    // Configure registry
    await registry.approveOperatorContract(BondedECDSAKeepFactory.address)
    console.log(`approved operator contract [${BondedECDSAKeepFactory.address}] in registry`)

    // Set service contract owner as operator contract upgrader by default
    const operatorContractUpgrader = await proxy.owner()
    await registry.setOperatorContractUpgrader(vendor.address, operatorContractUpgrader)
    console.log(`set operator [${operatorContractUpgrader}] as [${vendor.address}] contract upgrader`)
}
