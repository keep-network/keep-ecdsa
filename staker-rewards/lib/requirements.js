import { callWithRetry } from "./contract-helper.js"

export default class Requirements {
  constructor(
    context,
    interval,
    bondedECDSAKeepFactory,
    keepBonding,
    sortitionPoolAddress
  ) {
    this.context = context
    this.interval = interval
    this.bondedECDSAKeepFactory = bondedECDSAKeepFactory
    this.keepBonding = keepBonding
    this.sortitionPoolAddress = sortitionPoolAddress
    this.operatorsDeauthorizedInInterval = new Set()
  }

  static async initialize(context, interval) {
    const { contracts } = context

    const bondedECDSAKeepFactory = await contracts.BondedECDSAKeepFactory.deployed()

    const sortitionPoolAddress = await callWithRetry(
      bondedECDSAKeepFactory.methods.getSortitionPool(
        contracts.sanctionedApplicationAddress
      )
    )

    const requirements = new Requirements(
      context,
      interval,
      bondedECDSAKeepFactory,
      await contracts.KeepBonding.deployed(),
      sortitionPoolAddress
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

  async checkAuthorizations(operator) {
    console.debug(`Checking authorizations for operator ${operator}`)

    // Authorizations at the interval start.
    const {
      wasFactoryAuthorized: factoryAuthorizedAtStart,
      wasSortitionPoolAuthorized: poolAuthorizedAtStart,
    } = await this.checkAuthorizationsAtIntervalStart(operator)

    // Deauthorizations during the interval.
    const poolDeauthorizedInInterval = await this.wasSortitionPoolDeauthorized(
      operator
    )

    return new OperatorAuthorizations(
      operator,
      factoryAuthorizedAtStart,
      poolAuthorizedAtStart,
      poolDeauthorizedInInterval
    )
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
}

export function OperatorAuthorizations(
  address,
  factoryAuthorizedAtStart,
  poolAuthorizedAtStart,
  poolDeauthorizedInInterval
) {
  ;(this.address = address),
    (this.factoryAuthorizedAtStart = factoryAuthorizedAtStart),
    (this.poolAuthorizedAtStart = poolAuthorizedAtStart),
    (this.poolDeauthorizedInInterval = poolDeauthorizedInInterval)
}
