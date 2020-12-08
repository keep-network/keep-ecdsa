import clc from "cli-color"
import BlockByDate from "ethereum-block-by-date"
import BigNumber from "bignumber.js"

import Context from "./lib/context.js"
import FraudDetector from "./lib/fraud-detector.js"
import Requirements from "./lib/requirements.js"
import SLACalculator from "./lib/sla-calculator.js"
import AssetsCalculator from "./lib/assets-calculator.js"
import RewardsCalculator from "./lib/rewards-calculator.js"

async function run() {
  // URL to the websocket endpoint of the Ethereum node.
  const ethHostname = process.env.ETH_HOSTNAME
  // Start of the interval passed as UNIX timestamp.
  const intervalStart = process.argv[2]
  // End of the interval passed as UNIX timestamp.
  const intervalEnd = process.argv[3]
  // Total KEEP rewards distributed within the given interval passed as
  // 18-decimals number.
  const intervalTotalRewards = process.argv[4]
  // Determines whether debug logs should be disabled.
  const isDebugDisabled = process.env.DEBUG !== "on"
  // Determines whether the cache refresh process should be performed.
  // This option should be used only for development purposes. If the
  // cache refresh is disabled, rewards distribution may be calculated
  // based on outdated information from the chain.
  const isCacheRefreshEnabled = process.env.CACHE_REFRESH !== "off"
  // Tenderly API project URL. If not provided a default value for Thesis
  // Keep project will be used.
  const tenderlyProjectUrl = process.env.TENDERLY_PROJECT_URL
  // Access Token for Tenderly API used to fetch transactions from the chain.
  // Setting it is optional. If not set the script won't call Tenderly API
  // and rely on cached transactions.
  const tenderlyAccessToken = process.env.TENDERLY_ACCESS_TOKEN

  if (isDebugDisabled) {
    console.debug = function () {}
  }

  if (!ethHostname) {
    console.error(clc.red("Please provide ETH_HOSTNAME value"))
    return
  }

  const context = await Context.initialize(
    ethHostname,
    tenderlyProjectUrl,
    tenderlyAccessToken
  )

  const interval = {
    start: parseInt(intervalStart),
    end: parseInt(intervalEnd),
    totalRewards: new BigNumber(intervalTotalRewards),
  }

  validateIntervalTimestamps(interval)
  validateIntervalTotalRewards(interval)
  await determineIntervalBlockspan(context, interval)

  if (isCacheRefreshEnabled) {
    console.log("Refreshing keeps cache...")
    await context.cache.refresh()
  }

  const operatorsRewards = await calculateOperatorsRewards(context, interval)

  console.table(operatorsRewards)
}

function validateIntervalTimestamps(interval) {
  const startDate = new Date(interval.start * 1000)
  const endDate = new Date(interval.end * 1000)

  const isValidStartDate = startDate.getTime() > 0
  if (!isValidStartDate) {
    throw new Error("Invalid interval start timestamp")
  }

  const isValidEndDate = endDate.getTime() > 0
  if (!isValidEndDate) {
    throw new Error("Invalid interval end timestamp")
  }

  const isEndAfterStart = endDate.getTime() > startDate.getTime()
  if (!isEndAfterStart) {
    throw new Error(
      "Interval end timestamp should be bigger than the interval start"
    )
  }

  console.log(clc.green(`Interval start timestamp: ${startDate.toISOString()}`))
  console.log(clc.green(`Interval end timestamp: ${endDate.toISOString()}`))
}

function validateIntervalTotalRewards(interval) {
  if (interval.totalRewards.isLessThanOrEqualTo(new BigNumber(0))) {
    throw new Error("Interval total rewards should be set")
  }

  console.log(
    clc.green(
      `Interval total rewards: ${display18DecimalsValue(
        interval.totalRewards
      )} KEEP`
    )
  )
}

async function determineIntervalBlockspan(context, interval) {
  const blockByDate = new BlockByDate(context.web3)

  interval.startBlock = (await blockByDate.getDate(interval.start * 1000)).block
  interval.endBlock = (await blockByDate.getDate(interval.end * 1000)).block

  console.log(clc.green(`Interval start block: ${interval.startBlock}`))
  console.log(clc.green(`Interval end block: ${interval.endBlock}`))
}

