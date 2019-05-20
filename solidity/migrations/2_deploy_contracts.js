const ECDSAKeepFactory = artifacts.require("./ECDSAKeepFactory.sol");
const KeepRegistry = artifacts.require("./KeepRegistry.sol");

module.exports = async function (deployer) {
    await deployer.deploy(ECDSAKeepFactory);
    await deployer.deploy(KeepRegistry, ECDSAKeepFactory.address);
};
