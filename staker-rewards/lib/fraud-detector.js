export default class FraudDetector {
  constructor(context, tokenStaking, factoryDeploymentBlock) {
    this.context = context

    this.tokenStaking = tokenStaking

    this.factoryDeploymentBlock = factoryDeploymentBlock
  }

  static async initialize(context) {
    const { TokenStaking, factoryDeploymentBlock } = context.contracts

    const tokenStaking = await TokenStaking.deployed()

    return new FraudDetector(context, tokenStaking, factoryDeploymentBlock)
  }

  async isOperatorFraudulent(operator) {
    if (process.env.NODE_ENV !== "test") {
      console.log(`Checking fraudulent activity for operator ${operator}`)
    }

    // Get all slashing events for the operator.
    const events = await this.tokenStaking.getPastEvents("TokensSlashed", {
      fromBlock: this.factoryDeploymentBlock,
      toBlock: "latest",
      filter: { operator: operator },
    })

    // At this moment token slashing can only be originated from tBTC application
    // on fraud detection. Random Beacon does not use slashing but seizing. We
    // assume that any slashing event is related to a fraud detection for ECDSA
    // Keep.
    // This section has to be revisited in case of implementing additional usage
    // of slashing function.
    if (events.length > 0) {
      return true
    } else {
      return false
    }
  }
}
