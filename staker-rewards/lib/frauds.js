import clc from "cli-color"

export async function isOperatorKeepECDSAFraudulent(context, operator) {
  const { TokenStaking, factoryDeploymentBlock } = context.contracts

  const tokenStaking = await TokenStaking.deployed()

  // Get all slashing events for the operator.
  const events = await tokenStaking.getPastEvents("TokensSlashed", {
    fromBlock: factoryDeploymentBlock,
    toBlock: "latest",
    filter: { operator: operator },
  })

  if (events.length > 0) {
    console.log(
      clc.yellow(
        `found slashing event for operator [${operator}];` +
          ` please double check correctness of a calculation result`
      )
    )
  } else {
    console.debug(
      `found [${events.length}] slashing events for operator [${operator}]`
    )
  }

  let result
  for (let i = 0; i < events.length; i++) {
    // We need to determine if a slashing event had originated from Keep ECDSA
    // fraud report. A fraud can be reported by calling `Deposit` contract or
    // directly `BondedECDSAKeep` contract. We are checking if transaction
    // that emitted a `TokensSlashed` event had been made to one of the mentioned
    // contracts.

    const txHash = events[i].transactionHash

    console.debug(`checking event from transaction: ${txHash}`)

    result = await isKeepECDSATransaction(context, txHash)

    if (result) {
      break
    }
  }

  return result
}

async function isKeepECDSATransaction(context, transactionHash) {
  const { web3 } = context
  const {
    BondedECDSAKeepFactory,
    TBTCSystem,
    factoryDeploymentBlock,
  } = context.contracts

  const tx = await web3.eth.getTransaction(transactionHash)

  // Check if the called contract is a BondedECDSAKeep.
  const bondedECDSAKeepFactory = await BondedECDSAKeepFactory.deployed()

  const keepOpenedTimestamp = await bondedECDSAKeepFactory.methods
    .getKeepOpenedTimestamp(tx.to)
    .call()

  const isBondedECDSAKeep = keepOpenedTimestamp > 0

  // Check if the called contract is a Deposit.
  const tbtcSystem = await TBTCSystem.deployed()

  const depositCreatedEvents = await tbtcSystem.getPastEvents("Created", {
    fromBlock: factoryDeploymentBlock,
    toBlock: "latest",
    filter: { _depositContractAddress: tx.to },
  })

  const isDeposit = depositCreatedEvents.length > 0

  return isBondedECDSAKeep || isDeposit
}
