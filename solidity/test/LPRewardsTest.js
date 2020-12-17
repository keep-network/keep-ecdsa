const {contract, web3} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("./helpers/snapshot")
const {accounts} = require("@openzeppelin/test-environment")

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
  let wallet1
  let wallet2

  before(async () => {
    wallet1 = accounts[1]
    wallet2 = accounts[2]
    keepToken = await KeepToken.new()
    // This is a "Pair" Uniswap Token which is created here:
    // https://github.com/Uniswap/uniswap-v2-core/blob/master/contracts/UniswapV2Factory.sol#L23
    //
    // Before the deployment, we need to have 3 addresses for the following pairs:
    // - KEEP/ETH (https://info.uniswap.org/pair/0xe6f19dab7d43317344282f803f8e8d240708174a)
    // - TBTC/ETH (https://info.uniswap.org/pair/0x854056fd40c1b52037166285b2e54fee774d33f6)
    // - KEEP/TBTC (tbd)
    wrappedToken = await WrappedToken.new()
    lpRewards = await LPRewards.new(keepToken.address, wrappedToken.address)
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("allocating tokens", () => {
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
      // 1,000 - 4,000 = 6,000
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

  async function mintAndApproveWrappedTokens(token, address, wallet, amount) {
    await token.mint(wallet, amount)
    await token.approve(address, amount, {from: wallet})
  }

  async function fundKEEPReward(address, amount) {
    await keepToken.approveAndCall(address, amount, "0x0")
  }
})