async function calculateOperatorsRewards(context, interval) {
  const { cache } = context

  const fraudDetector = await FraudDetector.initialize(context)
  const requirements = await Requirements.initialize(context, interval)
  const slaCalculator = await SLACalculator.initialize(context, interval)
  const assetsCalculator = await AssetsCalculator.initialize(context, interval)

  const operatorsParameters = []

  for (const operator of getOperators(cache)) {
    const isFraudulent = await fraudDetector.isOperatorFraudulent(operator)
    const operatorAuthorizations = await requirements.checkAuthorizations(
      operator
    )
    const operatorSLA = slaCalculator.calculateOperatorSLA(operator)
    const operatorAssets = await assetsCalculator.calculateOperatorAssets(
      operator
    )

    operatorsParameters.push(
      new OperatorParameters(
        operator,
        isFraudulent,
        operatorAuthorizations,
        operatorSLA,
        operatorAssets
      )
    )
  }

  const rewardsCalculator = await RewardsCalculator.initialize(
    context,
    interval,
    operatorsParameters
  )

  const operatorsSummary = []

  for (const operatorParameters of operatorsParameters) {
    const { operator } = operatorParameters
    const operatorRewards = rewardsCalculator.getOperatorRewards(operator)

    operatorsSummary.push(
      new OperatorSummary(operator, operatorParameters, operatorRewards)
    )
  }

  return operatorsSummary
}

// TODO: Change the way operators are fetched. Currently only the ones which
//  have members in existing keeps are taken into account. Instead of that,
//  we should take all operators which are registered in the sorition pool.
function getOperators(cache) {
  const operators = new Set()

  cache
    .getKeeps()
    .forEach((keep) => keep.members.forEach((member) => operators.add(member)))

  return operators
}

function OperatorParameters(
  operator,
  isFraudulent,
  authorizations,
  operatorSLA,
  operatorAssets
) {
  ;(this.operator = operator),
    (this.isFraudulent = isFraudulent),
    (this.authorizations = authorizations)((this.operatorSLA = operatorSLA)),
    (this.operatorAssets = operatorAssets)
}

function OperatorSummary(operator, operatorParameters, operatorRewards) {
  ;(this.operator = operator),
    (this.isFraudulent = operatorParameters.isFraudulent),
    (this.factoryAuthorizedAtStart =
      operatorParameters.authorizations.factoryAuthorizedAtStart),
    (this.poolAuthorizedAtStart =
      operatorParameters.authorizations.poolAuthorizedAtStart),
    (this.poolDeauthorizedInInterval =
      operatorParameters.authorizations.poolDeauthorizedInInterval),
    (this.keygenCount = operatorParameters.operatorSLA.keygenCount),
    (this.keygenFailCount = operatorParameters.operatorSLA.keygenFailCount),
    (this.keygenSLA = operatorParameters.operatorSLA.keygenSLA),
    (this.signatureCount = operatorParameters.operatorSLA.signatureCount),
    (this.signatureFailCount =
      operatorParameters.operatorSLA.signatureFailCount),
    (this.signatureSLA = operatorParameters.operatorSLA.signatureSLA),
    (this.keepStaked = display18DecimalsValue(
      operatorParameters.operatorAssets.keepStaked
    )),
    (this.ethBonded = display18DecimalsValue(
      operatorParameters.operatorAssets.ethBonded
    )),
    (this.ethUnbonded = display18DecimalsValue(
      operatorParameters.operatorAssets.ethUnbonded
    )),
    (this.ethTotal = display18DecimalsValue(
      operatorParameters.operatorAssets.ethTotal
    )),
    (this.ethScore = display18DecimalsValue(operatorRewards.ethScore)),
    (this.boost = operatorRewards.boost.toFixed(2)),
    (this.rewardWeight = display18DecimalsValue(operatorRewards.rewardWeight)),
    (this.totalRewards = display18DecimalsValue(operatorRewards.totalRewards))
}

function display18DecimalsValue(value) {
  return value.dividedBy(new BigNumber(1e18)).toFixed(2)
}

run()
  .then((result) => {
    console.log(
      clc.green(
        "Staker rewards distribution calculations completed successfully"
      )
    )

    process.exit(0)
  })
  .catch((error) => {
    console.trace(
      clc.red(
        "Staker rewards distribution calculations errored out with error: "
      ),
      error
    )

    process.exit(1)
  })
