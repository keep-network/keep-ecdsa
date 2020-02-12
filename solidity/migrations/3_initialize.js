const ECDSAKeepFactory = artifacts.require("ECDSAKeepFactory");

module.exports = async function (deployer) {
    await ECDSAKeepFactory.deployed()

    let registry
    if (process.env.TEST) {
        RegistryStub = artifacts.require("RegistryStub")
        registry = await RegistryStub.new()
    } else {
        Registry = artifacts.require("Registry")
        registry = await Registry.deployed()
    }

    await registry.approveOperatorContract(ECDSAKeepFactory.address)
    console.log(`approved operator contract [${ECDSAKeepFactory.address}] in registry`)
};
