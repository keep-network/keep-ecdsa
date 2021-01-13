import clc from "cli-color"
import BigNumber from "bignumber.js"
import { toFormat, shorten18Decimals } from "./lib/numbers.js"

import Context from "./lib/context.js"
import FraudDetector from "./lib/fraud-detector.js"
import Requirements from "./lib/requirements.js"
import SLACalculator from "./lib/sla-calculator.js"
import AssetsCalculator from "./lib/assets-calculator.js"
import RewardsCalculator from "./lib/rewards-calculator.js"
import { getPastEvents } from "./lib/contract-helper.js"

import * as fs from "fs"

async function run() {
  // URL to the websocket endpoint of the Ethereum node.
  const ethHostname = process.env.ETH_HOSTNAME
  // Start of the interval passed as UNIX timestamp.
  const intervalStart = process.argv[2]
  // End of the interval passed as UNIX timestamp.
  const intervalEnd = process.argv[3]
  // Start block of the interval.
  const intervalStartBlock = process.argv[4]
  // End block of the interval.
  const intervalEndBlock = process.argv[5]
  // Total KEEP rewards distributed within the given interval passed as
  // 18-decimals number.
  const intervalTotalRewards = process.argv[6]
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
    startBlock: parseInt(intervalStartBlock),
    endBlock: parseInt(intervalEndBlock),
    totalRewards: new BigNumber(intervalTotalRewards),
  }

  validateIntervalTimestamps(interval)
  validateIntervalBlockspan(context, interval)
  validateIntervalTotalRewards(interval)

  if (isCacheRefreshEnabled) {
    console.log("Refreshing keeps cache...")
    await context.cache.refresh()
  }

  const operatorsRewards = await calculateOperatorsRewards(context, interval)

  if (process.env.OUTPUT_MODE === "text") {
    const rewards = {}

    for (const operatorRewards of operatorsRewards) {
      console.log(
        `${operatorRewards.operator}
           ${operatorRewards.isFraudulent} 
           ${operatorRewards.factoryAuthorizedAtStart} 
           ${operatorRewards.poolAuthorizedAtStart} 
           ${operatorRewards.poolDeauthorizedInInterval} 
           ${operatorRewards.minimumStakeAtStart} 
           ${operatorRewards.poolRequirementFulfilledAtStart} 
           ${operatorRewards.keygenCount}
           ${operatorRewards.keygenFailCount} 
           ${operatorRewards.keygenSLA} 
           ${operatorRewards.signatureCount} 
           ${operatorRewards.signatureFailCount} 
           ${operatorRewards.signatureSLA} 
           ${operatorRewards.SLAViolated} 
           ${operatorRewards.undelegated} 
           ${toFormat(operatorRewards.keepStaked)} 
           ${toFormat(operatorRewards.ethBonded)} 
           ${toFormat(operatorRewards.ethUnbonded)}
           ${toFormat(operatorRewards.ethWithdrawn)}
           ${toFormat(operatorRewards.ethTotal)} 
           ${toFormat(operatorRewards.ethScore, false, BigNumber.ROUND_DOWN)} 
           ${toFormat(operatorRewards.boost)} 
           ${toFormat(
             operatorRewards.rewardWeight,
             false,
             BigNumber.ROUND_DOWN
           )} 
           ${toFormat(
             operatorRewards.totalRewards,
             false,
             BigNumber.ROUND_DOWN
           )}
          `.replace(/\n/g, "\t")
      )

      if (!operatorRewards.totalRewards.isZero()) {
        // Amount of KEEP reward is converted to hex format to avoid problems with
        // precision during generation of a merkle tree.
        rewards[
          operatorRewards.operator
        ] = operatorRewards.totalRewards
          .integerValue(BigNumber.ROUND_DOWN)
          .toString(16)
      }
    }

    writeOperatorsRewardsToFile(rewards)
  } else {
    console.table(operatorsRewards.map(shortenSummaryValues))
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

function validateIntervalBlockspan(context, interval) {
  if (!interval.startBlock) {
    throw new Error("Invalid interval start block")
  }

  if (!interval.endBlock) {
    throw new Error("Invalid interval end block")
  }

  const isEndAfterStart = interval.endBlock > interval.startBlock
  if (!isEndAfterStart) {
    throw new Error(
      "Interval end block should be bigger than the interval start block"
    )
  }

  console.log(clc.green(`Interval start block: ${interval.startBlock}`))
  console.log(clc.green(`Interval end block: ${interval.endBlock}`))
}

async function calculateOperatorsRewards(context, interval) {
  const fraudDetector = await FraudDetector.initialize(context)
  const requirements = await Requirements.initialize(context, interval)
  const slaCalculator = await SLACalculator.initialize(context, interval)
  const assetsCalculator = await AssetsCalculator.initialize(context, interval)

  const operatorsParameters = []

  for (const operator of await getOperators(context)) {
    const isFraudulent = await fraudDetector.isOperatorFraudulent(operator)
    const operatorRequirements = await requirements.check(operator)
    const operatorSLA = slaCalculator.calculateOperatorSLA(operator)
    const operatorAssets = await assetsCalculator.calculateOperatorAssets(
      operator
    )

    operatorsParameters.push(
      new OperatorParameters(
        operator,
        isFraudulent,
        operatorRequirements,
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

async function getOperators(context) {
  console.log(`Fetching operators list...`)

  const operators = new Set()

  const events = await getPastEvents(
    context.web3,
    await context.contracts.KeepBonding.deployed(),
    "UnbondedValueDeposited",
    context.contracts.factoryDeploymentBlock
  )

  events.forEach((event) => operators.add(event.returnValues.operator))

  return operators
}

function OperatorParameters(
  operator,
  isFraudulent,
  requirements,
  operatorSLA,
  operatorAssets
) {
  ;(this.operator = operator),
    (this.isFraudulent = isFraudulent),
    (this.requirements = requirements),
    (this.operatorSLA = operatorSLA),
    (this.operatorAssets = operatorAssets)
}

function OperatorSummary(operator, operatorParameters, operatorRewards) {
  ;(this.operator = operator),
    (this.isFraudulent = operatorParameters.isFraudulent),
    (this.factoryAuthorizedAtStart =
      operatorParameters.requirements.factoryAuthorizedAtStart),
    (this.poolAuthorizedAtStart =
      operatorParameters.requirements.poolAuthorizedAtStart),
    (this.poolDeauthorizedInInterval =
      operatorParameters.requirements.poolDeauthorizedInInterval),
    (this.minimumStakeAtStart =
      operatorParameters.requirements.minimumStakeAtStart),
    (this.poolRequirementFulfilledAtStart =
      operatorParameters.requirements.poolRequirementFulfilledAtStart),
    (this.keygenCount = operatorParameters.operatorSLA.keygenCount),
    (this.keygenFailCount = operatorParameters.operatorSLA.keygenFailCount),
    (this.keygenSLA = operatorParameters.operatorSLA.keygenSLA),
    (this.signatureCount = operatorParameters.operatorSLA.signatureCount),
    (this.signatureFailCount =
      operatorParameters.operatorSLA.signatureFailCount),
    (this.signatureSLA = operatorParameters.operatorSLA.signatureSLA),
    (this.SLAViolated = operatorRewards.SLAViolated),
    (this.undelegated = operatorParameters.operatorAssets.isUndelegating),
    (this.keepStaked = operatorParameters.operatorAssets.keepStaked),
    (this.ethBonded = operatorParameters.operatorAssets.ethBonded),
    (this.ethUnbonded = operatorParameters.operatorAssets.ethUnbonded),
    (this.ethWithdrawn = operatorParameters.operatorAssets.ethWithdrawn),
    (this.ethTotal = operatorParameters.operatorAssets.ethTotal),
    (this.ethScore = operatorRewards.ethScore),
    (this.boost = operatorRewards.boost),
    (this.rewardWeight = operatorRewards.rewardWeight),
    (this.totalRewards = operatorRewards.totalRewards)
}

function shortenSummaryValues(summary) {
  summary.keepStaked = shorten18Decimals(summary.keepStaked)
  summary.ethBonded = shorten18Decimals(summary.ethBonded)
  summary.ethUnbonded = shorten18Decimals(summary.ethUnbonded)
  summary.ethWithdrawn = shorten18Decimals(summary.ethWithdrawn)
  summary.ethTotal = shorten18Decimals(summary.ethTotal)
  summary.ethScore = shorten18Decimals(summary.ethScore)
  summary.boost = toFormat(summary.boost)
  summary.rewardWeight = shorten18Decimals(summary.rewardWeight)
  summary.totalRewards = shorten18Decimals(summary.totalRewards)

  return summary
}

function writeOperatorsRewardsToFile(rewards) {
  let path = "./distributor/staker-reward-allocation.json"
  if (process.env.REWARDS_PATH) {
    path = process.env.REWARDS_PATH
  }
  fs.writeFileSync(path, JSON.stringify(rewards, null, 2))
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
