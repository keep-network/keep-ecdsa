import chai from "chai"

import { mockMethod } from "./helpers/blockchain.js"
import { createMockContext } from "./helpers/mock.js"

import {
  SortitionPoolAddress,
  KeepBondingAddress,
  BondedECDSAKeepFactoryAddress,
} from "./helpers/constants.js"

import Requirements from "../lib/requirements.js"
import { OperatorAuthorizations } from "../lib/requirements.js"

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
    beforeEach(async () => {
      await requirements.checkDeauthorizations()
    })

    it("discovers authorizations on interval start", async () => {
      const operator = "0xA000000000000000000000000000000000000001"

      const result = await requirements.checkAuthorizations(operator)

      expect(result).deep.equal(
        new OperatorAuthorizations(operator, true, true, false)
      )
    })

    it("finds if deauthorized during interval", async () => {
      const operator = "0xA000000000000000000000000000000000000002"

      const result = await requirements.checkAuthorizations(operator)

      expect(result).deep.equal(
        new OperatorAuthorizations(operator, true, false, true)
      )
    })

    it("finds missing authorizations on interval start", async () => {
      const operator = "0xA000000000000000000000000000000000000003"

      const result = await requirements.checkAuthorizations(operator)

      expect(result).deep.equal(
        new OperatorAuthorizations(operator, false, false, false)
      )
    })
  })
})
