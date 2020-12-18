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

  let keepToken
  let lpRewards
  let wrappedToken
  let owner
  let wallet1
  let wallet2

  before(async () => {
    owner = accounts[0]
    wallet1 = accounts[1]
    wallet2 = accounts[2]
    keepToken = await KeepToken.new()
    // This is a "Pair" Uniswap Token which is created here:
    // https://github.com/Uniswap/uniswap-v2-core/blob/master/contracts/UniswapV2Factory.sol#L23
    //
    // There are 3 addresses for the following pairs:
    // - KEEP/ETH (https://info.uniswap.org/pair/0xe6f19dab7d43317344282f803f8e8d240708174a)
    // - TBTC/ETH (https://info.uniswap.org/pair/0x854056fd40c1b52037166285b2e54fee774d33f6)
    // - KEEP/TBTC (https://info.uniswap.org/pair/0x38c8ffee49f286f25d25bad919ff7552e5daf081)
    wrappedToken = await WrappedToken.new()
    lpRewards = await LPRewards.new(keepToken.address, wrappedToken.address)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("tokens allocation", () => {
    it("should successfully allocate KEEP tokens", async () => {
      let keepBalance = await keepToken.balanceOf(lpRewards.address)
      expect(keepBalance).to.eq.BN(0)

      const rewards = web3.utils.toBN(1000042).mul(tokenDecimalMultiplier)
      await fundKEEPReward(lpRewards.address, rewards)

      keepBalance = await keepToken.balanceOf(lpRewards.address)
      expect(keepBalance).to.eq.BN(rewards)
    })

    it("should successfully allocate wrapped tokens", async () => {
      const walletBallance1 = web3.utils.toBN(10000).mul(tokenDecimalMultiplier)
      const walletBallance2 = web3.utils.toBN(20000).mul(tokenDecimalMultiplier)

      await mintAndApproveWrappedTokens(
        wrappedToken,
        lpRewards.address,
        wallet1,
        walletBallance1
      )
      await mintAndApproveWrappedTokens(
        wrappedToken,
        lpRewards.address,
        wallet2,
        walletBallance2
      )

      let wrappedTokenBalance = await wrappedToken.balanceOf(lpRewards.address)
      expect(wrappedTokenBalance).to.eq.BN(0)

      let wrappedTokenWalletBalance1 = await wrappedToken.balanceOf(wallet1)
      expect(wrappedTokenWalletBalance1).to.eq.BN(walletBallance1)
      let wrappedTokenWalletBalance2 = await wrappedToken.balanceOf(wallet2)
      expect(wrappedTokenWalletBalance2).to.eq.BN(walletBallance2)

      await lpRewards.stake(web3.utils.toBN(4000).mul(tokenDecimalMultiplier), {
        from: wallet1,
      })
      await lpRewards.stake(web3.utils.toBN(5000).mul(tokenDecimalMultiplier), {
        from: wallet2,
      })

      wrappedTokenBalance = await wrappedToken.balanceOf(lpRewards.address)
      // 4,000 + 5,000 = 9,000
      expect(wrappedTokenBalance).to.eq.BN(
        web3.utils.toBN(9000).mul(tokenDecimalMultiplier)
      )

      wrappedTokenWalletBalance1 = await wrappedToken.balanceOf(wallet1)
      // 10,000 - 4,000 = 6,000
      expect(wrappedTokenWalletBalance1).to.eq.BN(
        web3.utils.toBN(6000).mul(tokenDecimalMultiplier)
      )

      wrappedTokenWalletBalance2 = await wrappedToken.balanceOf(wallet2)
      // 20,000 - 5,000 = 15,000
      expect(wrappedTokenWalletBalance2).to.eq.BN(
        web3.utils.toBN(15000).mul(tokenDecimalMultiplier)
      )
    })
  })

  describe("tokens distribution", () => {
    const precision = 1

    it("should be possible to check earned rewards", async () => {
      const keepRewards = new BN(10000).mul(tokenDecimalMultiplier)
      const wrappedTokenWalletBallance = web3.utils
        .toBN(10000)
        .mul(tokenDecimalMultiplier)

      await fundKEEPReward(lpRewards.address, keepRewards)
      await mintAndApproveWrappedTokens(
        wrappedToken,
        lpRewards.address,
        wallet1,
        wrappedTokenWalletBallance
      )

      await lpRewards.setRewardDistribution(owner)
      await lpRewards.notifyRewardAmount(keepRewards, {from: owner})

      const wrappedTokensToStake = web3.utils
        .toBN(2000)
        .mul(tokenDecimalMultiplier)
      await lpRewards.stake(wrappedTokensToStake, {from: wallet1})

      const future = (await time.latest()).add(time.duration.days(7))
      await timeIncreaseTo(future)

      const actualEarnings = (await lpRewards.earned(wallet1)).div(
        tokenDecimalMultiplier
      )
      const expectedEarnings = new BN(10000)
      expect(actualEarnings).to.gte.BN(expectedEarnings.subn(precision))
      expect(actualEarnings).to.lte.BN(expectedEarnings.addn(precision))

      // 10,000 / 2,000 = 5 KEEP,
      // Contract function will return 4999999999999999881 KEEP. Solidity
      // does not have floating numbers and when dividing by 10^18, precision will
      // not be on point.
      const actualRewardPerToken = (await lpRewards.rewardPerToken()).div(
        tokenDecimalMultiplier
      )
      const expectedRewardPerWrappedToken = new BN(5)
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
      const wrappedTokenWalletBallance = web3.utils
        .toBN(10000)
        .mul(tokenDecimalMultiplier)

      await fundKEEPReward(lpRewards.address, keepAllocated)
      await mintAndApproveWrappedTokens(
        wrappedToken,
        lpRewards.address,
        wallet1,
        wrappedTokenWalletBallance
      )

      await lpRewards.setRewardDistribution(owner)
      await lpRewards.notifyRewardAmount(keepAllocated, {from: owner})

      const wrappedTokensToStake = web3.utils
        .toBN(2000)
        .mul(tokenDecimalMultiplier)
      await lpRewards.stake(wrappedTokensToStake, {from: wallet1})

      const future = (await time.latest()).add(time.duration.days(7))
      await timeIncreaseTo(future)

      // Withdraw wrapped tokens and KEEP rewards
      await lpRewards.exit({from: wallet1})

      // Earned KEEP rewards for adding liquidity
      const keepEarnedRewards = (await keepToken.balanceOf(wallet1)).div(
        tokenDecimalMultiplier
      )
      expect(keepEarnedRewards).to.gte.BN(rewardsAmount.subn(precision))
      expect(keepEarnedRewards).to.lte.BN(rewardsAmount.addn(precision))

      // Check that all wrapped tokens were transferred back to the wallet
      const wrappedTokenWalletBalance = await wrappedToken.balanceOf(wallet1)
      expect(wrappedTokenWalletBalance).to.eq.BN(wrappedTokenWalletBallance)
    })
  })

  async function mintAndApproveWrappedTokens(token, address, wallet, amount) {
    await token.mint(wallet, amount)
    await token.approve(address, amount, {from: wallet})
  }

  async function fundKEEPReward(address, amount) {
    await keepToken.approveAndCall(address, amount, "0x0")
  }

  async function timeIncreaseTo(seconds) {
    const delay = 10 - new Date().getMilliseconds()
    await new Promise((resolve) => setTimeout(resolve, delay))
    await time.increaseTo(seconds)
  }
})
