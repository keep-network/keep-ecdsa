import path from "path"
import pAll from "p-all"
import low from "lowdb"
import FileSync from "lowdb/adapters/FileSync.js"

import { getPastEvents, callWithRetry } from "./contract-helper.js"
import { KeepStatus, getKeepStatus } from "./keep.js"

const DATA_DIR_PATH = path.resolve(process.env.DATA_DIR_PATH || "./data")
const KEEPS_CACHE_PATH = path.resolve(DATA_DIR_PATH, "cache-keeps.json")
const TX_CACHE_PATH = path.resolve(DATA_DIR_PATH, "cache-transactions.json")

// Our expectation on how deep can chain reorganization be. We need this
// parameter because the previous cache refresh could store data that
// are no longer valid due to a chain reorganization. To overcome this
// problem we lookup `REORG_DEPTH_BLOCKS` earlier than the last refresh
// block when starting the cache refresh.
const REORG_DEPTH_BLOCKS = 12

const CONCURRENCY_LIMIT = 3

export default class Cache {
  constructor(web3, contracts) {
    this.web3 = web3
    this.contracts = contracts
  }

  async initialize() {
    this.keepsCache = low(new FileSync(KEEPS_CACHE_PATH))
    this.transactionsCache = low(new FileSync(TX_CACHE_PATH))

    await this.keepsCache
      .defaults({
        keeps: [],
        lastRefreshBlock: this.contracts.factoryDeploymentBlock,
      })
      .write()

    await this.transactionsCache.defaults({ transactions: [] }).write()
  }

  async refresh() {
    await this.fetchNewKeeps()
    await this.refreshActiveKeeps()
  }

  async fetchNewKeeps() {
    const factory = await this.contracts.BondedECDSAKeepFactory.deployed()

    const previousRefreshBlock = this.keepsCache.get("lastRefreshBlock").value()

    const startBlock =
      previousRefreshBlock - REORG_DEPTH_BLOCKS > 0
        ? previousRefreshBlock - REORG_DEPTH_BLOCKS
        : 0

    console.log(`Looking for keeps created since block [${startBlock}]...`)

    const keepCreatedEvents = await getPastEvents(
      this.web3,
      factory,
      "BondedECDSAKeepCreated",
      startBlock
    )

    const newKeeps = []
    keepCreatedEvents.forEach((event) => {
      newKeeps.push({
        address: event.returnValues.keepAddress,
        members: event.returnValues.members,
        creationBlock: event.blockNumber,
      })
    })

    const cachedKeepsCount = this.keepsCache.get("keeps").value().length

    console.log(
      `Number of keeps created since block ` +
        `[${startBlock}]: ${newKeeps.length}`
    )

    console.log(`Number of keeps in the cache: ${cachedKeepsCount}`)

    const actions = []
    newKeeps.forEach((keep) => {
      const isCached = this.keepsCache
        .get("keeps")
        .find({ address: keep.address })
        .value()

      if (!isCached) {
        actions.push(() => this.fetchFullKeepData(keep))
      }
    })

    if (actions.length === 0) {
      console.log("Cached keeps list is up to date")
    } else {
      console.log(`Fetching information about [${actions.length}] new keeps...`)

      const results = await pAll(actions, {
        concurrency: CONCURRENCY_LIMIT,
      })

      console.log(
        `Successfully fetched information about [${results.length}] new keeps`
      )
    }

    const latestBlockNumber = keepCreatedEvents.slice(-1)[0].blockNumber
    this.keepsCache.assign({ lastRefreshBlock: latestBlockNumber }).write()
  }

  async fetchFullKeepData(keepData) {
    return new Promise(async (resolve, reject) => {
      try {
        const { address, members, creationBlock } = keepData
        const keepContract = await this.contracts.BondedECDSAKeep.at(address)

        const creationTimestamp = await callWithRetry(
          keepContract.methods.getOpenedTimestamp()
        )

        this.keepsCache
          .get("keeps")
          .push({
            address: address,
            members: members,
            creationBlock: creationBlock,
            creationTimestamp: parseInt(creationTimestamp),
            status: await getKeepStatus(keepData, this.contracts, this.web3),
          })
          .write()

        console.log(`Successfully fetched information about keep ${address}`)

        return resolve()
      } catch (err) {
        return reject(err)
      }
    })
  }

  async refreshActiveKeeps() {
    const activeKeeps = this.getKeeps(KeepStatus.ACTIVE)

    console.log(`Refreshing [${activeKeeps.length}] active keeps in the cache`)

    const actions = []
    activeKeeps.forEach((keepData) => {
      actions.push(() => this.refreshKeepStatus(keepData))
    })

    await pAll(actions, { concurrency: CONCURRENCY_LIMIT })
  }

  async refreshKeepStatus(keepData) {
    console.log(`Checking current status of keep ${keepData.address}`)

    const lastStatus = keepData.status
    const currentStatus = await getKeepStatus(
      keepData,
      this.contracts,
      this.web3
    )

    if (lastStatus.name !== currentStatus.name) {
      console.log(
        `Updating current status of keep ${keepData.address} ` +
          `from [${lastStatus.name}] to [${currentStatus.name}]`
      )

      this.keepsCache
        .get("keeps")
        .find({ address: keepData.address })
        .assign({ status: currentStatus })
        .write()
    }
  }

  async storeTransactions(transactions) {
    transactions.forEach((transaction) => {
      const isCached = this.transactionsCache
        .get("transactions")
        .find({ hash: transaction.hash })
        .value()

      if (!isCached) {
        this.transactionsCache.get("transactions").push(transaction).write()
      }
    })
  }

  getKeeps(status) {
    return this.keepsCache
      .get("keeps")
      .filter((keep) => !status || keep.status.name === status)
      .value()
  }

  getTransactionFunctionCalls(to, method) {
    return this.transactionsCache
      .get("transactions")
      .filter(
        (tx) => tx.to.toLowerCase() === to.toLowerCase() && tx.method === method
      )
      .value()
  }
}
