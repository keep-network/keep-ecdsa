const {contract, web3, accounts} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("./helpers/snapshot")
const {time} = require("@openzeppelin/test-helpers")

const LPRewards = contract.fromArtifact("LPRewards")
const KeepToken = contract.fromArtifact("KeepToken")
const WrappedToken = contract.fromArtifact("TestToken")

const BN = web3.utils.BN
const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

describe("LPRewards", () => {
  const tokenDecimalMultiplier = web3.utils.toBN(10).pow(web3.utils.toBN(18))
  const allocationForDistribution = web3.utils
    .toBN(10000000)
    .mul(tokenDecimalMultiplier)

  const staker1 = accounts[1]
  const staker2 = accounts[2]
  const rewardDistribution = accounts[3]
  const wrappedTokenOwner = accounts[4]
  const lpRewardsOwner = accounts[5]
  const keepTokenOwner = accounts[6]

  let keepToken
  let lpRewards
  let wrappedToken

  before(async () => {
    keepToken = await KeepToken.new({from: keepTokenOwner})
    // This is a "Pair" Uniswap Token which is created here:
    // https://github.com/Uniswap/uniswap-v2-core/blob/master/contracts/UniswapV2Factory.sol#L23
    //
    // There are 3 addresses for the following pairs:
    // - KEEP/ETH (https://info.uniswap.org/pair/0xe6f19dab7d43317344282f803f8e8d240708174a)
    // - TBTC/ETH (https://info.uniswap.org/pair/0x854056fd40c1b52037166285b2e54fee774d33f6)
    // - KEEP/TBTC (https://info.uniswap.org/pair/0x38c8ffee49f286f25d25bad919ff7552e5daf081)
    wrappedToken = await WrappedToken.new({from: wrappedTokenOwner})
    lpRewards = await LPRewards.new(keepToken.address, wrappedToken.address, {
      from: lpRewardsOwner,
    })

    await keepToken.transfer(rewardDistribution, allocationForDistribution, {
      from: keepTokenOwner,
    })
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("tokens allocation", () => {
    it("should successfully transfer KEEP and notify LPRewards contract on rewards distribution", async () => {
      const rewards = web3.utils.toBN(1042).mul(tokenDecimalMultiplier)

      await lpRewards.setRewardDistribution(rewardDistribution, {
        from: lpRewardsOwner,
      })
      await keepToken.approve(lpRewards.address, rewards, {
        from: rewardDistribution,
      })
      await lpRewards.notifyRewardAmount(rewards, {
        from: rewardDistribution,
      })

      const finalBalance = await keepToken.balanceOf(lpRewards.address)
      expect(finalBalance).to.eq.BN(rewards)
    })

    it("should successfully stake wrapped tokens", async () => {
      const wrappedTokenStakerBalance1 = web3.utils
        .toBN(10000)
        .mul(tokenDecimalMultiplier)
      const wrappedTokenStakerBalance2 = web3.utils
        .toBN(20000)
        .mul(tokenDecimalMultiplier)

      await mintAndApproveWrappedTokens(staker1, wrappedTokenStakerBalance1)
      await mintAndApproveWrappedTokens(staker2, wrappedTokenStakerBalance2)

      let wrappedTokenBalance = await wrappedToken.balanceOf(lpRewards.address)
      expect(wrappedTokenBalance).to.eq.BN(0)

      await lpRewards.stake(web3.utils.toBN(4000).mul(tokenDecimalMultiplier), {
        from: staker1,
      })
      await lpRewards.stake(web3.utils.toBN(5000).mul(tokenDecimalMultiplier), {
        from: staker2,
      })

      wrappedTokenBalance = await wrappedToken.balanceOf(lpRewards.address)
      // 4,000 + 5,000 = 9,000
      expect(wrappedTokenBalance).to.eq.BN(
        web3.utils.toBN(9000).mul(tokenDecimalMultiplier)
      )

      const totalSupply = await lpRewards.totalSupply()
      expect(totalSupply).to.eq.BN(
        web3.utils.toBN(9000).mul(tokenDecimalMultiplier)
      )

      const balanceStaker1 = await lpRewards.balanceOf(staker1)
      expect(balanceStaker1).to.eq.BN(
        web3.utils.toBN(4000).mul(tokenDecimalMultiplier)
      )

      const balanceStaker2 = await lpRewards.balanceOf(staker2)
      expect(balanceStaker2).to.eq.BN(
        web3.utils.toBN(5000).mul(tokenDecimalMultiplier)
      )

      const actualWrappedTokenStakerBalance1 = await wrappedToken.balanceOf(
        staker1
      )
      // 10,000 - 4,000 = 6,000
      expect(actualWrappedTokenStakerBalance1).to.eq.BN(
        web3.utils.toBN(6000).mul(tokenDecimalMultiplier)
      )

      const actualWrappedTokenStakerBalance2 = await wrappedToken.balanceOf(
        staker2
      )
      // 20,000 - 5,000 = 15,000
      expect(actualWrappedTokenStakerBalance2).to.eq.BN(
        web3.utils.toBN(15000).mul(tokenDecimalMultiplier)
      )
    })
  })

  describe("tokens distribution", () => {
    const precision = 1

    it("should be possible to check earned rewards", async () => {
      const keepRewards = new BN(10000).mul(tokenDecimalMultiplier)
      const wrappedTokenStakerBalance = web3.utils
        .toBN(10000)
        .mul(tokenDecimalMultiplier)

      await mintAndApproveWrappedTokens(staker1, wrappedTokenStakerBalance)

      await fundAndNotifyLPRewards(lpRewards.address, keepRewards)

      const wrappedTokensToStake = web3.utils
        .toBN(2000)
        .mul(tokenDecimalMultiplier)
      await lpRewards.stake(wrappedTokensToStake, {from: staker1})

      const periodFinish = (await time.latest()).add(time.duration.days(7))
      await time.increaseTo(periodFinish)

      const actualEarnings = (await lpRewards.earned(staker1)).div(
        tokenDecimalMultiplier
      )
      const expectedEarnings = new BN(10000)
      expect(actualEarnings).to.gte.BN(expectedEarnings.subn(precision))
      expect(actualEarnings).to.lte.BN(expectedEarnings.addn(precision))

      // duration = 7 days = 604,800 sec
      // rewardRate = keep rewards / duration
      // rewardRate = 10,000 / 604,800 = 0,016534391534391534
      // expected reward per wrapped token = rewardRate * duration / staked amount
      // expected reward per wrapped token = 5 KEEP
      // Contract function rewardPerToken() will return 4999999999999999881 KEEP wei.
      // Solidity does not have floating numbers and when dividing by 10^18,
      // precision will not be on point.
      const expectedRewardPerWrappedToken = new BN(5)

      const actualRewardPerToken = (await lpRewards.rewardPerToken()).div(
        tokenDecimalMultiplier
      )

      expect(actualRewardPerToken).to.gte.BN(
        expectedRewardPerWrappedToken.subn(precision)
      )
      expect(actualRewardPerToken).to.lte.BN(
        expectedRewardPerWrappedToken.addn(precision)
      )
    })

    it("should be possible to withdraw rewards after staking wrapped tokens", async () => {
      const rewardsAmount = new BN(5000)
      const keepAllocated = rewardsAmount.mul(tokenDecimalMultiplier)
      const wrappedTokenStakerBalance = web3.utils
        .toBN(10000)
        .mul(tokenDecimalMultiplier)

      await mintAndApproveWrappedTokens(staker1, wrappedTokenStakerBalance)

      await fundAndNotifyLPRewards(lpRewards.address, keepAllocated)

      const wrappedTokensToStake = web3.utils
        .toBN(2000)
        .mul(tokenDecimalMultiplier)
      await lpRewards.stake(wrappedTokensToStake, {from: staker1})

      const future = (await time.latest()).add(time.duration.days(7))
      await time.increaseTo(future)

      // Withdraw wrapped tokens and KEEP rewards
      await lpRewards.exit({from: staker1})

      // Earned KEEP rewards for adding liquidity
      const keepEarnedRewards = (await keepToken.balanceOf(staker1)).div(
        tokenDecimalMultiplier
      )
      expect(keepEarnedRewards).to.gte.BN(rewardsAmount.subn(precision))
      expect(keepEarnedRewards).to.lte.BN(rewardsAmount.addn(precision))

      // Check that all wrapped tokens were transferred back to the staker
      const actualWrappedTokenStakerBalance = await wrappedToken.balanceOf(
        staker1
      )
      expect(wrappedTokenStakerBalance).to.eq.BN(
        actualWrappedTokenStakerBalance
      )
    })
  })

  async function fundAndNotifyLPRewards(address, amount) {
    await lpRewards.setRewardDistribution(rewardDistribution, {
      from: lpRewardsOwner,
    })
    await keepToken.approve(address, amount, {from: rewardDistribution})
    await lpRewards.notifyRewardAmount(amount, {
      from: rewardDistribution,
    })
  }

  async function mintAndApproveWrappedTokens(staker, amount) {
    await wrappedToken.mint(staker, amount)
    await wrappedToken.approve(lpRewards.address, amount, {from: staker})
  }
})
