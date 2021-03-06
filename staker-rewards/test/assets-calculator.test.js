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
const undelegatingOperator = "0x88Ed51f84c21Ae246c23137D090cdF441009D916"

const setupContractsMock = (context) => {
  context.contracts.BondedECDSAKeepFactory = {
    deployed: () => ({
      methods: {
        getSortitionPool: () => ({
          call: () => SortitionPoolAddress,
        }),
        getKeepOpenedTimestamp: (keepAddress) => ({
          call: mockMethod(
            BondedECDSAKeepFactoryAddress,
            "getKeepOpenedTimestamp",
            (inputs) => inputs.keepAddress === keepAddress
          ),
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
        balanceOf: (operator) => ({
          call: mockMethod(
            TokenStakingAddress,
            "balanceOf",
            (inputs) => inputs.operator === operator
          ),
        }),
        getLocks: (operator) => ({
          call: mockMethod(
            TokenStakingAddress,
            "getLocks",
            (inputs) => inputs.operator === operator
          ),
        }),
        getDelegationInfo: (operator) => ({
          call: mockMethod(
            TokenStakingAddress,
            "getDelegationInfo",
            (inputs) => inputs.operator === operator
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

  it(
    "should return the right value of KEEP staked when operator undelegated " +
      "but still has a locked stake",
    async () => {
      const mockContext = createMockContext()

      setupContractsMock(mockContext)

      const assetsCalculator = await AssetsCalculator.initialize(
        mockContext,
        interval
      )

      const assets = await assetsCalculator.calculateOperatorAssets(
        undelegatingOperator
      )

      assert.equal(
        assets.keepStaked.isEqualTo(new BigNumber(350000).multipliedBy(1e18)),
        true
      )
    }
  )

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

  it("should return the right value of ETH withdrawn", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const assetsCalculator = await AssetsCalculator.initialize(
      mockContext,
      interval
    )

    const assets = await assetsCalculator.calculateOperatorAssets(operator)

    assert.equal(
      assets.ethWithdrawn.isEqualTo(new BigNumber(15).multipliedBy(1e18)),
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
      assets.ethBonded.isEqualTo(new BigNumber(25).multipliedBy(1e18)),
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
      assets.ethTotal.isEqualTo(new BigNumber(50).multipliedBy(1e18)),
      true
    )
  })

  it(
    "should return the right value of ETH total " +
      "for an undelegating operator",
    async () => {
      const mockContext = createMockContext()

      setupContractsMock(mockContext)

      const assetsCalculator = await AssetsCalculator.initialize(
        mockContext,
        interval
      )

      const assets = await assetsCalculator.calculateOperatorAssets(
        undelegatingOperator
      )

      assert.equal(
        assets.ethTotal.isEqualTo(new BigNumber(10).multipliedBy(1e18)),
        true
      )
    }
  )

  it(
    "should return the right value of ETH total " +
      "for a deauthorized operator",
    async () => {
      const mockContext = createMockContext()

      setupContractsMock(mockContext)

      const assetsCalculator = await AssetsCalculator.initialize(
        mockContext,
        interval
      )

      const assets = await assetsCalculator.calculateOperatorAssets(operator, {
        poolDeauthorizedInInterval: true,
      })

      assert.equal(
        assets.ethTotal.isEqualTo(new BigNumber(20).multipliedBy(1e18)),
        true
      )
    }
  )

  it(
    "should return the right value of ETH total " +
      "for an operator without pool authorization at start",
    async () => {
      const mockContext = createMockContext()

      setupContractsMock(mockContext)

      const assetsCalculator = await AssetsCalculator.initialize(
        mockContext,
        interval
      )

      const assets = await assetsCalculator.calculateOperatorAssets(operator, {
        poolAuthorizedAtStart: false,
      })

      assert.equal(
        assets.ethTotal.isEqualTo(new BigNumber(20).multipliedBy(1e18)),
        true
      )
    }
  )

  it(
    "should return the right value of ETH total " +
      "for an operator without pool requirement fulfilled at start",
    async () => {
      const mockContext = createMockContext()

      setupContractsMock(mockContext)

      const assetsCalculator = await AssetsCalculator.initialize(
        mockContext,
        interval
      )

      const assets = await assetsCalculator.calculateOperatorAssets(operator, {
        poolRequirementFulfilledAtStart: false,
      })

      assert.equal(
        assets.ethTotal.isEqualTo(new BigNumber(20).multipliedBy(1e18)),
        true
      )
    }
  )
})
