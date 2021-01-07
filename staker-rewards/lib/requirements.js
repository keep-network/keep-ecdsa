import { callWithRetry } from "./contract-helper.js"
import BigNumber from "bignumber.js"
import { shorten18Decimals } from "./numbers.js"

export default class Requirements {
  constructor(
    context,
    interval,
    bondedECDSAKeepFactory,
    keepBonding,
    sanctionedApplicationAddress,
    sortitionPoolAddress,
    minimumBondableValueAtStart
  ) {
    this.context = context
    this.interval = interval
    this.bondedECDSAKeepFactory = bondedECDSAKeepFactory
    this.keepBonding = keepBonding
    this.sanctionedApplicationAddress = sanctionedApplicationAddress
    this.sortitionPoolAddress = sortitionPoolAddress
    this.minimumBondableValueAtStart = minimumBondableValueAtStart
    this.operatorsDeauthorizedInInterval = new Set()
  }

  static async initialize(context, interval) {
    const { contracts } = context

    const sanctionedApplicationAddress = contracts.sanctionedApplicationAddress

    const bondedECDSAKeepFactory = await contracts.BondedECDSAKeepFactory.deployed()

    const sortitionPoolAddress = await callWithRetry(
      bondedECDSAKeepFactory.methods.getSortitionPool(
        sanctionedApplicationAddress
      )
    )

    const bondedSortitionPool = await contracts.BondedSortitionPool.at(
      sortitionPoolAddress
    )
    const minimumBondableValueAtStart = new BigNumber(
      await callWithRetry(
        bondedSortitionPool.methods.getMinimumBondableValue(),
        interval.startBlock
      )
    )

    console.log(
      `Minimum Bondable Value at interval start: ${shorten18Decimals(
        minimumBondableValueAtStart
      )} ether`
    )

    const requirements = new Requirements(
      context,
      interval,
      bondedECDSAKeepFactory,
      await contracts.KeepBonding.deployed(),
      sanctionedApplicationAddress,
      sortitionPoolAddress,
      minimumBondableValueAtStart
    )

    await requirements.checkDeauthorizations()

    return requirements
  }

  async checkDeauthorizations() {
    // If Tenderly API is available refresh cached data.
    if (this.context.tenderly) {
      await this.refreshDeauthorizationsCache()
    }

    // Fetch deauthorization transactions from cache.
    // We expect only successful transactions to be returned, which means we don't
    // need to double check authorizer vs operator as this was already handled
    // by the contract on the function call.
    const deauthorizations = this.context.cache.getTransactionFunctionCalls(
      this.keepBonding.options.address,
      "deauthorizeSortitionPoolContract"
    )

    console.debug(
      `Found ${deauthorizations.length} sortition pool contract deauthorizations`
    )

    if (deauthorizations && deauthorizations.length > 0) {
      for (let i = 0; i < deauthorizations.length; i++) {
        const transaction = deauthorizations[i]

        console.debug(`Checking transaction ${transaction.hash}`)

        // Starting block is included in the interval, end block is not included
        // to avoid overlapping: `start <= interval < end`.
        if (
          transaction.block_number < this.interval.startBlock ||
          transaction.block_number >= this.interval.endBlock
        ) {
          console.debug(
            `Skipping transaction made in block ${transaction.block_number}`
          )
          continue
        }

        const transactionOperator = transaction.decoded_input[0].value
        const transactionSortitionPool = transaction.decoded_input[1].value

        if (
          transactionSortitionPool.toLowerCase() !=
          this.sortitionPoolAddress.toLowerCase()
        ) {
          console.debug(
            `Skipping transaction for sortition pool ${transactionSortitionPool}`
          )
          continue
        }

        console.log(
          `Discovered deauthorization for operator ${transactionOperator}` +
            ` in transaction ${transaction.hash} in the current interval`
        )

        this.operatorsDeauthorizedInInterval.add(
          transactionOperator.toLowerCase()
        )
      }
    }
  }

