import clc from "cli-color"
import { callWithRetry } from "./contract-helper.js"
import BigNumber from "bignumber.js"

export default class RewardsCalculator {
  constructor(context, interval) {
    this.context = context
    this.interval = interval
    this.ethScoreThreshold = new BigNumber(3000).multipliedBy(
      new BigNumber(1e18)
    ) // 3000 ETH
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
      const rewardWeight = ethScore.multipliedBy(boost)

      operatorsRewardsFactors.push({
        operator: operatorParameters.operator,
        ethScore,
        boost,
        rewardWeight,
      })
    }

    const rewardWeightSum = operatorsRewardsFactors.reduce(
      (accumulator, factors) => accumulator.plus(factors.rewardWeight),
      new BigNumber(0)
    )

    console.log(clc.yellow(`Rewards weight sum ${rewardWeightSum.toFixed(2)}`))

    const operatorsRewards = []

    for (const operatorRewardsFactors of operatorsRewardsFactors) {
      const { rewardWeight } = operatorRewardsFactors
      const rewardRatio = rewardWeightSum.isGreaterThan(new BigNumber(0))
        ? rewardWeight.dividedBy(rewardWeightSum)
        : new BigNumber(0)

      const totalRewards = this.interval.totalRewards.multipliedBy(rewardRatio)

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

    const totalRewardsSum = operatorsRewards.reduce(
      (accumulator, rewards) =>
        accumulator.plus(rewards.totalRewards.toFixed(0)),
      new BigNumber(0)
    )

    console.log(clc.yellow(`Total rewards sum ${totalRewardsSum}`))

    this.operatorsRewards = operatorsRewards
  }

  async getMinimumStake() {
    const tokenStaking = await this.context.contracts.TokenStaking.deployed()

    const minimumStake = await callWithRetry(
      tokenStaking.methods.minimumStake()
    )

    return new BigNumber(minimumStake)
  }

  calculateETHScore(ethTotal) {
    if (ethTotal.isLessThan(this.ethScoreThreshold)) {
      return ethTotal
    }

    const sqrt = this.ethScoreThreshold.multipliedBy(ethTotal).squareRoot()

    return new BigNumber(2).multipliedBy(sqrt).minus(this.ethScoreThreshold)
  }

  calculateBoost(keepStaked, ethTotal, minimumStake) {
    const a = keepStaked.dividedBy(minimumStake)
    const b = ethTotal.isGreaterThan(new BigNumber(0))
      ? keepStaked
          .dividedBy(ethTotal.multipliedBy(new BigNumber(500)))
          .squareRoot()
      : new BigNumber(0)

    return new BigNumber(1).plus(BigNumber.minimum(a, b))
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
