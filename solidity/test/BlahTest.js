
const {contract, web3, accounts} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("./helpers/snapshot")
const {time, expectRevert} = require("@openzeppelin/test-helpers")

const LPRewards = contract.fromArtifact("LPRewards")
const KeepToken = contract.fromArtifact("KeepToken")
const WrappedToken = contract.fromArtifact("TestToken")

const LPRewardsTBTCETH = contract.fromArtifact("LPRewardsTBTCETH")
const LPRewardsKEEPETH = contract.fromArtifact("LPRewardsKEEPETH")
const LPRewardsKEEPTBTC = contract.fromArtifact("LPRewardsKEEPTBTC")
const LPRewardsTBTCSaddle = contract.fromArtifact("LPRewardsTBTCSaddle")

const LPRewardsStaker = contract.fromArtifact("LPRewardsStaker")

const BatchedPhasedEscrow = contract.fromArtifact("BatchedPhasedEscrow")
const StakingPoolRewardsEscrowBeneficiary = contract.fromArtifact(
  "StakingPoolRewardsEscrowBeneficiary"
)

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
    it("blah", async () => {
      const wrappedTokenStakerBalance1 = web3.utils
        .toBN(10000)
        .mul(tokenDecimalMultiplier)

      await mintAndApproveWrappedTokens(staker1, wrappedTokenStakerBalance1)

      let wrappedTokenBalance = await wrappedToken.balanceOf(lpRewards.address)
      expect(wrappedTokenBalance).to.eq.BN(0)

      await lpRewards.stake(web3.utils.toBN(4000).mul(tokenDecimalMultiplier), {
        from: staker1,
      })

      const totalSupply = await lpRewards.totalSupply()
      expect(totalSupply).to.eq.BN(
        web3.utils.toBN(4000).mul(tokenDecimalMultiplier)
      )

      const future = (await time.latest()).add(time.duration.days(7))
      await time.increaseTo(future)

      const keepRewards = new BN(10000).mul(tokenDecimalMultiplier)
      await fundAndNotifyLPRewards(lpRewards.address, keepRewards)

      const earned = await lpRewards.earned(staker1)
      console.log(earned)
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

