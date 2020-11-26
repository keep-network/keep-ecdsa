import clc from "cli-color";
import Context from "./lib/context.js"

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

        const terminatedKeeps = cache.getKeeps("terminated")
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

run()
    .then(result => {
        console.log(clc.green("Inspection completed successfully"))
        process.exit(0)
    })
    .catch(error => {
        console.error(clc.red("Inspection errored out with error: "), error)
        process.exit(1)
    })