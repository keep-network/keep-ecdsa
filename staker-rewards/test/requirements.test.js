import chai from "chai"

import { mockMethod } from "./helpers/blockchain.js"
import { createMockContext } from "./helpers/mock.js"

import {
  SortitionPoolAddress,
  KeepBondingAddress,
  BondedECDSAKeepFactoryAddress,
} from "./helpers/constants.js"

import Requirements from "../lib/requirements.js"

const { expect } = chai

const setupContext = (context) => {
  context.contracts.BondedECDSAKeepFactory = {
    deployed: () => ({
      methods: {
        getSortitionPool: (application) => ({
          call: async (_, blockNumber) => {
            return SortitionPoolAddress
          },
        }),
        isOperatorAuthorized: (operator) => ({
          call: mockMethod(
            BondedECDSAKeepFactoryAddress,
            "isOperatorAuthorized",
            (inputs) => inputs.operator === operator,
            false
          ),
        }),
        hasMinimumStake: (operator) => ({
          call: mockMethod(
            BondedECDSAKeepFactoryAddress,
            "hasMinimumStake",
            (inputs) => inputs.operator === operator,
            false
          ),
        }),
        isOperatorRegistered: (operator, application) => ({
          call: mockMethod(
            BondedECDSAKeepFactoryAddress,
            "isOperatorRegistered",
            (inputs) =>
              inputs.operator === operator &&
              inputs.application === application,
            false
          ),
        }),
      },
    }),
  }

  context.contracts.KeepBonding = {
    deployed: () => ({
      options: {
        address: KeepBondingAddress,
      },
      methods: {
        hasSecondaryAuthorization: (operator, sortitionPool) => ({
          call: mockMethod(
            KeepBondingAddress,
            "hasSecondaryAuthorization",
            (inputs) =>
              inputs.operator === operator &&
              inputs.poolAddress === sortitionPool,
            false
          ),
        }),
        unbondedValue: (operator) => ({
          call: mockMethod(
            KeepBondingAddress,
            "unbondedValue",
            (inputs) => inputs.operator === operator,
            "0"
          ),
        }),
      },
    }),
  }

  context.contracts.BondedSortitionPool = {
    at: (_) => ({
      methods: {
        getMinimumBondableValue: () => ({
          call: () => "100000000000000000", // 0.1 ether
        }),
      },
    }),
  }

  return context
}

describe("requirements", async () => {
  const interval = { startBlock: 1000, endBlock: 2000 }

  let mockContext
  let requirements

  before(() => {
    mockContext = createMockContext()
    setupContext(mockContext)
  })

  beforeEach(async () => {
    requirements = await Requirements.initialize(mockContext, interval)
  })

  describe("deauthorizations check", async () => {
    it("finds operators from cached transactions for the current interval", async () => {
      // The mocked transaction cache contains following deauthoriation transactions:
      //   1. made before interval start for the sortition pool
      //   2. made on interval start for the sortition pool
      //   3. made during interval for another sortition pool
      //   4. made during interval for the sortition pool
      //   5. made on interval end for the sortition pool
      //   6. made after interval end for the sortition pool
      //
      // We assume the check to find operators 2 and 4.

      expect(requirements.operatorsDeauthorizedInInterval).to.have.deep.keys([
        "0xa000000000000000000000000000000000000002",
        "0xa000000000000000000000000000000000000004",
      ])
    })

    it("does not duplicate operators entries", async () => {
      await requirements.checkDeauthorizations()
      await requirements.checkDeauthorizations()

      expect(requirements.operatorsDeauthorizedInInterval).to.have.deep.keys([
        "0xa000000000000000000000000000000000000002",
        "0xa000000000000000000000000000000000000004",
      ])
    })
  })

  describe("operator's authorizations check", async () => {
    it("discovers authorizations on interval start", async () => {
      const operator = "0xA000000000000000000000000000000000000001"

      const result = await requirements.checkAuthorizations(operator)

      expect(result).deep.equal({
        factoryAuthorizedAtStart: true,
        poolAuthorizedAtStart: true,
        poolDeauthorizedInInterval: false,
      })
    })

    it("finds if the pool was unauthorized during interval", async () => {
      const operator = "0xA000000000000000000000000000000000000002"

      const result = await requirements.checkAuthorizations(operator)

      expect(result).deep.equal({
        factoryAuthorizedAtStart: true,
        poolAuthorizedAtStart: false,
        poolDeauthorizedInInterval: true,
      })
    })

    it("finds missing authorizations on interval start", async () => {
      const operator = "0xA000000000000000000000000000000000000003"

      const result = await requirements.checkAuthorizations(operator)

      expect(result).deep.equal({
        factoryAuthorizedAtStart: false,
        poolAuthorizedAtStart: false,
        poolDeauthorizedInInterval: false,
      })
    })
  })

  describe("checks minimum stake at interval start", async () => {
    it("for operator that has minimum stake", async () => {
      const operator = "0xA000000000000000000000000000000000000001"

      const result = await requirements.checkMinimumStakeAtIntervalStart(
        operator
      )

      expect(result).to.be.true
    })

    it("for operator that has no minimum stake", async () => {
      const operator = "0xA000000000000000000000000000000000000002"

      const result = await requirements.checkMinimumStakeAtIntervalStart(
        operator
      )

      expect(result).to.be.false
    })
  })

  describe("checks unbonded value registration at interval start", async () => {
    // We defined mocks of `unbondedValue` and `isOperatorRegistered` functions
    // responses for operators used in tests:
    //  | operator | unbondedValue | isOperatorRegistered
    //  |  1       |  < minimum    |  false
    //  |  2       |  = minimum    |  true
    //  |  3       |  = minimum    |  false (*)
    //  |  4       |  > minimum    |  true
    //  |  5       |  > minimum    |  false
    //  |  6       |  2000 ether   |  false
    //
    // (*) Operator 3 is registered for another application but isn't registered
    //   for the sanctioned application we use in test context.

    it("for operator that has no minimum unbonded value", async () => {
      const operator = "0xA000000000000000000000000000000000000001"

      const result = await requirements.checkWasInPoolIfRequiredAtIntervalStart(
        operator
      )

      expect(result).to.be.true
    })

    it("for operator that has exactly minimum unbonded value and is registered", async () => {
      const operator = "0xA000000000000000000000000000000000000002"

      const result = await requirements.checkWasInPoolIfRequiredAtIntervalStart(
        operator
      )

      expect(result).to.be.true
    })

    it("for operator that has exactly minimum unbonded value and is not registered", async () => {
      const operator = "0xA000000000000000000000000000000000000003"

      const result = await requirements.checkWasInPoolIfRequiredAtIntervalStart(
        operator
      )

      expect(result).to.be.false
    })

    it("for operator that has more than minimum unbonded value and is registered", async () => {
      const operator = "0xA000000000000000000000000000000000000004"

      const result = await requirements.checkWasInPoolIfRequiredAtIntervalStart(
        operator
      )

      expect(result).to.be.true
    })

    it("for operator that has more than minimum unbonded value and is not registered", async () => {
      const operator = "0xA000000000000000000000000000000000000005"

      const result = await requirements.checkWasInPoolIfRequiredAtIntervalStart(
        operator
      )

      expect(result).to.be.false
    })

    it("for operator that has very high unbonded value and is not registered", async () => {
      const operator = "0xA000000000000000000000000000000000000006"

      const result = await requirements.checkWasInPoolIfRequiredAtIntervalStart(
        operator
      )

      expect(result).to.be.false
    })
  })
})
