const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const Registry = artifacts.require("./Registry.sol");

module.exports = async function (deployer) {
    await ECDSAKeepFactory.deployed()
    const registry = await Registry.deployed();

    await registry.approveOperatorContract(ECDSAKeepFactory.address)
    console.log(`approved operator contract [${ECDSAKeepFactory.address}] in registry`)
};
