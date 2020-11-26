import clc from "cli-color"

import Context from "./lib/context.js"
import { KeepStatus, KeepTerminationCause } from "./lib/contract-helper.js"

async function run() {
    try {
        const ethHostname = process.env.ETH_HOSTNAME
        const intervalStart = process.argv[2]
        const intervalEnd = process.argv[3]
        const isDebugDisabled = process.env.DEBUG !== "on"
        const isCacheRefreshEnabled = process.env.CACHE_REFRESH !== "off"

        if (!ethHostname) {
            console.error(clc.red("Wrong ETH_HOSTNAME value"))
            return
        }

        validateIntervalTimestamps(intervalStart, intervalEnd)

        if (isDebugDisabled) {
            console.debug = function() {}
        }

        const context = await Context.initialize(ethHostname)

        const { cache } = context

        if (isCacheRefreshEnabled) {
            console.log("Refreshing keeps cache...")
            await cache.refresh()
        }

        // get keeps opened within the given interval
        const openedKeeps = cache
            .getKeeps() // get keeps with all statuses
            .filter(keep =>
                // get keeps whose creation timestamps are within
                // the given interval
                intervalStart <= keep.creationTimestamp &&
                keep.creationTimestamp <= intervalEnd
            )

        // from keeps opened within the given interval, get the ones which
        // have been eventually terminated due to keygen fail
        const failedOpenedKeeps = openedKeeps
            .filter(keep =>
                keep.status.name === KeepStatus.TERMINATED &&
                keep.status.cause === KeepTerminationCause.KEYGEN_FAIL
            )

        // get keeps closed or terminated within the given interval
        // but without the ones terminated due to keygen fail as they
        // are not relevant for the closed keeps SLA
        const closedKeeps = cache
            .getKeeps() // get keeps with all statuses
            .filter(keep => {
                // get keeps whose statuses have been changed within the
                // given interval
                return intervalStart <= keep.status.timestamp &&
                    keep.status.timestamp <= intervalEnd
            })
            .filter(keep =>
                // get keeps which are currently in the state `closed` or
                // `terminated` due to causes other than keygen fail
                keep.status.name === KeepStatus.CLOSED ||
                (keep.status.name === KeepStatus.TERMINATED &&
                    keep.status.cause !== KeepTerminationCause.KEYGEN_FAIL)
            )

        // from keeps closed or terminated (causes other than keygen fail)
        // within the given interval, get the ones which have been terminated
        // due to signature fail
        const failedClosedKeeps = closedKeeps
            .filter(keep =>
                keep.status.name === KeepStatus.TERMINATED &&
                keep.status.cause === KeepTerminationCause.SIGNATURE_FAIL
            )

        const allOperators = new Set()
        cache.getKeeps().forEach(keep => {
            keep.members.forEach(member => allOperators.add(member))
        })

        const operatorSummary = []

        for (const operator of allOperators) {
            const keepOpeningsSummary = computeSLASummary(
                openedKeeps,
                failedOpenedKeeps,
                operator
            )

            const keepClosuresSummary = computeSLASummary(
                closedKeeps,
                failedClosedKeeps,
                operator
            )

            operatorSummary.push(
                new OperatorSummary(
                    operator,
                    keepOpeningsSummary.totalCount,
                    keepOpeningsSummary.failsCount,
                    keepOpeningsSummary.SLA,
                    keepClosuresSummary.totalCount,
                    keepClosuresSummary.failsCount,
                    keepClosuresSummary.SLA
                )
            )
        }

        console.table(operatorSummary)
    } catch (error) {
        throw new Error(error)
    }
}

function validateIntervalTimestamps(start, end) {
    const startDate = new Date(start * 1000)
    const endDate = new Date(end * 1000)

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

    console.log(clc.green(`Interval start: ${startDate.toISOString()}`))
    console.log(clc.green(`Interval end: ${endDate.toISOString()}`))
}

function computeSLASummary(totalKeeps, failedKeeps, operator) {
    const countOperatorKeeps = (keeps) =>
        keeps.filter(keep => new Set(keep.members).has(operator)).length

    const totalCount = countOperatorKeeps(totalKeeps)
    const failsCount = countOperatorKeeps(failedKeeps)

    return {
        totalCount: totalCount,
        failsCount: failsCount,
        SLA: (totalCount > 0) ?
            Math.floor(100 - ((failsCount * 100) / totalCount)) : "N/A"
    }
}

function OperatorSummary(
    address,
    keepOpenings,
    keepOpeningsFails,
    keepOpeningsSLA,
    keepClosures,
    keepClosuresFails,
    keepClosuresSLA,
) {
    this.address = address,
    this.keepOpenings = keepOpenings,
    this.keepOpeningsFails = keepOpeningsFails,
    this.keepOpeningsSLA = keepOpeningsSLA,
    this.keepClosures = keepClosures,
    this.keepClosuresFails = keepClosuresFails,
    this.keepClosuresSLA = keepClosuresSLA
}

run()
    .then(result => {
        console.log(clc.green("Inspection completed successfully"))
        process.exit(0)
    })
    .catch(error => {
        console.error(clc.red("Inspection errored out with error: "), error)
        process.exit(1)
    })