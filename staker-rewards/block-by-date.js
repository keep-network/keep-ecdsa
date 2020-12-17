import clc from "cli-color"
import BlockByDate from "ethereum-block-by-date"
import Context from "./lib/context.js"

async function run() {
  const ethHostname = process.env.ETH_HOSTNAME
  const timestamp = process.argv[2]

  if (!ethHostname) {
    console.error(clc.red("Please provide ETH_HOSTNAME value"))
    return
  }

  const { web3 } = await Context.initialize(ethHostname)
  const blockByDate = new BlockByDate(web3)
  return (await blockByDate.getDate(timestamp * 1000)).block
}

run()
  .then((result) => {
    console.log(clc.green(`Block ${result}`))

    process.exit(0)
  })
  .catch((error) => {
    console.trace(clc.red("Script errored out with error: "), error)

    process.exit(1)
  })
