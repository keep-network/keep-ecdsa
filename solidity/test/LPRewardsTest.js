const {contract, web3, accounts} = require("@openzeppelin/test-environment")
const {createSnapshot, restoreSnapshot} = require("./helpers/snapshot")
const {time} = require("@openzeppelin/test-helpers")

const LPRewards = contract.fromArtifact("LPRewards")
const KeepToken = contract.fromArtifact("KeepToken")
const WrappedToken = contract.fromArtifact("TestToken")

const LPRewardsTBTCETH = contract.fromArtifact("LPRewardsTBTCETH")
const LPRewardsKEEPETH = contract.fromArtifact("LPRewardsKEEPETH")
const LPRewardsKEEPTBTC = contract.fromArtifact("LPRewardsKEEPTBTC")
const BatchedPhasedEscrow = contract.fromArtifact("BatchedPhasedEscrow")
const LPRewardsEscrowBeneficiary = contract.fromArtifact(
  "LPRewardsEscrowBeneficiary"
)

const BN = web3.utils.BN
const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

describe("LPRewards", () => {
  const tokenDecimalMultiplier = web3.utils.toBN(10).pow(web3.utils.toBN(18))

  let keepToken
  let lpRewards
  let wrappedToken
  let rewardDistribution
  let staker1
  let staker2
  let keepTokenContractOwner
  let wrappedTokenContractOwner
  let lpRewardsContractOwner

  before(async () => {
    staker1 = accounts[1]
    staker2 = accounts[2]
    keepTokenContractOwner = accounts[3]
    wrappedTokenContractOwner = accounts[4]
    lpRewardsContractOwner = accounts[5]
    rewardDistribution = accounts[6]

    keepToken = await KeepToken.new({from: keepTokenContractOwner})
    // This is a "Pair" Uniswap Token which is created here:
    // https://github.com/Uniswap/uniswap-v2-core/blob/master/contracts/UniswapV2Factory.sol#L23
    //
    // There are 3 addresses for the following pairs:
    // - KEEP/ETH (https://info.uniswap.org/pair/0xe6f19dab7d43317344282f803f8e8d240708174a)
    // - TBTC/ETH (https://info.uniswap.org/pair/0x854056fd40c1b52037166285b2e54fee774d33f6)
    // - KEEP/TBTC (https://info.uniswap.org/pair/0x38c8ffee49f286f25d25bad919ff7552e5daf081)
    wrappedToken = await WrappedToken.new({from: wrappedTokenContractOwner})
    lpRewards = await LPRewards.new(keepToken.address, wrappedToken.address, {
      from: lpRewardsContractOwner,
    })
  })

  beforeEach(async () => {
    await createSnapshot()
  })

  afterEach(async () => {
    await restoreSnapshot()
  })

  describe("tokens allocation", () => {
    it("should successfully allocate KEEP tokens via receiveApproval function", async () => {
      const initialBalance = await keepToken.balanceOf(lpRewards.address)

      const rewards = web3.utils.toBN(1000042).mul(tokenDecimalMultiplier)

      await keepToken.approveAndCall(lpRewards.address, rewards, "0x0", {
        from: keepTokenContractOwner,
      })

      const finalBalance = await keepToken.balanceOf(lpRewards.address)
      expect(finalBalance).to.eq.BN(rewards.add(initialBalance))
    })

    it("should successfully allocate wrapped tokens", async () => {
      const initialWrappedTokenStakerBallance1 = web3.utils
        .toBN(10000)
        .mul(tokenDecimalMultiplier)
      const initialWrappedTokenStakerBallance2 = web3.utils
        .toBN(20000)
        .mul(tokenDecimalMultiplier)

      await mintAndApproveWrappedTokens(
        wrappedToken,
        lpRewards.address,
        staker1,
        initialWrappedTokenStakerBallance1
      )
      await mintAndApproveWrappedTokens(
        wrappedToken,
        lpRewards.address,
        staker2,
        initialWrappedTokenStakerBallance2
      )

      let wrappedTokenBalance = await wrappedToken.balanceOf(lpRewards.address)
      expect(wrappedTokenBalance).to.eq.BN(0)

      let wrappedTokenStakerBalance1 = await wrappedToken.balanceOf(staker1)
      expect(wrappedTokenStakerBalance1).to.eq.BN(
        initialWrappedTokenStakerBallance1
      )
      let wrappedTokenStakerBalance2 = await wrappedToken.balanceOf(staker2)
      expect(wrappedTokenStakerBalance2).to.eq.BN(
        initialWrappedTokenStakerBallance2
      )

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

      wrappedTokenStakerBalance1 = await wrappedToken.balanceOf(staker1)
      // 10,000 - 4,000 = 6,000
      expect(wrappedTokenStakerBalance1).to.eq.BN(
        web3.utils.toBN(6000).mul(tokenDecimalMultiplier)
      )

      wrappedTokenStakerBalance2 = await wrappedToken.balanceOf(staker2)
      // 20,000 - 5,000 = 15,000
      expect(wrappedTokenStakerBalance2).to.eq.BN(
        web3.utils.toBN(15000).mul(tokenDecimalMultiplier)
      )
    })
  })

  describe("tokens distribution", () => {
    const precision = 1

    it("should be possible to check earned rewards", async () => {
      const keepRewards = new BN(10000).mul(tokenDecimalMultiplier)
      const wrappedTokenStakerBallance = web3.utils
        .toBN(10000)
        .mul(tokenDecimalMultiplier)

      await fundKEEPReward(lpRewards.address, keepRewards)
      await mintAndApproveWrappedTokens(
        wrappedToken,
        lpRewards.address,
        staker1,
        wrappedTokenStakerBallance
      )

      await lpRewards.setRewardDistribution(rewardDistribution, {
        from: lpRewardsContractOwner,
      })
      await lpRewards.notifyRewardAmount(keepRewards, {
        from: rewardDistribution,
      })

      const wrappedTokensToStake = web3.utils
        .toBN(2000)
        .mul(tokenDecimalMultiplier)
      await lpRewards.stake(wrappedTokensToStake, {from: staker1})

      const future = (await time.latest()).add(time.duration.days(7))
      await timeIncreaseTo(future)

      const actualEarnings = (await lpRewards.earned(staker1)).div(
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
      const wrappedTokenStakerBallance = web3.utils
        .toBN(10000)
        .mul(tokenDecimalMultiplier)

      await fundKEEPReward(lpRewards.address, keepAllocated)
      await mintAndApproveWrappedTokens(
        wrappedToken,
        lpRewards.address,
        staker1,
        wrappedTokenStakerBallance
      )

      await lpRewards.setRewardDistribution(rewardDistribution, {
        from: lpRewardsContractOwner,
      })
      await lpRewards.notifyRewardAmount(keepAllocated, {
        from: rewardDistribution,
      })

      const wrappedTokensToStake = web3.utils
        .toBN(2000)
        .mul(tokenDecimalMultiplier)
      await lpRewards.stake(wrappedTokensToStake, {from: staker1})

      const future = (await time.latest()).add(time.duration.days(7))
      await timeIncreaseTo(future)

      // Withdraw wrapped tokens and KEEP rewards
      await lpRewards.exit({from: staker1})

      // Earned KEEP rewards for adding liquidity
      const keepEarnedRewards = (await keepToken.balanceOf(staker1)).div(
        tokenDecimalMultiplier
      )
      expect(keepEarnedRewards).to.gte.BN(rewardsAmount.subn(precision))
      expect(keepEarnedRewards).to.lte.BN(rewardsAmount.addn(precision))

      // Check that all wrapped tokens were transferred back to the staker
      const wrappedTokenStakerBalance = await wrappedToken.balanceOf(staker1)
      expect(wrappedTokenStakerBalance).to.eq.BN(wrappedTokenStakerBallance)
    })
  })

  describe("dedicated rewards contracts", () => {
    const owner = accounts[1]

    it("deploys contract for TBTC-ETH pair", async () => {
      const lpRewards = await LPRewardsTBTCETH.new(
        keepToken.address,
        wrappedToken.address
      )

      expect(await lpRewards.keepToken.call()).equal(keepToken.address)
      expect(await lpRewards.wrappedToken.call()).equal(wrappedToken.address)
    })

    it("deploys contract for KEEP-ETH pair", async () => {
      const lpRewards = await LPRewardsKEEPETH.new(
        keepToken.address,
        wrappedToken.address
      )

      expect(await lpRewards.keepToken.call()).equal(keepToken.address)
      expect(await lpRewards.wrappedToken.call()).equal(wrappedToken.address)
    })

    it("deploys contract for KEEP-TBTC pair", async () => {
      const lpRewards = await LPRewardsKEEPTBTC.new(
        keepToken.address,
        wrappedToken.address
      )

      expect(await lpRewards.keepToken.call()).equal(keepToken.address)
      expect(await lpRewards.wrappedToken.call()).equal(wrappedToken.address)
    })

    it("got funded from batched escrow", async () => {
      // Here we want to test funding of LP Rewards contracts from Batched Phased
      // Escrow. It is a type of Phased Escrow that allows batched withdrawals
      // in a one function call. This way of funding requires intermediary
      // Escrow Beneficiary contracts, that will automatically transfer funds
      // to the correct LP Rewards contract.
      //
      // Tokens are transferred in the following way:
      //                        |-> LPRewardsEscrowBeneficiary -> LPRewardsTBTCETH
      // Batched Phased Escrow -|-> LPRewardsEscrowBeneficiary -> LPRewardsKEEPETH
      //                        |-> LPRewardsEscrowBeneficiary -> LPRewardsKEEPTBTC
      //
      // It is expected that one call of `BatchedPhasedEscrow.batchedWithdrawal`
      // function will transfer specific number of tokens to dedicated LP Rewards
      // contracts.

      const rewards = [
        web3.utils.toBN(1001).mul(tokenDecimalMultiplier),
        web3.utils.toBN(2010).mul(tokenDecimalMultiplier),
        web3.utils.toBN(3100).mul(tokenDecimalMultiplier),
      ]
      const totalRewards = web3.utils.toBN(6111).mul(tokenDecimalMultiplier)

      // Deploy BatchedPhasedEscrow contract and fund it.
      const fundingEscrow = await BatchedPhasedEscrow.new(keepToken.address, {
        from: owner,
      })
      await keepToken.approveAndCall(
        fundingEscrow.address,
        totalRewards,
        "0x0",
        {
          from: keepTokenContractOwner,
        }
      )

      // Deploy LP Rewards contracts for each pair.
      const lpReward1 = await LPRewardsTBTCETH.new(
        keepToken.address,
        wrappedToken.address
      )
      const lpReward2 = await LPRewardsKEEPETH.new(
        keepToken.address,
        wrappedToken.address
      )

      const lpReward3 = await LPRewardsKEEPTBTC.new(
        keepToken.address,
        wrappedToken.address
      )

      // Deploy beneficiaries contracts for each LP Rewards contract.
      const beneficiary1 = await newLPRewardsBeneficiary(
        lpReward1.address,
        fundingEscrow.address
      )

      const beneficiary2 = await newLPRewardsBeneficiary(
        lpReward2.address,
        fundingEscrow.address
      )

      const beneficiary3 = await newLPRewardsBeneficiary(
        lpReward3.address,
        fundingEscrow.address
      )

      const beneficiaries = [
        beneficiary1.address,
        beneficiary2.address,
        beneficiary3.address,
      ]

      // Perform batched withdrawal from Phased Escrow contract.
      await fundingEscrow.batchedWithdraw(beneficiaries, rewards, {
        from: owner,
      })

      // Verify funds got transferred to the LP Rewards contracts.
      expect(
        await keepToken.balanceOf(lpReward1.address),
        "invalid balance of LP Rewards 1"
      ).to.eq.BN(rewards[0])
      expect(
        await keepToken.balanceOf(lpReward2.address),
        "invalid balance of LP Rewards 2"
      ).to.eq.BN(rewards[1])
      expect(
        await keepToken.balanceOf(lpReward3.address),
        "invalid balance of LP Rewards 3"
      ).to.eq.BN(rewards[2])
    })

    async function newLPRewardsBeneficiary(
      destinationRewardsContractAddress,
      fundingEscrowAddress
    ) {
      const beneficiary = await LPRewardsEscrowBeneficiary.new(
        keepToken.address,
        destinationRewardsContractAddress,
        {from: owner}
      )
      await beneficiary.transferOwnership(fundingEscrowAddress, {
        from: owner,
      })

      return beneficiary
    }
  })

  async function mintAndApproveWrappedTokens(token, address, staker, amount) {
    await token.mint(staker, amount)
    await token.approve(address, amount, {from: staker})
  }

  async function fundKEEPReward(address, amount) {
    await keepToken.approveAndCall(address, amount, "0x0", {
      from: keepTokenContractOwner,
    })
  }

  async function timeIncreaseTo(seconds) {
    const delay = 10 - new Date().getMilliseconds()
    await new Promise((resolve) => setTimeout(resolve, delay))
    await time.increaseTo(seconds)
  }
})
