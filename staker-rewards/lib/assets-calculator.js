import { callWithRetry } from "./contract-helper.js"

export default class AssetsCalculator {
  constructor(interval, tokenStaking, keepRandomBeaconOperator) {
    this.interval = interval
    this.tokenStaking = tokenStaking
    this.keepRandomBeaconOperator = keepRandomBeaconOperator
  }

  static async initialize(context, interval) {
    return new AssetsCalculator(
      interval,
      await context.contracts.TokenStaking.deployed(),
      await context.contracts.KeepRandomBeaconOperator.deployed()
    )
  }

  async calculateOperatorAssets(operator) {
    const keepStaked = await this.calculateKeepStaked(operator)

    return new OperatorAssets(operator, keepStaked)
  }

  async calculateKeepStaked(operator) {
    // TODO: Check at interval end block
    const block = "latest"

    return await callWithRetry(
      this.tokenStaking.methods.activeStake(
        operator,
        this.keepRandomBeaconOperator.options.address
      ),
      block
    )
  }
}

function OperatorAssets(address, keepStaked) {
  ;(this.address = address), (this.keepStaked = keepStaked)
}
