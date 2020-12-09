import chai from "chai"
import testBlockchain from "./data/test-blockchain.json"
import AssetsCalculator from "../lib/assets-calculator.js"
import BigNumber from "bignumber.js"

const { assert } = chai

console.debug = function () {}

const interval = {
  startBlock: 1000,
  endBlock: 2000,
}

const operator = "0xF1De9490Bf7298b5F350cE74332Ad7cf8d5cB181"
const SortitionPoolAddress = "0xf876aE82E3Ef9a67ad4E9eA23eFa4de2D85DA6fb"
const TokenStakingAddress = "0x186E82df0a09534537d0D8680D03b548628ab288"
const KeepFactoryAddress = "0x5CA1F949c75833432d6BC8E3cb6FB23386F63426"
const KeepBondingAddress = "0x39d2aCBCD80d80080541C6eed7e9feBb8127B2Ab"

const createMockContext = () => ({
  contracts: {},
})

const getStateAtBlock = (contractAddress, blockNumber) => {
  const block = testBlockchain.find(
    (block) => block.blockNumber === blockNumber
  )

  return (
    block &&
    block.contracts.find((contract) => contract.address === contractAddress)
  )
}

const setupContractsMock = (context) => {
  context.contracts.BondedECDSAKeepFactory = {
    deployed: () => ({
      methods: {
        getSortitionPool: () => ({
          call: () => SortitionPoolAddress,
        }),
      },
      options: {
        address: KeepFactoryAddress,
      },
    }),
  }

  context.contracts.TokenStaking = {
    deployed: () => ({
      methods: {
        activeStake: (operator, operatorContract) => ({
          call: async (_, blockNumber) => {
            const state = getStateAtBlock(TokenStakingAddress, blockNumber)

            const stake =
              state &&
              state.activeStakes.find(
                (stake) =>
                  stake.operator === operator &&
                  stake.operatorContract === operatorContract
              )

            return (stake && stake.activeStake) || "0"
          },
        }),
      },
    }),
  }

  context.contracts.KeepBonding = {
    deployed: () => ({
      methods: {
        availableUnbondedValue: (
          operator,
          bondCreator,
          sortitionPoolAddress
        ) => ({
          call: async (_, blockNumber) => {
            const state = getStateAtBlock(KeepBondingAddress, blockNumber)

            const unbonded =
              state &&
              state.unbondedValues.find(
                (value) =>
                  value.operator === operator &&
                  value.operatorContract === bondCreator &&
                  value.sortitionPool === sortitionPoolAddress
              )

            return (unbonded && unbonded.amount) || "0"
          },
        }),
      },
      getPastEvents: async (eventName, options) => {
        const state = getStateAtBlock(KeepBondingAddress, options.toBlock)

        return state && state.events.filter((event) => event.name === eventName)
      },
    }),
  }

  return context
}

describe("assets calculator", async () => {
  it("should return the right value of KEEP staked", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const assetsCalculator = await AssetsCalculator.initialize(
      mockContext,
      interval
    )

    const assets = await assetsCalculator.calculateOperatorAssets(operator)

    assert.equal(
      assets.keepStaked.isEqualTo(new BigNumber(500000).multipliedBy(1e18)),
      true
    )
  })

  it("should return the right value of ETH unbonded", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const assetsCalculator = await AssetsCalculator.initialize(
      mockContext,
      interval
    )

    const assets = await assetsCalculator.calculateOperatorAssets(operator)

    assert.equal(
      assets.ethUnbonded.isEqualTo(new BigNumber(40).multipliedBy(1e18)),
      true
    )
  })

  it("should return the right value of ETH bonded", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const assetsCalculator = await AssetsCalculator.initialize(
      mockContext,
      interval
    )

    const assets = await assetsCalculator.calculateOperatorAssets(operator)

    assert.equal(
      assets.ethBonded.isEqualTo(new BigNumber(15).multipliedBy(1e18)),
      true
    )
  })

  it("should return the right value of ETH total", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const assetsCalculator = await AssetsCalculator.initialize(
      mockContext,
      interval
    )

    const assets = await assetsCalculator.calculateOperatorAssets(operator)

    assert.equal(
      assets.ethTotal.isEqualTo(new BigNumber(55).multipliedBy(1e18)),
      true
    )
  })
})
