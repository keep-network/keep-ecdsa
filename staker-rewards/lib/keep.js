import { callWithRetry, getPastEvents } from "./contract-helper.js"

export const KeepStatus = Object.freeze({
  ACTIVE: "active",
  CLOSED: "closed",
  TERMINATED: "terminated",
})

export const KeepTerminationCause = Object.freeze({
  KEYGEN_FAIL: "keygen-fail",
  SIGNATURE_FAIL: "signature-fail",
  OTHER: "other",
})

// Returns the status of the given keep as an object with the following fields:
// - `name`: The name of the status
// - `timestamp`: UNIX timestamp of the moment when the status has been set
// - `cause` (optional): Determines why the status has been set.
//   Currently used for the `terminated` status only.
export async function getKeepStatus(keepData, contracts, web3) {
  const closeTimestamp = await getKeepCloseTime(keepData, contracts, web3)
  if (closeTimestamp) {
    return {
      name: KeepStatus.CLOSED,
      timestamp: closeTimestamp,
    }
  }

  const terminationTimestamp = await getKeepTerminationTime(
    keepData,
    contracts,
    web3
  )
  if (terminationTimestamp) {
    return {
      name: KeepStatus.TERMINATED,
      timestamp: terminationTimestamp,
      cause: await resolveKeepTerminationCause(keepData, contracts, web3),
    }
  }

  return {
    name: KeepStatus.ACTIVE,
    timestamp: (await web3.eth.getBlock(keepData.creationBlock)).timestamp,
  }
}

async function getKeepCloseTime(keepData, contracts, web3) {
  return await getKeepEventTimestamp(keepData, "KeepClosed", contracts, web3)
}

async function getKeepTerminationTime(keepData, contracts, web3) {
  return await getKeepEventTimestamp(
    keepData,
    "KeepTerminated",
    contracts,
    web3
  )
}

// Looks for a specific event for the given keep and returns the
// UNIX timestamp of the moment when the event occurred. If there
// are multiple events only the first one is taken into account.
async function getKeepEventTimestamp(keepData, eventName, contracts, web3) {
  const { address, creationBlock } = keepData

  const keepContract = await contracts.BondedECDSAKeep.at(address)

  const events = await getPastEvents(
    web3,
    keepContract,
    eventName,
    creationBlock
  )

  if (events.length > 0) {
    return (await web3.eth.getBlock(events[0].blockNumber)).timestamp
  }
}

async function resolveKeepTerminationCause(keepData, contracts, web3) {
  const { address, creationBlock } = keepData

  const keepContract = await contracts.BondedECDSAKeep.at(address)

  const publicKey = await callWithRetry(keepContract.methods.getPublicKey())
  if (!publicKey) {
    return KeepTerminationCause.KEYGEN_FAIL
  }

  const signatureRequestedEvents = (
    await getPastEvents(web3, keepContract, "SignatureRequested", creationBlock)
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

export async function wasAskedForSignature(keepData, contracts, web3) {
  const { address, creationBlock } = keepData

  const keepContract = await contracts.BondedECDSAKeep.at(address)

  const events = await getPastEvents(
    web3,
    keepContract,
    "SignatureRequested",
    creationBlock
  )

  return events.length > 0
}
