import chai from "chai"

import { createMockContext } from "./helpers/mock.js"
import { mockMethod, mockEvents } from "./helpers/blockchain.js"

import {
  SortitionPoolAddress,
  TokenStakingAddress,
  BondedECDSAKeepFactoryAddress,
  KeepBondingAddress,
} from "./helpers/constants.js"

import AssetsCalculator from "../lib/assets-calculator.js"
import BigNumber from "bignumber.js"

const { assert } = chai

console.debug = function () {}

const interval = {
  startBlock: 1000,
  endBlock: 2000,
}

const operator = "0xF1De9490Bf7298b5F350cE74332Ad7cf8d5cB181"

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
          call: mockMethod(
            TokenStakingAddress,
            "activeStake",
            (inputs) =>
              inputs.operator === operator &&
              inputs.operatorContract === operatorContract
          ),
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
          call: mockMethod(
            KeepBondingAddress,
            "availableUnbondedValue",
            (inputs) =>
              inputs.operator === operator &&
              inputs.operatorContract === bondCreator &&
              inputs.sortitionPool === sortitionPoolAddress
          ),
        }),
      },
      getPastEvents: mockEvents(KeepBondingAddress),
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
