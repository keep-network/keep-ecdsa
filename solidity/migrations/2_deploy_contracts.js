const KeepTECDSAGroup = artifacts.require("./KeepTECDSAGroup.sol");

module.exports = async function (deployer) {
  await deployer.deploy(KeepTECDSAGroup);
};
