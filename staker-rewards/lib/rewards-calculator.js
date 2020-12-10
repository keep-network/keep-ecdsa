import clc from "cli-color"
import { callWithRetry } from "./contract-helper.js"
import BigNumber from "bignumber.js"
import { noDecimalPlaces } from "./numbers.js"

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

      const requirementsViolations = this.checkRequirementsViolations(
        operatorParameters
      )

      const ethScore = this.calculateETHScore(ethTotal)
      const boost = this.calculateBoost(keepStaked, ethTotal, minimumStake)
      const rewardWeight = ethScore.multipliedBy(boost)

      operatorsRewardsFactors.push({
        operator: operatorParameters.operator,
        ethScore,
        boost,
        rewardWeight,
        requirementsViolations,
      })
    }

    const rewardWeightSum = operatorsRewardsFactors.reduce(
      (accumulator, factors) => accumulator.plus(factors.rewardWeight),
      new BigNumber(0)
    )

    console.log(
      clc.yellow(
        `Rewards weight sum ${rewardWeightSum.toFixed(
          noDecimalPlaces,
          BigNumber.ROUND_DOWN
        )}`
      )
    )

    const operatorsRewards = []

    for (const operatorRewardsFactors of operatorsRewardsFactors) {
      const { rewardWeight, requirementsViolations } = operatorRewardsFactors
      const rewardRatio = rewardWeightSum.isGreaterThan(new BigNumber(0))
        ? rewardWeight.dividedBy(rewardWeightSum)
        : new BigNumber(0)

      const totalRewards =
        requirementsViolations.length > 0
          ? new BigNumber(0)
          : this.interval.totalRewards.multipliedBy(rewardRatio)

      operatorsRewards.push(
        new OperatorRewards(
          operatorRewardsFactors.operator,
          operatorRewardsFactors.ethScore,
          operatorRewardsFactors.boost,
          rewardWeight,
          totalRewards,
          requirementsViolations
        )
      )
    }

    const totalRewardsSum = operatorsRewards.reduce(
      (accumulator, rewards) =>
        accumulator.plus(
          rewards.totalRewards.toFixed(noDecimalPlaces, BigNumber.ROUND_DOWN)
        ),
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

  checkRequirementsViolations(operatorParameters) {
    const violations = []
    const { isFraudulent, requirements, operatorSLA } = operatorParameters

    if (isFraudulent === true) {
      violations.push("isFraudulent")
    }

    if (requirements.factoryAuthorizedAtStart === false) {
      violations.push("factoryAuthorizedAtStart")
    }

    if (requirements.poolAuthorizedAtStart === false) {
      violations.push("poolAuthorizedAtStart")
    }

    if (requirements.poolDeauthorizedInInterval === true) {
      violations.push("poolDeauthorizedInInterval")
    }

    if (requirements.minimumStakeAtStart === false) {
      violations.push("minimumStakeAtStart")
    }

    if (requirements.poolRequirementFulfilledAtStart === false) {
      violations.push("poolRequirementFulfilledAtStart")
    }

    if (operatorSLA.keygenSLA !== "N/A") {
      if (operatorSLA.keygenCount < 10) {
        if (operatorSLA.keygenFailCount > 1) {
          violations.push("keygenSLA")
        }
      } else {
        if (operatorSLA.keygenSLA < 90) {
          violations.push("keygenSLA")
        }
      }
    }

    if (operatorSLA.signatureSLA !== "N/A") {
      if (operatorSLA.signatureCount < 20) {
        if (operatorSLA.signatureFailCount > 1) {
          violations.push("signatureSLA")
        }
      } else {
        if (operatorSLA.signatureSLA < 95) {
          violations.push("signatureSLA")
        }
      }
    }

    return violations
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
  totalRewards,
  requirementsViolations
) {
  ;(this.operator = operator),
    (this.ethScore = ethScore),
    (this.boost = boost),
    (this.rewardWeight = rewardWeight),
    (this.totalRewards = totalRewards),
    (this.requirementsViolations = requirementsViolations),
    (this.SLAViolated =
      requirementsViolations.filter((r) => r.includes("SLA")).length > 0)
}
