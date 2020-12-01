import chai from "chai"
import testCache from "./data/test-cache.json"
import SLACalculator from "../lib/sla-calculator.js"

const { assert } = chai

const operator = "0xF1De9490Bf7298b5F350cE74332Ad7cf8d5cB181"

const mockCache = {
  getKeeps: (status) =>
    testCache.keeps.filter((keep) => !status || keep.status.name === status),
}

describe("SLA calculator", () => {
  it(
    "should return keygen SLA equal to 100% " +
      "if operator has no failed keygens",
    () => {
      // The operator has `2` keeps created within this interval. Both keeps
      // are currently active. This means `2` key generations occurred and
      // all of them completed successfully.
      const interval = {
        start: 9000,
        end: 20000,
      }

      const slaCalculator = SLACalculator.initialize(mockCache, interval)

      const operatorSLA = slaCalculator.calculateOperatorSLA(operator)

      assert.equal(operatorSLA.keygenCount, 2)
      assert.equal(operatorSLA.keygenFailCount, 0)
      assert.equal(operatorSLA.keygenSLA, 100)
    }
  )

  it(
    "should return keygen SLA less than 100% " +
      "if operator has some failed keygens",
    () => {
      // The operator has `9` keeps created within this interval. One of them
      // has been terminated due to `keygen-fail`. This means `9` key
      // generations occurred, but only `8` of them completed successfully.
      const interval = {
        start: 1000,
        end: 20000,
      }

      const slaCalculator = SLACalculator.initialize(mockCache, interval)

      const operatorSLA = slaCalculator.calculateOperatorSLA(operator)

      assert.equal(operatorSLA.keygenCount, 9)
      assert.equal(operatorSLA.keygenFailCount, 1)
      assert.equal(operatorSLA.keygenSLA, 88)
    }
  )

  it(
    "should return N/A instead of keygen SLA " +
      "if operator has no keygens at all",
    () => {
      // The operator has no keeps created within this interval. This means
      // no key generations occurred at all.
      const interval = {
        start: 20000,
        end: 30000,
      }

      const slaCalculator = SLACalculator.initialize(mockCache, interval)

      const operatorSLA = slaCalculator.calculateOperatorSLA(operator)

      assert.equal(operatorSLA.keygenCount, 0)
      assert.equal(operatorSLA.keygenFailCount, 0)
      assert.equal(operatorSLA.keygenSLA, "N/A")
    }
  )

  it(
    "should return signature SLA equal to 100% " +
      "if operator has no failed signatures",
    () => {
      // The operator has `4` keeps which changed their statuses from
      // `active` to `closed/terminated` within this interval. `2` have been
      // closed and `2` have been terminated. However, the two terminations
      // have not been caused by signature fail. This means `2` signings
      // occurred and all of them completed successfully.
      const interval = {
        start: 2000,
        end: 7000,
      }

      const slaCalculator = SLACalculator.initialize(mockCache, interval)

      const operatorSLA = slaCalculator.calculateOperatorSLA(operator)

      assert.equal(operatorSLA.signatureCount, 2)
      assert.equal(operatorSLA.signatureFailCount, 0)
      assert.equal(operatorSLA.signatureSLA, 100)
    }
  )

  it(
    "should return signature SLA less than 100% " +
      "if operator has some failed signatures",
    () => {
      // The operator has `5` keeps which changed their statuses from
      // `active` to `closed/terminated` within this interval. `2` have been
      // closed and `3` have been terminated. One of the terminations has been
      // caused by a signature fail. Two other terminations are not related
      // with signing. This means `3` signings occurred, but only `2` of them
      // completed successfully.
      const interval = {
        start: 2000,
        end: 8000,
      }

      const slaCalculator = SLACalculator.initialize(mockCache, interval)

      const operatorSLA = slaCalculator.calculateOperatorSLA(operator)

      assert.equal(operatorSLA.signatureCount, 3)
      assert.equal(operatorSLA.signatureFailCount, 1)
      assert.equal(operatorSLA.signatureSLA, 66)
    }
  )

  it(
    "should return N/A instead of signature SLA " +
      "if operator has no signatures at all",
    () => {
      // The operator has no keeps which changed their statuses from `active`
      // to `closed/terminated` within this interval. This mean no signings
      // occurred at all.
      const interval = {
        start: 20000,
        end: 30000,
      }

      const slaCalculator = SLACalculator.initialize(mockCache, interval)

      const operatorSLA = slaCalculator.calculateOperatorSLA(operator)

      assert.equal(operatorSLA.signatureCount, 0)
      assert.equal(operatorSLA.signatureFailCount, 0)
      assert.equal(operatorSLA.signatureSLA, "N/A")
    }
  )
})
