import { callWithRetry, getPastEvents } from "./contract-helper.js"
import BigNumber from "bignumber.js"

export default class AssetsCalculator {
  constructor(
    context,
    interval,
    tokenStaking,
    bondedECDSAKeepFactory,
    keepBonding,
    sortitionPoolAddress
  ) {
    this.context = context
    this.interval = interval
    this.tokenStaking = tokenStaking
    this.bondedECDSAKeepFactory = bondedECDSAKeepFactory
    this.keepBonding = keepBonding
    this.sortitionPoolAddress = sortitionPoolAddress
  }

  static async initialize(context, interval) {
    const { contracts } = context

    const bondedECDSAKeepFactory = await contracts.BondedECDSAKeepFactory.deployed()

    const sortitionPoolAddress = await callWithRetry(
      bondedECDSAKeepFactory.methods.getSortitionPool(
        contracts.sanctionedApplicationAddress
      )
    )

    const assetsCalculator = new AssetsCalculator(
      context,
      interval,
      await contracts.TokenStaking.deployed(),
      bondedECDSAKeepFactory,
      await contracts.KeepBonding.deployed(),
      sortitionPoolAddress
    )

    await assetsCalculator.fetchBondEvents()
    await assetsCalculator.fetchUnbondedValueWithdrawnEvents()

    return assetsCalculator
  }

  async calculateOperatorAssets(operator, operatorRequirements = {}) {
    if (process.env.NODE_ENV !== "test") {
      console.log(
        `Calculating KEEP and ETH under management for operator ${operator}`
      )
    }

    const keepStaked = await this.calculateKeepStaked(operator)
    const ethBonded = await this.calculateETHBonded(
      operator,
      this.bondEventsAtStart
    )
    const ethUnbonded = await this.calculateETHUnbonded(operator)
    const ethWithdrawn = await this.calculateETHWithdrawn(operator)
    let ethTotal = ethBonded.plus(ethUnbonded).minus(ethWithdrawn)

    const isUndelegating = await this.isUndelegatingOperator(operator)
    if (
      isUndelegating ||
      operatorRequirements.poolDeauthorizedInInterval === true ||
      operatorRequirements.poolAuthorizedAtStart === false ||
      operatorRequirements.poolRequirementFulfilledAtStart === false
    ) {
      const ethBondedAtEnd = await this.calculateETHBonded(
        operator,
        this.bondEventsAtEnd
      )

      ethTotal = BigNumber.min(ethBonded, ethBondedAtEnd)
    }

    if (ethTotal.isLessThan(new BigNumber(0))) {
      ethTotal = new BigNumber(0)
    }

    return new OperatorAssets(
      operator,
      keepStaked,
      ethBonded,
      ethUnbonded,
      ethWithdrawn,
      ethTotal,
      isUndelegating
    )
  }

  async fetchBondEvents() {
    const fromBlock = this.context.contracts.factoryDeploymentBlock
    const toBlock = this.interval.endBlock

    const eventBySortitionPool = (event) =>
      event.returnValues.sortitionPool === this.sortitionPoolAddress

    const getBondEvents = async (eventName) =>
      await getPastEvents(
        this.context.web3,
        this.keepBonding,
        eventName,
        fromBlock,
        toBlock
      )

    this.bondEventsAtEnd = {
      bondCreatedEvents: (await getBondEvents("BondCreated")).filter(
        eventBySortitionPool
      ),
      bondReleasedEvents: await getBondEvents("BondReleased"),
      bondSeizedEvents: await getBondEvents("BondSeized"),
    }

    const eventAtIntervalStart = (event) =>
      event.blockNumber <= this.interval.startBlock

    this.bondEventsAtStart = {
      bondCreatedEvents: this.bondEventsAtEnd.bondCreatedEvents.filter(
        eventAtIntervalStart
      ),
      bondReleasedEvents: this.bondEventsAtEnd.bondReleasedEvents.filter(
        eventAtIntervalStart
      ),
      bondSeizedEvents: this.bondEventsAtEnd.bondSeizedEvents.filter(
        eventAtIntervalStart
      ),
    }
  }

