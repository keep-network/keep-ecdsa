import clc from "cli-color";
import Context from "./lib/context.js"

async function run() {
    try {
        if (process.env.DEBUG !== "on") {
            console.debug = function() {}
        }

        const context = await Context.initialize()

        const { cache } = context

        if (process.env.CACHE_REFRESH !== "off") {
            console.log("Refreshing keeps cache...")
            await cache.refresh()
        }
    } catch (error) {
        throw new Error(error)
    }
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