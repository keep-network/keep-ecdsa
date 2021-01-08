/**
 * This script generates the input file for the staker rewards distributor for
 * accounts for the first four accounts. This script was written in order to
 * test integration between `ECDSARewardsDistributor` contract and KEEP token
 * dashboard with the real merkle roots data. Based on the generated file by
 * this script, the merkle root generator can generate merkle tree.
 */

const fs = require("fs")

const distributorInput =
  "../staker-rewards/distributor/staker-reward-allocation.json"

module.exports = async function () {
  try {
    const accounts = await web3.eth.getAccounts()
    const input = {}

    for (let i = 0; i < 5; i++) {
      const amount = web3.utils.toWei("200", "ether")
      input[accounts[i]] = web3.utils.numberToHex(amount).substring(2)
    }

    fs.writeFileSync(distributorInput, JSON.stringify(input, null, 2))
  } catch (err) {
    console.error("unexpected error:", err)
    process.exit(1)
  }

  process.exit()
}
