import { callWithRetry } from "./contract-helper.js"

export default class RewardsCalculator {
  constructor(context, interval, minimumStake) {
    this.context = context
    this.interval = interval
    this.ethScoreThreshold = 3000
  }

  static async initialize(context, interval, operatorsParameters) {
    const rewardsCalculator = new RewardsCalculator(context, interval)

    await rewardsCalculator.calculateOperatorsRewards(operatorsParameters)

    return rewardsCalculator
  }

  async calculateOperatorsRewards(operatorsParameters) {
    const minimumStake = await this.getMinimumStake()

    const operatorsRewardsFactors = []

    for (const operatorParameters of operatorsParameters) {
      const { keepStaked, ethTotal } = operatorParameters.operatorAssets

      const ethScore = this.calculateETHScore(ethTotal)
      const boost = this.calculateBoost(keepStaked, ethTotal, minimumStake)
      const rewardWeight = ethScore * boost

      operatorsRewardsFactors.push({
        operator: operatorParameters.operator,
        ethScore,
        boost,
        rewardWeight,
      })
    }

    const rewardWeightSum = operatorsRewardsFactors.reduce(
      (accumulator, factors) => accumulator + factors.rewardWeight,
      0
    )

    console.log(`Rewards weight sum ${Math.round(rewardWeightSum * 100) / 100}`)

    const operatorsRewards = []

    for (const operatorRewardsFactors of operatorsRewardsFactors) {
      const { rewardWeight } = operatorRewardsFactors
      const rewardRatio =
        rewardWeightSum > 0 ? rewardWeight / rewardWeightSum : 0

      const totalRewards = this.interval.totalRewards * rewardRatio

      operatorsRewards.push(
        new OperatorRewards(
          operatorRewardsFactors.operator,
          operatorRewardsFactors.ethScore,
          operatorRewardsFactors.boost,
          rewardWeight,
          totalRewards
        )
      )
    }

    this.operatorsRewards = operatorsRewards
  }

  async getMinimumStake() {
    const { contracts, web3 } = this.context
    const tokenStaking = await contracts.TokenStaking.deployed()

    const minimumStake = await callWithRetry(
      tokenStaking.methods.minimumStake()
    )

    return web3.utils
      .toBN(minimumStake)
      .div(web3.utils.toBN("1000000000000000000"))
      .toNumber()
  }

  calculateETHScore(ethTotal) {
    if (ethTotal < this.ethScoreThreshold) {
      return ethTotal
    }

    return (
      2 * Math.sqrt(this.ethScoreThreshold * ethTotal) - this.ethScoreThreshold
    )
  }

  calculateBoost(keepStaked, ethTotal, minimumStake) {
    const firstFactor = keepStaked / minimumStake
    const secondFactor =
      ethTotal > 0 ? Math.sqrt(keepStaked / (ethTotal * 500)) : 0
    return 1 + Math.min(firstFactor, secondFactor)
  }

  getOperatorRewards(operator) {
    return this.operatorsRewards.find(
      (operatorRewards) => operatorRewards.operator === operator
    )
  }
}

function OperatorRewards(
  operator,
  ethScore,
  boost,
  rewardWeight,
  totalRewards
) {
  ;(this.operator = operator),
    (this.ethScore = ethScore),
    (this.boost = boost),
    (this.rewardWeight = rewardWeight),
    (this.totalRewards = totalRewards)
}