  async fetchUnbondedValueWithdrawnEvents() {
    // We are interested in all unbonded value withdrawals which occurred
    // within the given interval.
    const fromBlock = this.interval.startBlock
    const toBlock = this.interval.endBlock

    this.unbondedValueWithdrawnEvents = await getPastEvents(
      this.context.web3,
      this.keepBonding,
      "UnbondedValueWithdrawn",
      fromBlock,
      toBlock
    )
  }

  async calculateKeepStaked(operator) {
    const block = this.interval.endBlock

    const operatorContract = this.bondedECDSAKeepFactory.options.address

    let keepStaked = await callWithRetry(
      this.tokenStaking.methods.activeStake(operator, operatorContract),
      block
    )

    // If keep staked (active stake) checked against the factory contract is
    // zero, there is a possibility the operator undelegated. We need to check
    // whether it still has a locked active stake in one of the keep contracts.
    // If so, we count that stake as keep staked.
    if (keepStaked === "0") {
      const locks = await callWithRetry(
        this.tokenStaking.methods.getLocks(operator),
        block
      )

      for (let i = 0; i < locks.creators.length; i++) {
        const keepOpenedTimestamp = await callWithRetry(
          this.bondedECDSAKeepFactory.methods.getKeepOpenedTimestamp(
            locks.creators[i]
          ),
          block
        )

        if (keepOpenedTimestamp > 0) {
          keepStaked = await callWithRetry(
            this.tokenStaking.methods.balanceOf(operator),
            block
          )
          break
        }
      }
    }

    return new BigNumber(keepStaked)
  }

  async calculateETHBonded(operator, bondEvents) {
    const eventByOperator = (event) => event.returnValues.operator === operator
    const eventByReferenceID = (referenceID) => {
      return (event) => event.returnValues.referenceID === referenceID
    }

    const bondCreatedEvents = bondEvents.bondCreatedEvents.filter(
      eventByOperator
    )
    const bondReleasedEvents = bondEvents.bondReleasedEvents.filter(
      eventByOperator
    )
    const bondSeizedEvents = bondEvents.bondSeizedEvents.filter(eventByOperator)

    let weiBonded = new BigNumber(0)

    for (const bondCreatedEvent of bondCreatedEvents) {
      let bondAmount = new BigNumber(bondCreatedEvent.returnValues.amount)
      const referenceID = bondCreatedEvent.returnValues.referenceID

      // Check if the bond has been released.
      if (bondReleasedEvents.find(eventByReferenceID(referenceID))) {
        // If the bond has been released, don't count its amount
        continue
      }

      // Check if the bond has been seized
      const bondSeizedEvent = bondSeizedEvents.find(
        eventByReferenceID(referenceID)
      )
      if (bondSeizedEvent) {
        // If the bond has been seized, subtract the seized amount
        const seizedAmount = new BigNumber(bondSeizedEvent.returnValues.amount)
        bondAmount = bondAmount.minus(seizedAmount)
      }

      weiBonded = weiBonded.plus(bondAmount)
    }

    return weiBonded
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

    return new BigNumber(weiUnbonded)
  }

  async calculateETHWithdrawn(operator) {
    return this.unbondedValueWithdrawnEvents
      .filter((event) => event.returnValues.operator === operator)
      .reduce(
        (accumulator, event) =>
          accumulator.plus(new BigNumber(event.returnValues.amount)),
        new BigNumber(0)
      )
  }

  async isUndelegatingOperator(operator) {
    const tokenStaking = await this.context.contracts.TokenStaking.deployed()

    const delegationInfo = await callWithRetry(
      tokenStaking.methods.getDelegationInfo(operator),
      this.interval.startBlock
    )

    return delegationInfo.undelegatedAt !== "0"
  }
}

function OperatorAssets(
  address,
  keepStaked,
  ethBonded,
  ethUnbonded,
  ethWithdrawn,
  ethTotal,
  isUndelegating
) {
  ;(this.address = address),
    (this.keepStaked = keepStaked),
    (this.ethBonded = ethBonded),
    (this.ethUnbonded = ethUnbonded),
    (this.ethWithdrawn = ethWithdrawn),
    (this.ethTotal = ethTotal),
    (this.isUndelegating = isUndelegating)
}
