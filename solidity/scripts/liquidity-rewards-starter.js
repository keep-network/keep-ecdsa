const LPRewardsTBTCETH = artifacts.require("LPRewardsTBTCETH")
const LPRewardsKEEPETH = artifacts.require("LPRewardsKEEPETH")
const LPRewardsKEEPTBTC = artifacts.require("LPRewardsKEEPTBTC")
const LPRewardsTBTCSaddle = artifacts.require("LPRewardsTBTCSaddle")
const LPRewardsTBTCv2Saddle = artifacts.require("LPRewardsTBTCv2Saddle")
const LPRewardsTBTCv2SaddleV2 = artifacts.require("LPRewardsTBTCv2SaddleV2")

const TestToken = artifacts.require("./test/TestToken")
const KeepToken = artifacts.require(
  "@keep-network/keep-core/build/truffle/KeepToken"
)
const { contracts } = require("@keep-network/common.js")
const { readExternalContractAddress } = contracts

const initLPRewardContract = async (
  LPRewardsContract,
  KEEPTokenContract,
  lprewardsOwner,
  reward
) => {
  await LPRewardsContract.setRewardDistribution(lprewardsOwner, {
    from: lprewardsOwner,
  })
  await KEEPTokenContract.approve(LPRewardsContract.address, reward, {
    from: lprewardsOwner,
  })
  await KEEPTokenContract.transfer(LPRewardsContract.address, reward, {
    from: lprewardsOwner,
  })
  await LPRewardsContract.notifyRewardAmount(reward, {
    from: lprewardsOwner,
  })
}

const mintAndApproveLPReward = async (
  WrappedTokenContract,
  lpRewardsAddress,
  address,
  amount
) => {
  await WrappedTokenContract.mint(address, amount)
  await WrappedTokenContract.approve(lpRewardsAddress, amount, {
    from: address,
  })
}

const getWrappedTokenContract = async (lpRewardsContract) => {
  const address = await lpRewardsContract.wrappedToken()
  return await TestToken.at(address)
}

module.exports = async function () {
  try {
    const networkID = await web3.eth.net.getId()

    const accounts = await web3.eth.getAccounts()
    const lpRewardsOwner = accounts[0]
    const KeepTokenAddress = readExternalContractAddress(
      "@keep-network/keep-core",
      "KeepToken",
      networkID
    )
    const keepToken = await KeepToken.at(KeepTokenAddress)
    const reward = web3.utils.toWei("1000000")
    const LPRewardsContracts = [
      LPRewardsKEEPETH,
      LPRewardsTBTCETH,
      LPRewardsKEEPTBTC,
      LPRewardsTBTCSaddle,
      LPRewardsTBTCv2Saddle,
      LPRewardsTBTCv2SaddleV2,
    ]

    for (const LPRewardsContract of LPRewardsContracts) {
      const lpRewardsContract = await LPRewardsContract.deployed()

      await initLPRewardContract(
        lpRewardsContract,
        keepToken,
        lpRewardsOwner,
        reward
      )

      const wrappedTokenContract = await getWrappedTokenContract(
        lpRewardsContract
      )

      for (let i = 8; i < 10; i++) {
        const staker = accounts[i]
        const amount = web3.utils.toWei(`${i * 100}`)

        await mintAndApproveLPReward(
          wrappedTokenContract,
          lpRewardsContract.address,
          staker,
          amount
        )

        await lpRewardsContract.stake(amount, { from: staker })
      }
    }
  } catch (err) {
    console.error(err)
    process.exit(1)
  }

  process.exit(0)
}
