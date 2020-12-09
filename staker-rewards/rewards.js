import clc from "cli-color"
import BlockByDate from "ethereum-block-by-date"
import BigNumber from "bignumber.js"

import Context from "./lib/context.js"
import FraudDetector from "./lib/fraud-detector.js"
import SLACalculator from "./lib/sla-calculator.js"
import AssetsCalculator from "./lib/assets-calculator.js"
import RewardsCalculator from "./lib/rewards-calculator.js"

import * as fs from "fs"

const decimalPlaces = 2

async function run() {
  try {
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

    if (isDebugDisabled) {
      console.debug = function () {}
    }

    if (!ethHostname) {
      console.error(clc.red("Please provide ETH_HOSTNAME value"))
      return
    }

    const context = await Context.initialize(ethHostname)

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

    if (process.env.OUTPUT_MODE === "text") {
      const format = {
        groupSeparator: "",
        decimalSeparator: ".",
      }

      const rewards = {}

      for (const operatorReward of operatorsRewards) {
        console.log(
          `${operatorReward.operator}
           ${operatorReward.isFraudulent} 
           ${operatorReward.keygenCount}
           ${operatorReward.keygenFailCount} 
           ${operatorReward.keygenSLA} 
           ${operatorReward.signatureCount} 
           ${operatorReward.signatureFailCount}
           ${operatorReward.signatureSLA} 
           ${operatorReward.keepStaked.toFormat(format)} 
           ${operatorReward.ethBonded.toFormat(format)} 
           ${operatorReward.ethUnbonded.toFormat(format)}
           ${operatorReward.ethTotal.toFormat(format)} 
           ${operatorReward.ethScore.toFormat(decimalPlaces, format)} 
           ${operatorReward.boost.toFormat(decimalPlaces, format)} 
           ${operatorReward.rewardWeight.toFormat(decimalPlaces, format)} 
           ${operatorReward.totalRewards.toFormat(decimalPlaces, format)}
          `.replace(/\s+/gm, " ")
        )

        if (operatorReward.totalRewards != 0) {
          let rewardsBN = new BigNumber(operatorReward.totalRewards.toFormat(0, format))
          rewards[operatorReward.operator] = rewardsBN.toString(16) // convert BN to hex
        }
      }

      writeOperatorsRewardsToFile(rewards)
    } else {
      console.table(operatorsRewards.map(shortenSummaryValues))
    }
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

function validateIntervalTotalRewards(interval) {
  if (interval.totalRewards.isLessThanOrEqualTo(new BigNumber(0))) {
    throw new Error("Interval total rewards should be set")
  }

  console.log(
    clc.green(`Interval total rewards: ${interval.totalRewards} KEEP`)
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
  const slaCalculator = await SLACalculator.initialize(context, interval)
  const assetsCalculator = await AssetsCalculator.initialize(context, interval)

  const operatorsParameters = []

  for (const operator of getOperators(cache)) {
    const isFraudulent = await fraudDetector.isOperatorFraudulent(operator)
    const operatorSLA = slaCalculator.calculateOperatorSLA(operator)
    const operatorAssets = await assetsCalculator.calculateOperatorAssets(
      operator
    )

    operatorsParameters.push(
      new OperatorParameters(
        operator,
        isFraudulent,
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
  operatorSLA,
  operatorAssets
) {
  ;(this.operator = operator),
    (this.isFraudulent = isFraudulent),
    (this.operatorSLA = operatorSLA),
    (this.operatorAssets = operatorAssets)
}

function OperatorSummary(operator, operatorParameters, operatorRewards) {
  ;(this.operator = operator),
    (this.isFraudulent = operatorParameters.isFraudulent),
    (this.keygenCount = operatorParameters.operatorSLA.keygenCount),
    (this.keygenFailCount = operatorParameters.operatorSLA.keygenFailCount),
    (this.keygenSLA = operatorParameters.operatorSLA.keygenSLA),
    (this.signatureCount = operatorParameters.operatorSLA.signatureCount),
    (this.signatureFailCount =
      operatorParameters.operatorSLA.signatureFailCount),
    (this.signatureSLA = operatorParameters.operatorSLA.signatureSLA),
    (this.keepStaked = operatorParameters.operatorAssets.keepStaked),
    (this.ethBonded = operatorParameters.operatorAssets.ethBonded),
    (this.ethUnbonded = operatorParameters.operatorAssets.ethUnbonded),
    (this.ethTotal = operatorParameters.operatorAssets.ethTotal),
    (this.ethScore = operatorRewards.ethScore),
    (this.boost = operatorRewards.boost),
    (this.rewardWeight = operatorRewards.rewardWeight),
    (this.totalRewards = operatorRewards.totalRewards)
}

function shortenSummaryValues(summary) {
  const shorten18Decimals = (value) =>
    value.dividedBy(new BigNumber(1e18)).toFixed(decimalPlaces)

  summary.keepStaked = shorten18Decimals(summary.keepStaked)
  summary.ethBonded = shorten18Decimals(summary.ethBonded)
  summary.ethUnbonded = shorten18Decimals(summary.ethUnbonded)
  summary.ethTotal = shorten18Decimals(summary.ethTotal)
  summary.ethScore = shorten18Decimals(summary.ethScore)
  summary.boost = summary.boost.toFixed(decimalPlaces)
  summary.rewardWeight = shorten18Decimals(summary.rewardWeight)
  summary.totalRewards = shorten18Decimals(summary.totalRewards)

  return summary
}

function writeOperatorsRewardsToFile(rewards) {
  fs.writeFileSync("./generated-rewards/rewards-input.json", JSON.stringify(rewards, null, 2))
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
