import chai from "chai"

import { createMockContext } from "./helpers/mock.js"

import testBlockchain from "./data/test-blockchain.json"
import {
  SortitionPoolAddress,
  TokenStakingAddress,
  BondedECDSAKeepFactoryAddress,
  KeepBondingAddress,
} from "./helpers/constants.js"

import AssetsCalculator from "../lib/assets-calculator.js"

const { assert } = chai

console.debug = function () {}

const interval = {
  startBlock: 1000,
  endBlock: 2000,
}

const operator = "0xF1De9490Bf7298b5F350cE74332Ad7cf8d5cB181"

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
        address: BondedECDSAKeepFactoryAddress,
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

    assert.equal(assets.keepStaked, "500000")
  })

  it("should return the right value of ETH unbonded", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const assetsCalculator = await AssetsCalculator.initialize(
      mockContext,
      interval
    )

    const assets = await assetsCalculator.calculateOperatorAssets(operator)

    assert.equal(assets.ethUnbonded, 40)
  })

  it("should return the right value of ETH bonded", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const assetsCalculator = await AssetsCalculator.initialize(
      mockContext,
      interval
    )

    const assets = await assetsCalculator.calculateOperatorAssets(operator)

    assert.equal(assets.ethBonded, 15)
  })

  it("should return the right value of ETH total", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const assetsCalculator = await AssetsCalculator.initialize(
      mockContext,
      interval
    )

    const assets = await assetsCalculator.calculateOperatorAssets(operator)

    assert.equal(assets.ethTotal, 55)
  })
})