  async refreshDeauthorizationsCache() {
    console.log("Refreshing cached sortition pool deauthorization transactions")

    let data
    try {
      data = await this.context.tenderly.getFunctionCalls(
        this.keepBonding.options.address,
        "deauthorizeSortitionPoolContract(address,address)"
      )
    } catch (error) {
      throw new Error(
        `Failed to fetch function calls from Tenderly: ${error.message}`
      )
    }

    console.debug(`Fetched ${data.length} transactions from Tenderly`)

    const transactions = data
      .filter((tx) => tx.status === true) // filter only successful transactions
      .map((tx) => {
        return {
          hash: tx.hash,
          from: tx.from,
          to: tx.to,
          block_number: tx.block_number,
          method: tx.method,
          decoded_input: tx.decoded_input,
        }
      })

    this.context.cache.storeTransactions(transactions)
  }

  async check(operator) {
    console.log(`Checking requirements for operator ${operator}`)

    const {
      factoryAuthorizedAtStart,
      poolAuthorizedAtStart,
      poolDeauthorizedInInterval,
    } = await this.checkAuthorizations(operator)

    const minimumStakeAtStart = await this.checkMinimumStakeAtIntervalStart(
      operator
    )

    const poolRequirementFulfilledAtStart = await this.checkWasInPoolIfRequiredAtIntervalStart(
      operator
    )

    return new OperatorRequirements(
      operator,
      factoryAuthorizedAtStart,
      poolAuthorizedAtStart,
      poolDeauthorizedInInterval,
      minimumStakeAtStart,
      poolRequirementFulfilledAtStart
    )
  }

  async checkAuthorizations(operator) {
    // Authorizations at the interval start.
    const {
      wasFactoryAuthorized: factoryAuthorizedAtStart,
      wasSortitionPoolAuthorized: poolAuthorizedAtStart,
    } = await this.checkAuthorizationsAtIntervalStart(operator)

    // Deauthorizations during the interval.
    const poolDeauthorizedInInterval = await this.wasSortitionPoolDeauthorized(
      operator
    )

    return {
      factoryAuthorizedAtStart,
      poolAuthorizedAtStart,
      poolDeauthorizedInInterval,
    }
  }

  async checkAuthorizationsAtIntervalStart(operator) {
    // Operator contract
    const wasFactoryAuthorized = await callWithRetry(
      this.bondedECDSAKeepFactory.methods.isOperatorAuthorized(operator),
      this.interval.startBlock
    )

    // Sortition pool
    const wasSortitionPoolAuthorized = await callWithRetry(
      this.keepBonding.methods.hasSecondaryAuthorization(
        operator,
        this.sortitionPoolAddress
      ),
      this.interval.startBlock
    )

    return { wasFactoryAuthorized, wasSortitionPoolAuthorized }
  }

  async wasSortitionPoolDeauthorized(operator) {
    return this.operatorsDeauthorizedInInterval.has(operator.toLowerCase())
  }

  async checkMinimumStakeAtIntervalStart(operator) {
    return await callWithRetry(
      this.bondedECDSAKeepFactory.methods.hasMinimumStake(operator),
      this.interval.startBlock
    )
  }

  // If the operator has at least minimum unbonded value available they have
  // to be registered in the sortition pool. Operators who are not in the sortition
  // pool because all of their ether is bonded are still getting rewards because
  // that ether is still under the system’s management.
  async checkWasInPoolIfRequiredAtIntervalStart(operator) {
    const unbondedValueAtStart = new BigNumber(
      await callWithRetry(
        this.keepBonding.methods.unbondedValue(operator),
        this.interval.startBlock
      )
    )

    const hadMinimumBondableValueAtStart = unbondedValueAtStart.gte(
      this.minimumBondableValueAtStart
    )

    if (hadMinimumBondableValueAtStart) {
      const wasRegisteredAtStart = await callWithRetry(
        this.bondedECDSAKeepFactory.methods.isOperatorRegistered(
          operator,
          this.sanctionedApplicationAddress
        ),
        this.interval.startBlock
      )

      return wasRegisteredAtStart
    } else {
      return true
    }
  }
}

export function OperatorRequirements(
  address,
  factoryAuthorizedAtStart,
  poolAuthorizedAtStart,
  poolDeauthorizedInInterval,
  minimumStakeAtStart,
  poolRequirementFulfilledAtStart
) {
  ;(this.address = address),
    (this.factoryAuthorizedAtStart = factoryAuthorizedAtStart),
    (this.poolAuthorizedAtStart = poolAuthorizedAtStart),
    (this.minimumStakeAtStart = minimumStakeAtStart),
    (this.poolRequirementFulfilledAtStart = poolRequirementFulfilledAtStart),
    (this.poolDeauthorizedInInterval = poolDeauthorizedInInterval)
}
