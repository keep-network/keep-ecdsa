import {
  KeepStatus,
  KeepTerminationCause,
  wasAskedForSignature,
} from "./keep.js"

export default class SLACalculator {
  constructor(
    openedKeeps,
    keygenFailKeeps,
    deactivatedAndAskedForSigning,
    signatureFailKeeps
  ) {
    this.openedKeeps = openedKeeps
    this.keygenFailKeeps = keygenFailKeeps
    this.deactivatedAndAskedForSigning = deactivatedAndAskedForSigning
    this.signatureFailKeeps = signatureFailKeeps
  }

  static async initialize(context, interval) {
    const { cache, contracts, web3 } = context

    const isInInterval = (timestamp) =>
      interval.start <= timestamp && timestamp < interval.end

    // Step 1 of keygen SLA: get keeps opened within the given interval
    const openedKeeps = cache
      .getKeeps() // get keeps with all statuses
      .filter((keep) =>
        // get keeps whose creation timestamps are within
        // the given interval
        isInInterval(keep.creationTimestamp)
      )

    // Step 2 of keygen SLA: from keeps opened within the given interval,
    // get the ones which have been eventually terminated due to keygen fail
    const keygenFailKeeps = openedKeeps.filter(
      (keep) =>
        keep.status.name === KeepStatus.TERMINATED &&
        keep.status.cause === KeepTerminationCause.KEYGEN_FAIL
    )

    // Step 1 of signature SLA: get keeps closed within the given interval
    const closedKeeps = cache
      .getKeeps(KeepStatus.CLOSED) // get closed keeps
      .filter((keep) =>
        // get keeps whose statuses have been changed within the
        // given interval
        isInInterval(keep.status.timestamp)
      )

    // Step 2 of signature SLA: get keeps terminated due to signature fail,
    // within the given interval
    const signatureFailKeeps = cache
      .getKeeps(KeepStatus.TERMINATED) // get terminated keeps
      .filter((keep) =>
        // get keeps whose statuses have been changed within the
        // given interval
        isInInterval(keep.status.timestamp)
      )
      .filter(
        (keep) =>
          // get keeps which have been terminated due to signature fail
          keep.status.cause === KeepTerminationCause.SIGNATURE_FAIL
      )

    // Step 3 of signature SLA: concatenate keeps closed and terminated
    // due to signature fail, within the given interval. This way
    // we obtain an array of keeps whose statuses have been changed from
    // `active` to `closed` or from `active` to `terminated` due to
    // signature fails. Implicitly, this means a keep became not active
    // due to one of the following causes:
    // - keep has been closed after delivering a signature successfully
    // - keep has been terminated after not delivering a signature
    const deactivatedKeeps = [].concat(closedKeeps, signatureFailKeeps)

    // Step 4 of signature SLA: from the deactivated keeps set, get the ones
    // which have been asked for signing at least one time.
    const deactivatedAndAskedForSigning = []
    for (const keep of deactivatedKeeps) {
      if (process.env.NODE_ENV !== "test") {
        console.log(`Checking signature requests for keep ${keep.address}`)
      }

      if (await wasAskedForSignature(keep, contracts, web3)) {
        deactivatedAndAskedForSigning.push(keep)
      }
    }

    return new SLACalculator(
      openedKeeps,
      keygenFailKeeps,
      deactivatedAndAskedForSigning,
      signatureFailKeeps
    )
  }

  calculateOperatorSLA(operator) {
    const keygen = this.calculateSLA(
      operator,
      this.openedKeeps,
      this.keygenFailKeeps
    )

    const signature = this.calculateSLA(
      operator,
      this.deactivatedAndAskedForSigning,
      this.signatureFailKeeps
    )

    return new OperatorSLA(
      operator,
      keygen.totalCount,
      keygen.failsCount,
      keygen.SLA,
      signature.totalCount,
      signature.failsCount,
      signature.SLA
    )
  }

  calculateSLA(operator, totalKeeps, failedKeeps) {
    const countOperatorKeeps = (keeps) =>
      keeps.filter((keep) => new Set(keep.members).has(operator)).length

    const totalCount = countOperatorKeeps(totalKeeps)
    const failsCount = countOperatorKeeps(failedKeeps)

    return {
      totalCount: totalCount,
      failsCount: failsCount,
      SLA:
        totalCount > 0
          ? Math.floor(100 - (failsCount * 100) / totalCount)
          : "N/A",
    }
  }
}

function OperatorSLA(
  address,
  keygenCount,
  keygenFailCount,
  keygenSLA,
  signatureCount,
  signatureFailCount,
  signatureSLA
) {
  ;(this.address = address),
    (this.keygenCount = keygenCount),
    (this.keygenFailCount = keygenFailCount),
    (this.keygenSLA = keygenSLA),
    (this.signatureCount = signatureCount),
    (this.signatureFailCount = signatureFailCount),
    (this.signatureSLA = signatureSLA)
}
