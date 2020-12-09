import chai from "chai"
import BigNumber from "bignumber.js"

import { createMockContext } from "./helpers/mock.js"

import RewardsCalculator from "../lib/rewards-calculator.js"

const { assert } = chai

console.debug = function () {}
console.log = function () {}

const interval = {
  totalRewards: new BigNumber(18000000).multipliedBy(1e18), // 18M KEEP
}

const operator = "0xF1De9490Bf7298b5F350cE74332Ad7cf8d5cB181"

const createOperatorParameters = (operator, keepStaked, ethTotal) => ({
  operator: operator,
  operatorAssets: {
    keepStaked: new BigNumber(keepStaked).multipliedBy(1e18),
    ethTotal: new BigNumber(ethTotal).multipliedBy(1e18),
  },
})

const setupContractsMock = (context) => {
  context.contracts.TokenStaking = {
    deployed: () => ({
      methods: {
        minimumStake: () => ({
          call: async () => {
            // 70k KEEP
            return new BigNumber(70000).multipliedBy(1e18).toString()
          },
        }),
      },
    }),
  }

  return context
}

describe("rewards calculator", async () => {
  it(
    "should return the right value of ETH score " +
      "if ETH total is below the ETH threshold",
    async () => {
      const mockContext = createMockContext()

      setupContractsMock(mockContext)

      const rewardsCalculator = await RewardsCalculator.initialize(
        mockContext,
        interval,
        [createOperatorParameters(operator, 70000, 100)]
      )

      const rewards = rewardsCalculator.getOperatorRewards(operator)

      assert.equal(
        rewards.ethScore.isEqualTo(new BigNumber(100).multipliedBy(1e18)),
        true
      )
    }
  )

  it(
    "should return the right value of ETH score " +
      "if ETH total is above the ETH threshold",
    async () => {
      const mockContext = createMockContext()

      setupContractsMock(mockContext)

      const rewardsCalculator = await RewardsCalculator.initialize(
        mockContext,
        interval,
        [createOperatorParameters(operator, 70000, 12000)]
      )

      const rewards = rewardsCalculator.getOperatorRewards(operator)

      assert.equal(
        rewards.ethScore.isEqualTo(new BigNumber(9000).multipliedBy(1e18)),
        true
      )
    }
  )

  it("should return the right value of boost if KEEP_staked/KEEP_minStake is smaller", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const rewardsCalculator = await RewardsCalculator.initialize(
      mockContext,
      interval,
      [createOperatorParameters(operator, 70000, 100)]
    )

    const rewards = rewardsCalculator.getOperatorRewards(operator)

    assert.equal(rewards.boost.isEqualTo(new BigNumber(2)), true)
  })

  it("should return the right value of boost if KEEP_staked/KEEP_minStake is greater", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const rewardsCalculator = await RewardsCalculator.initialize(
      mockContext,
      interval,
      [createOperatorParameters(operator, 70000, 560)]
    )

    const rewards = rewardsCalculator.getOperatorRewards(operator)

    assert.equal(rewards.boost.isEqualTo(new BigNumber(1.5)), true)
  })

  it("should return the right value of reward weight", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const rewardsCalculator = await RewardsCalculator.initialize(
      mockContext,
      interval,
      [createOperatorParameters(operator, 70000, 100)]
    )

    const rewards = rewardsCalculator.getOperatorRewards(operator)

    assert.equal(
      rewards.rewardWeight.isEqualTo(new BigNumber(200).multipliedBy(1e18)),
      true
    )
  })

  it("should return the right value of total rewards", async () => {
    const mockContext = createMockContext()

    setupContractsMock(mockContext)

    const operatorParameters = createOperatorParameters(operator, 70000, 100)
    const operatorsParameters = []

    for (let i = 0; i < 10; i++) {
      operatorsParameters.push(operatorParameters)
    }

    const rewardsCalculator = await RewardsCalculator.initialize(
      mockContext,
      interval,
      operatorsParameters
    )

    const rewards = rewardsCalculator.getOperatorRewards(operator)

    assert.equal(
      rewards.totalRewards.isEqualTo(new BigNumber(1800000).multipliedBy(1e18)),
      true
    )
  })
})
