const LPRewardsTBTCETH = artifacts.require("LPRewardsTBTCETH")
const LPRewardsKEEPETH = artifacts.require("LPRewardsKEEPETH")
const LPRewardsKEEPTBTC = artifacts.require("LPRewardsKEEPTBTC")
const LPRewardsTBTCSaddle = artifacts.require("LPRewardsTBTCSaddle")

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

const getWrappedTokenContract = async (LPRewardsContract) => {
  const address = await LPRewardsContract.wrappedToken()
  return await TestToken.at(address)
}

module.exports = async function () {
  try {
    const accounts = await web3.eth.getAccounts()
    const lpRewardsOwner = accounts[0]

    const keepToken = await KeepToken.at(KeepTokenAddress)
    const lpRewardsKEEPETH = await LPRewardsKEEPETH.deployed()
    const lpRewardsTBTCETH = await LPRewardsTBTCETH.deployed()
    const lpRewardsKEEPTBTC = await LPRewardsKEEPTBTC.deployed()
    const lpRewardsTBTCSaddle = await LPRewardsTBTCSaddle.deployed()
    const reward = web3.utils.toWei("1000000")

    const LP_REWARDS = {
      KEEP_ETH: {
        contract: lpRewardsKEEPETH,
        wrappedTokenContract: null,
      },
      TBTC_ETH: {
        contract: lpRewardsTBTCETH,
        wrappedTokenContract: null,
      },
      KEEP_TBTC: {
        contract: lpRewardsKEEPTBTC,
        wrappedTokenContract: null,
      },
      TBTC_SADDLE: {
        contract: lpRewardsTBTCSaddle,
        wrappedTokenContract: null,
      },
    }

    for (const key of Object.keys(LP_REWARDS)) {
      LP_REWARDS[key].wrappedTokenContract = await getWrappedTokenContract(
        LP_REWARDS[key].contract
      )

      await initLPRewardContract(
        LP_REWARDS[key].contract,
        keepToken,
        lpRewardsOwner,
        reward
      )
    }

    for (let i = 8; i < 10; i++) {
      const staker = accounts[i]
      const amount = web3.utils.toWei(`${i * 100}`)

      for (const values of Object.values(LP_REWARDS)) {
        const lpRewardsContract = values.contract

        await mintAndApproveLPReward(
          values.wrappedTokenContract,
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
