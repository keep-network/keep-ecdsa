const BondedECDSAKeep = artifacts.require("BondedECDSAKeep")

module.exports = async function () {
  const keepAddress = process.env.KEEP_ADDRESS
  const r = process.argv[4]
  const s = process.argv[5]
  const recoveryID = process.argv[6]

  console.log(`r: ${r}`)
  console.log(`s: ${s}`)
  console.log(`recoveryID: ${recoveryID}`)

  const keep = await BondedECDSAKeep.at(keepAddress)

  try {
    console.log("submitting signature")
    const result = await keep.submitSignature(r, s, recoveryID)
    console.log(`signature submitted with transaction: ${result.tx}`)
  } catch (e) {
    console.error(e)
    process.exit(1)
  }

  process.exit(0)
}
