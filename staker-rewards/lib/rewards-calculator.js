export default class RewardsCalculator {
  constructor() {}

  static initialize(operatorsParameters, interval) {
    // TODO: Calculate all rewards
    return new RewardsCalculator()
  }

  calculateOperatorRewards(operator) {
    // TODO: Fetch rewards for given operator
    return new OperatorRewards()
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
