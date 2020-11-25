import path from "path"
import pAll from "p-all"
import low from "lowdb"
import FileSync from "lowdb/adapters/FileSync.js"

import {callWithRetry, getPastEvents} from "./contract-helper.js";

const DATA_DIR_PATH = path.resolve(process.env.DATA_DIR_PATH || "./data")
const CACHE_PATH = path.resolve(DATA_DIR_PATH, "cache.json")

// Our expectation on how deep can chain reorganization be.
const REORG_DEPTH_BLOCKS = 12
const CONCURRENCY_LIMIT = 3

export default class Cache {
  constructor(web3, contracts) {
    this.web3 = web3
    this.contracts = contracts
  }

  async initialize() {
    this.cache = low(new FileSync(CACHE_PATH))
    await this.cache
        .defaults({
          keeps: [],
          lastRefreshBlock: this.contracts.deploymentBlock,
        })
        .write()
  }

  async refresh() {
    const factory = await this.contracts.BondedECDSAKeepFactory.deployed()

    const previousRefreshBlock = this.cache.get("lastRefreshBlock").value()

    console.log(
        `Looking for keeps created since block [${previousRefreshBlock}]...`
    )

    const startBlock =
        previousRefreshBlock - REORG_DEPTH_BLOCKS > 0
            ? previousRefreshBlock - REORG_DEPTH_BLOCKS
            : 0

    const keepCreatedEvents = await getPastEvents(
        this.web3,
        factory,
        "BondedECDSAKeepCreated",
        startBlock
    )

    const keepsOnChain = []
    keepCreatedEvents.forEach((event) => {
      keepsOnChain.push({
        keepAddress: event.returnValues.keepAddress,
        members: event.returnValues.members,
        creationBlock: event.blockNumber,
      })
    })

    const cachedKeepsCount = this.cache.get("keeps").value().length

    console.log(
        `Number of keeps created since block ` +
        `[${previousRefreshBlock}]: ${keepsOnChain.length}`
    )

    console.log(`Number of keeps in the cache: ${cachedKeepsCount}`)

    const actions = []
    keepsOnChain.forEach((keep) => {
      const currentlyCached = this.cache
        .get("keeps")
        .find({ address: keep.keepAddress })
        .value()

      if (!currentlyCached) {
        actions.push(() => this.fetchKeep(keep))
      }
    })

    if (actions.length === 0) {
      console.log("Cached keeps list is up to date")
    } else {
      console.log(
          `Fetching information about [${actions.length}] new keeps...`
      )

      const results = await pAll(actions, {
        concurrency: CONCURRENCY_LIMIT,
      })

      console.log(
          `Successfully fetched information about [${results.length}] new keeps`
      )
    }

    const latestBlockNumber = keepCreatedEvents.slice(-1)[0].blockNumber
    this.cache.assign({ lastRefreshBlock: latestBlockNumber }).write()
  }

  async fetchKeep(keepData) {
    return new Promise(async (resolve, reject) => {
      try {
        const { BondedECDSAKeepFactory } = this.contracts

        const factory = await BondedECDSAKeepFactory.deployed()

        const {
          keepAddress,
          members,
          creationBlock
        } = keepData

        const creationTimestamp = await callWithRetry(
            factory.methods.getKeepOpenedTimestamp(keepAddress)
        )

        this.cache
            .get("keeps")
            .push({
              address: keepAddress,
              members: members,
              creationTimestamp: creationTimestamp,
              creationBlock: creationBlock,
            })
            .write()

        console.log(
            `Successfully fetched information about keep ${keepAddress}`
        )

        return resolve()
      } catch (err) {
        return reject(err)
      }
    })
  }
}
