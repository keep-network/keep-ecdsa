import path from "path"
import pAll from "p-all"
import low from "lowdb"
import FileSync from "lowdb/adapters/FileSync.js"

import {
  getPastEvents,
  callWithRetry,
  KeepTerminationCause
} from "./contract-helper.js";

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
        address: event.returnValues.keepAddress,
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
        .find({ address: keep.address })
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

    await this.refreshActiveKeeps()
  }

  async fetchKeep(keepData) {
    return new Promise(async (resolve, reject) => {
      try {
        const { address, members, creationBlock } = keepData

        this.cache
            .get("keeps")
            .push({
              address: address,
              members: members,
              creationBlock: creationBlock,
              status: await this.getKeepStatus(keepData)
            })
            .write()

        console.log(
            `Successfully fetched information about keep ${address}`
        )

        return resolve()
      } catch (err) {
        return reject(err)
      }
    })
  }

  async refreshActiveKeeps() {
    const activeKeeps = this.getKeeps("active")

    console.log(`Refreshing [${activeKeeps.length}] active keeps in the cache`)

    const actions = []
    activeKeeps.forEach(keepData => {
      actions.push(() => this.refreshKeepStatus(keepData))
    })

    await pAll(actions, {concurrency: CONCURRENCY_LIMIT})
  }

  async refreshKeepStatus(keepData) {
    console.log(`Checking current status of keep ${keepData.address}`)

    const lastStatus = keepData.status
    const currentStatus = await this.getKeepStatus(keepData)

    if (lastStatus.name !== currentStatus.name) {
      console.log(`Updating current status of keep ${keepData.address}`)

      this.cache
          .get("keeps")
          .find({ address: keepData.address })
          .assign({ status: currentStatus })
          .write()
    }
  }

  async getKeepStatus(keepData) {
    const closedTimestamp = await this.getKeepEventTimestamp(
        keepData,
        "KeepClosed"
    )
    if (closedTimestamp) {
      return {
        name: "closed",
        timestamp: closedTimestamp
      }
    }

    const terminatedTimestamp = await this.getKeepEventTimestamp(
        keepData,
        "KeepTerminated"
    )
    if (terminatedTimestamp) {
      return {
        name: "terminated",
        timestamp: terminatedTimestamp,
        cause: await this.resolveKeepTerminationCause(keepData)
      }
    }

    return {
      name: "active",
      timestamp: (await this.web3.eth.getBlock(keepData.creationBlock)).timestamp
    }
  }

  async getKeepEventTimestamp(keepData, eventName) {
    const { address, creationBlock } = keepData

    const keepContract = await this.contracts.BondedECDSAKeep.at(address)

    const events = await getPastEvents(
        this.web3,
        keepContract,
        eventName,
        creationBlock
    )

    if (events.length > 0) {
      return (await this.web3.eth.getBlock(events[0].blockNumber)).timestamp
    }
  }

  async resolveKeepTerminationCause(keepData) {
    const keepContract = await this.contracts.BondedECDSAKeep.at(keepData.address)

    const publicKey = await callWithRetry(keepContract.methods.getPublicKey())
    if (!publicKey) {
      return KeepTerminationCause.KEYGEN_FAIL
    }

    const signatureRequestedEvents = (
        await getPastEvents(
            this.web3,
            keepContract,
            "SignatureRequested",
            keepData.creationBlock
        )
    ).sort((a, b) => a.blockNumber - b.blockNumber)

    const latestSignatureRequestedEvent = signatureRequestedEvents.slice(-1)[0]

    if (latestSignatureRequestedEvent) {
      const digest = latestSignatureRequestedEvent.returnValues.digest

      const isAwaitingSignature = await callWithRetry(
          keepContract.methods.isAwaitingSignature(digest)
      )

      if (isAwaitingSignature) {
        return KeepTerminationCause.SIGNATURE_FAIL
      }
    }

    return KeepTerminationCause.OTHER
  }

  getKeeps(status) {
    return this.cache
        .get("keeps")
        .filter(keep => !status || keep.status.name === status)
        .value()
  }
}
