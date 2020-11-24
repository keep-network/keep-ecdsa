import clc from "cli-color";
import Context from "./lib/context.js"

import { callWithRetry } from "./lib/contract-helper.js"

async function run(){
    try {
        if (process.env.DEBUG !== "on") {
            console.debug = function() {}
        }

        const context = await Context.initialize()

        const { BondedECDSAKeepFactory, BondedECDSAKeep } = context.contracts

        const factory = await BondedECDSAKeepFactory.deployed()

        const keepCount = await callWithRetry(factory.methods.getKeepCount())

        console.log(clc.green(`Keeps count: ${keepCount}`))

        const allOperators = new Set()

        for (let i = 0; i < keepCount; i++) {
            const keepAddress = await callWithRetry(factory.methods.getKeepAtIndex(i))
            const keep = await BondedECDSAKeep.at(keepAddress)
            const members = await callWithRetry(keep.methods.getMembers())

            members.forEach((member) => allOperators.add(member))
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