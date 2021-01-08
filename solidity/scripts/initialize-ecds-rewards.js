// This script allocates the fake rewards for the given operator in the given
// interval via stub contract. It helps test ECDSA rewards in a KEEP token
// dashboard. NOTE: the `ECDSARewardsStub` contract must be deployed to network.

const ECDSARewards = artifacts.require("./test/ECDSARewardsStub")
const KeepToken = artifacts.require(
  "@keep-network/keep-core/build/truffle/KeepToken"
)
const {KeepTokenAddress} = require("../migrations/external-contracts")

module.exports = async function () {
  try {
    const accounts = await web3.eth.getAccounts()
    const keepToken = await KeepToken.at(KeepTokenAddress)
    const rewardsContract = await ECDSARewards.deployed()

    const owner = accounts[0]
    const totalRewards = web3.utils.toWei("180000000", "ether")

    await keepToken.approveAndCall(
      rewardsContract.address,
      totalRewards,
      "0x0",
      {
        from: owner,
      }
    )
    await rewardsContract.markAsFunded({from: owner})

    // Fake rewards allocation
    for (let i = 1; i < 5; i++) {
      const operator = accounts[i]
      for (let interval = 0; interval <= 5; interval++) {
        await rewardsContract.allocateReward(
          operator,
          interval,
          web3.utils.toWei("3000"),
          {from: owner}
        )
      }
    }
  } catch (err) {
    console.error("unexpected error:", err)
    process.exit(1)
  }

  process.exit()
}
