const LPRewardsTBTCETH = artifacts.require("LPRewardsTBTCETH")
const LPRewardsKEEPETH = artifacts.require("LPRewardsKEEPETH")
const LPRewardsKEEPTBTC = artifacts.require("LPRewardsKEEPTBTC")
const TestToken = artifacts.require("./test/TestToken")
const KeepToken = artifacts.require(
  "@keep-network/keep-core/build/truffle/KeepToken"
)
const { KeepTokenAddress } = require("../migrations/external-contracts")

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

module.exports = async function () {
  try {
    const accounts = await web3.eth.getAccounts()
    const lpRewardsOwner = accounts[0]
    const lpRewardsKEEPETH = await LPRewardsKEEPETH.deployed()
    const lpRewardsTBTCETH = await LPRewardsTBTCETH.deployed()
    const lpRewardsKEEPTBTC = await LPRewardsKEEPTBTC.deployed()

    const KEEPETHWrappedTokenAddress = await lpRewardsKEEPETH.wrappedToken()
    const KEEPETHWrappedTokenContract = await TestToken.at(
      KEEPETHWrappedTokenAddress
    )

    const TBTCETHWrappedTokenAddress = await lpRewardsTBTCETH.wrappedToken()
    const TBTCETHHWrappedTokenContract = await TestToken.at(
      TBTCETHWrappedTokenAddress
    )

    const KEEPTBTCWrappedTokenAddress = await lpRewardsKEEPTBTC.wrappedToken()
    const KEEPTBTCHWrappedTokenContract = await TestToken.at(
      KEEPTBTCWrappedTokenAddress
    )

    const keepToken = await KeepToken.at(KeepTokenAddress)

    const reward = web3.utils.toWei("1000000")

    await initLPRewardContract(
      lpRewardsKEEPETH,
      keepToken,
      lpRewardsOwner,
      reward
    )
    await initLPRewardContract(
      lpRewardsTBTCETH,
      keepToken,
      lpRewardsOwner,
      reward
    )
    await initLPRewardContract(
      lpRewardsKEEPTBTC,
      keepToken,
      lpRewardsOwner,
      reward
    )

    const staker1 = accounts[8]
    const staker1WrappedTokenBalance = web3.utils.toWei("300")

    const staker2 = accounts[9]
    const staker2WrappedTokenBalance = web3.utils.toWei("700")

    await mintAndApproveLPReward(
      KEEPETHWrappedTokenContract,
      lpRewardsKEEPETH.address,
      staker1,
      staker1WrappedTokenBalance
    )
    await mintAndApproveLPReward(
      KEEPETHWrappedTokenContract,
      lpRewardsKEEPETH.address,
      staker2,
      staker2WrappedTokenBalance
    )

    await mintAndApproveLPReward(
      TBTCETHHWrappedTokenContract,
      lpRewardsTBTCETH.address,
      staker1,
      staker1WrappedTokenBalance
    )
    await mintAndApproveLPReward(
      TBTCETHHWrappedTokenContract,
      lpRewardsTBTCETH.address,
      staker2,
      staker2WrappedTokenBalance
    )

    await mintAndApproveLPReward(
      KEEPTBTCHWrappedTokenContract,
      lpRewardsKEEPTBTC.address,
      staker1,
      staker1WrappedTokenBalance
    )
    await mintAndApproveLPReward(
      KEEPTBTCHWrappedTokenContract,
      lpRewardsKEEPTBTC.address,
      staker2,
      staker2WrappedTokenBalance
    )

    await lpRewardsKEEPETH.stake(staker1WrappedTokenBalance, { from: staker1 })
    await lpRewardsKEEPETH.stake(staker2WrappedTokenBalance, { from: staker2 })

    await lpRewardsTBTCETH.stake(staker1WrappedTokenBalance, { from: staker1 })
    await lpRewardsTBTCETH.stake(staker2WrappedTokenBalance, { from: staker2 })

    await lpRewardsKEEPTBTC.stake(staker1WrappedTokenBalance, { from: staker1 })
    await lpRewardsKEEPTBTC.stake(staker2WrappedTokenBalance, { from: staker2 })
  } catch (err) {
    console.error(err)
    process.exit(1)
  }

  process.exit(0)
}
