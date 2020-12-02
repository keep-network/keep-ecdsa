import { callWithRetry } from "./contract-helper.js"

export default class AssetsCalculator {
  constructor(
    context,
    interval,
    tokenStaking,
    keepRandomBeaconOperator,
    bondedECDSAKeepFactory,
    keepBonding,
    sortitionPoolAddress
  ) {
    this.context = context
    this.interval = interval
    this.tokenStaking = tokenStaking
    this.keepRandomBeaconOperator = keepRandomBeaconOperator
    this.bondedECDSAKeepFactory = bondedECDSAKeepFactory
    this.keepBonding = keepBonding
    this.sortitionPoolAddress = sortitionPoolAddress
  }

  static async initialize(context, interval) {
    const { contracts } = context

    const bondedECDSAKeepFactory = await contracts.BondedECDSAKeepFactory.deployed()
    const tbtcSystem = await contracts.TBTCSystem.deployed()

    const sortitionPoolAddress = await callWithRetry(
      bondedECDSAKeepFactory.methods.getSortitionPool(
        tbtcSystem.options.address
      )
    )

    return new AssetsCalculator(
      context,
      interval,
      await contracts.TokenStaking.deployed(),
      await contracts.KeepRandomBeaconOperator.deployed(),
      bondedECDSAKeepFactory,
      await contracts.KeepBonding.deployed(),
      sortitionPoolAddress
    )
  }

  async calculateOperatorAssets(operator) {
    const keepStaked = await this.calculateKeepStaked(operator)
    const ethBonded = await this.calculateETHBonded(operator)
    const ethUnbonded = await this.calculateETHUnbonded(operator)
    const ethTotal = ethBonded + ethUnbonded

    return new OperatorAssets(
      operator,
      keepStaked,
      ethBonded,
      ethUnbonded,
      ethTotal
    )
  }

  async calculateKeepStaked(operator) {
    const block = this.interval.endBlock

    const operatorContract = this.keepRandomBeaconOperator.options.address

    const keepStaked = await callWithRetry(
      this.tokenStaking.methods.activeStake(operator, operatorContract),
      block
    )

    return keepStaked
  }

  async calculateETHBonded(operator) {
    // TODO: Implementation
    return 0
  }

  async calculateETHUnbonded(operator) {
    const block = this.interval.startBlock

    const bondCreator = this.bondedECDSAKeepFactory.options.address

    const weiUnbonded = await callWithRetry(
      this.keepBonding.methods.availableUnbondedValue(
        operator,
        bondCreator,
        this.sortitionPoolAddress
      ),
      block
    )

    const ethUnbonded = this.context.web3.utils.fromWei(weiUnbonded, "ether")

    return Math.round(parseFloat(ethUnbonded) * 100) / 100
  }
}

function OperatorAssets(address, keepStaked, ethBonded, ethUnbonded, ethTotal) {
  ;(this.address = address),
    (this.keepStaked = keepStaked),
    (this.ethBonded = ethBonded),
    (this.ethUnbonded = ethUnbonded),
    (this.ethTotal = ethTotal)
}
