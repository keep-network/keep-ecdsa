import testBlockchain from "../data/test-blockchain.json"

const getStateAtBlock = (contractAddress, blockNumber) => {
  const block = testBlockchain.find(
    (block) => block.blockNumber === blockNumber
  )

  return (
    block &&
    block.contracts.find((contract) => contract.address === contractAddress)
  )
}

// `inputCheck` is expected to be a function executed to match input parameters.
// Matching is done based on the rules defined in `test-blockchain.json` file in
// each contract's `methods` property.
export function mockMethod(
  contractAddress,
  method,
  inputCheck,
  defaultOutput = ""
) {
  return async (_, blockNumber) => {
    const state = getStateAtBlock(contractAddress, blockNumber)

    const result =
      state && state.methods && state.methods[method].find(inputCheck)

    return (result && result.output) || defaultOutput
  }
}

export function mockEvents(contractAddress) {
  return async (eventName, options) => {
    const state = getStateAtBlock(contractAddress, options.toBlock)

    return state && state.events.filter((event) => event.name === eventName)
  }
}
