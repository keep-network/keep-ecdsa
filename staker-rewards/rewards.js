import clc from "cli-color"

import Context from "./lib/context.js"
import { KeepStatus, KeepTerminationCause } from "./lib/keep.js"

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

        // Step 1 of keygen SLA: get keeps opened within the given interval
        const openedKeeps = cache
            .getKeeps() // get keeps with all statuses
            .filter(keep =>
                // get keeps whose creation timestamps are within
                // the given interval
                intervalStart <= keep.creationTimestamp &&
                keep.creationTimestamp < intervalEnd
            )

        // Step 2 of keygen SLA: from keeps opened within the given interval,
        // get the ones which have been eventually terminated due to keygen fail
        const keygenFailKeeps = openedKeeps
            .filter(keep =>
                keep.status.name === KeepStatus.TERMINATED &&
                keep.status.cause === KeepTerminationCause.KEYGEN_FAIL
            )

        // Step 1 of signature SLA: get keeps closed within the given interval
        const closedKeeps = cache
            .getKeeps(KeepStatus.CLOSED) // get closed keeps
            .filter(keep => {
                // get keeps whose statuses have been changed within the
                // given interval
                return intervalStart <= keep.status.timestamp &&
                    keep.status.timestamp < intervalEnd
            })

        // Step 2 of signature SLA: get keeps terminated within the
        // given interval
        const terminatedKeeps = cache
            .getKeeps(KeepStatus.TERMINATED) // get terminated keeps
            .filter(keep => {
                // get keeps whose statuses have been changed within the
                // given interval
                return intervalStart <= keep.status.timestamp &&
                    keep.status.timestamp < intervalEnd
            })

        // Step 3 of signature SLA: Concatenate keeps closed and terminated
        // within the given interval but exclude the keeps terminated due to
        // keygen fail as they are not relevant for the signature SLA. This way
        // we obtain an array of keeps whose statuses have been changed from
        // `active` to `closed`/`terminated`. Implicitly, this means a keep
        // became not active due to one of the following causes:
        // - keep has been closed after delivering a signature successfully
        // - keep has been terminated after not delivering a signature
        // - keep has been terminated from another reason not related
        //   with the signature or keygen context
        const deactivatedKeeps = [].concat(
            closedKeeps,
            terminatedKeeps.filter(
                // get keeps which have been terminated due to causes
                // other than keygen fail
                keep => keep.status.cause !== KeepTerminationCause.KEYGEN_FAIL
            )
        )

        // Step 4 of signature SLA: from keeps terminated within the given
        // interval, get the ones which have been terminated due
        // to signature fail
        const signatureFailKeeps = terminatedKeeps
            .filter(keep =>
                // get keeps which have been terminated due to signature fail
                keep.status.cause === KeepTerminationCause.SIGNATURE_FAIL
            )

        const allOperators = new Set()
        cache.getKeeps().forEach(keep => {
            keep.members.forEach(member => allOperators.add(member))
        })

        const operatorSummary = []

        for (const operator of allOperators) {
            const keygenSummary = computeSLASummary(
                openedKeeps,
                keygenFailKeeps,
                operator
            )

            const signatureSummary = computeSLASummary(
                deactivatedKeeps,
                signatureFailKeeps,
                operator
            )

            operatorSummary.push(
                new OperatorSummary(
                    operator,
                    keygenSummary.totalCount,
                    keygenSummary.failsCount,
                    keygenSummary.SLA,
                    signatureSummary.totalCount,
                    signatureSummary.failsCount,
                    signatureSummary.SLA
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
    keygenCount,
    keygenFailCount,
    keygenSLA,
    signatureCount,
    signatureFailCount,
    signatureSLA,
) {
    this.address = address,
    this.keygenCount = keygenCount,
    this.keygenFailCount = keygenFailCount,
    this.keygenSLA = keygenSLA,
    this.signatureCount = signatureCount,
    this.signatureFailCount = signatureFailCount,
    this.signatureSLA = signatureSLA
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