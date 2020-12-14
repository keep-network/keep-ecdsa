import { EthereumHelpers } from "@keep-network/tbtc.js"

const GET_PAST_EVENTS_BLOCK_INTERVAL = 2000

export class Contract {
  constructor(artifact, web3) {
    this.artifact = artifact
    this.web3 = web3
  }

  async deployed() {
    const { getDeployedContract } = EthereumHelpers

    const networkId = await this.web3.eth.net.getId()

    return getDeployedContract(this.artifact, this.web3, networkId)
  }

  async at(address) {
    const { buildContract } = EthereumHelpers

    return buildContract(this.web3, this.artifact.abi, address)
  }
}

export async function getDeploymentBlockNumber(artifact, web3) {
  const networkId = await web3.eth.net.getId()

  const transactionHash = artifact.networks[networkId].transactionHash

  const transaction = await web3.eth.getTransaction(transactionHash)

  return transaction.blockNumber
}

export async function callWithRetry(
  contractMethod,
  block = "latest",
  params = undefined,
  totalAttempts = 3
) {
  return EthereumHelpers.callWithRetry(
    contractMethod,
    params,
    totalAttempts,
    block
  )
}

export async function getPastEvents(
  web3,
  contract,
  eventName,
  fromBlock = 0,
  toBlock = "latest"
) {
  if (fromBlock < 0) {
    throw new Error(
      `FromBlock cannot be less than 0, current value: ${fromBlock}`
    )
  }

  if (toBlock !== "latest") {
    if (!Number.isInteger(toBlock) || toBlock < fromBlock) {
      throw new Error(
        `ToBlock should be \'latest'\ or an integer greater ` +
          `than FromBlock, current value: ${toBlock}`
      )
    }
  }

  return new Promise(async (resolve, reject) => {
    let resultEvents = []
    try {
      resultEvents = await contract.getPastEvents(eventName, {
        fromBlock: fromBlock,
        toBlock: toBlock,
      })
    } catch (error) {
      console.log(
        `Switching to partial events pulls; ` +
          `failed to get events in one request: [${error.message}]`
      )

      try {
        if (toBlock === "latest") {
          toBlock = await web3.eth.getBlockNumber()
        }

        let batchStartBlock = fromBlock

        while (batchStartBlock <= toBlock) {
          let batchEndBlock = batchStartBlock + GET_PAST_EVENTS_BLOCK_INTERVAL
          if (batchEndBlock > toBlock) {
            batchEndBlock = toBlock
          }
          const foundEvents = await contract.getPastEvents(eventName, {
            fromBlock: batchStartBlock,
            toBlock: batchEndBlock,
          })

          resultEvents = resultEvents.concat(foundEvents)

          batchStartBlock = batchEndBlock + 1
        }
      } catch (error) {
        return reject(error)
      }
    }

    return resolve(resultEvents)
  })
}
