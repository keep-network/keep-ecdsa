import clc from "cli-color"
import BlockByDate from "ethereum-block-by-date"

import Context from "./lib/context.js"
import FraudDetector from "./lib/fraud-detector.js"
import SLACalculator from "./lib/sla-calculator.js"
import AssetsCalculator from "./lib/assets-calculator.js"

async function run() {
  try {
    // URL to the websocket endpoint of the Ethereum node.
    const ethHostname = process.env.ETH_HOSTNAME
    // Start of the interval passed as UNIX timestamp.
    const intervalStart = process.argv[2]
    // End of the interval passed as UNIX timestamp.
    const intervalEnd = process.argv[3]
    // Determines whether debug logs should be disabled.
    const isDebugDisabled = process.env.DEBUG !== "on"
    // Determines whether the cache refresh process should be performed.
    // This option should be used only for development purposes. If the
    // cache refresh is disabled, rewards distribution may be calculated
    // based on outdated information from the chain.
    const isCacheRefreshEnabled = process.env.CACHE_REFRESH !== "off"

    if (!ethHostname) {
      console.error(clc.red("Please provide ETH_HOSTNAME value"))
      return
    }

    const interval = {
      start: parseInt(intervalStart),
      end: parseInt(intervalEnd),
    }

    validateIntervalTimestamps(interval)

    if (isDebugDisabled) {
      console.debug = function () {}
    }

    const context = await Context.initialize(ethHostname)

    await determineIntervalBlockspan(context, interval)

    if (isCacheRefreshEnabled) {
      console.log("Refreshing keeps cache...")
      await context.cache.refresh()
    }

    const operatorsRewards = await calculateOperatorsRewards(context, interval)

    console.table(operatorsRewards)
  } catch (error) {
    throw new Error(error)
  }
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
  const slaCalculator = await SLACalculator.initialize(context, interval)
  const assetsCalculator = await AssetsCalculator.initialize(context, interval)

  const operatorsRewards = []

  for (const operator of getOperators(cache)) {
    const isFraudulent = await fraudDetector.isOperatorFraudulent(operator)
    const operatorSLA = slaCalculator.calculateOperatorSLA(operator)
    const operatorAssets = await assetsCalculator.calculateOperatorAssets(
      operator
    )

    operatorsRewards.push(
      new OperatorRewards(operator, isFraudulent, operatorSLA, operatorAssets)
    )
  }

  return operatorsRewards
}

function getOperators(cache) {
  const operators = new Set()

  cache
    .getKeeps()
    .forEach((keep) => keep.members.forEach((member) => operators.add(member)))

  return operators
}

function OperatorRewards(operator, isFraudulent, operatorSLA, operatorAssets) {
  ;(this.operator = operator),
    (this.isFraudulent = isFraudulent),
    (this.keygenCount = operatorSLA.keygenCount),
    (this.keygenFailCount = operatorSLA.keygenFailCount),
    (this.keygenSLA = operatorSLA.keygenSLA),
    (this.signatureCount = operatorSLA.signatureCount),
    (this.signatureFailCount = operatorSLA.signatureFailCount),
    (this.signatureSLA = operatorSLA.signatureSLA),
    (this.keepStaked = operatorAssets.keepStaked),
    (this.ethBonded = roundFloat(operatorAssets.ethBonded)),
    (this.ethUnbonded = roundFloat(operatorAssets.ethUnbonded)),
    (this.ethTotal = roundFloat(operatorAssets.ethTotal))
}

function roundFloat(number) {
  return Math.round(number * 100) / 100
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
