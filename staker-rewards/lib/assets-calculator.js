import { callWithRetry, getPastEvents } from "./contract-helper.js"

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
        contracts.SanctionedApplication
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

    return assetsCalculator
  }

  async calculateOperatorAssets(operator) {
    if (process.env.NODE_ENV !== "test") {
      console.log(
        `Calculating KEEP and ETH under management for operator ${operator}`
      )
    }

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

  async fetchBondEvents() {
    // The amount of ETH under the system management
    // (the sum of bonded and unbonded ETH) for the operator is captured based
    // on the state at the beginning of the interval. To calculate this state,
    // we take all past BondCreated events from the moment the reward interval
    // started and cross-check them with BondReleased and BondSeized events.
    const fromBlock = this.context.contracts.factoryDeploymentBlock
    const toBlock = this.interval.startBlock

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

    this.bondEvents = {
      bondCreatedEvents: (await getBondEvents("BondCreated")).filter(
        eventBySortitionPool
      ),
      bondReleasedEvents: await getBondEvents("BondReleased"),
      bondSeizedEvents: await getBondEvents("BondSeized"),
    }
  }

  async calculateKeepStaked(operator) {
    const block = this.interval.endBlock

    const operatorContract = this.bondedECDSAKeepFactory.options.address

    const keepStaked = await callWithRetry(
      this.tokenStaking.methods.activeStake(operator, operatorContract),
      block
    )

    return keepStaked
  }

  async calculateETHBonded(operator) {
    const { web3 } = this.context

    const eventByOperator = (event) => event.returnValues.operator === operator
    const eventByReferenceID = (referenceID) => {
      return (event) => event.returnValues.referenceID === referenceID
    }

    const bondCreatedEvents = this.bondEvents.bondCreatedEvents.filter(
      eventByOperator
    )
    const bondReleasedEvents = this.bondEvents.bondReleasedEvents.filter(
      eventByOperator
    )
    const bondSeizedEvents = this.bondEvents.bondSeizedEvents.filter(
      eventByOperator
    )

    let weiBonded = web3.utils.toBN(0)

    for (const bondCreatedEvent of bondCreatedEvents) {
      let bondAmount = web3.utils.toBN(bondCreatedEvent.returnValues.amount)
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
        const seizedAmount = web3.utils.toBN(
          bondSeizedEvent.returnValues.amount
        )
        bondAmount = bondAmount.sub(seizedAmount)
      }

      weiBonded = weiBonded.add(bondAmount)
    }

    const ethBonded = this.context.web3.utils.fromWei(weiBonded, "ether")

    return parseFloat(ethBonded)
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

    return parseFloat(ethUnbonded)
  }
}

function OperatorAssets(address, keepStaked, ethBonded, ethUnbonded, ethTotal) {
  ;(this.address = address),
    (this.keepStaked = keepStaked),
    (this.ethBonded = ethBonded),
    (this.ethUnbonded = ethUnbonded),
    (this.ethTotal = ethTotal)
}
