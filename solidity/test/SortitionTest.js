const {contract, web3, accounts} = require("@openzeppelin/test-environment")

const TestSortitionPool = contract.fromArtifact("TestSortitionPool")
const StackLib = contract.fromArtifact("StackLib")

const BN = web3.utils.BN
const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

describe("blah", () => {
  let pool

  beforeEach(async () => {
    stackLib = await StackLib.new()
    await TestSortitionPool.detectNetwork()
    TestSortitionPool.link("StackLib", stackLib.address)

    pool = await TestSortitionPool.new()
  })

  it.only("bar", async () => {
    for (let i = 0; i < 1000; i++) {
      await pool.joinPool(accounts[i])
    }

    const iterations = 100
    for (let num = 0; num < iterations; num++) {
      const seed = "0x" + (num).toString(16)
      const tx = await pool.selectGroup(50, seed, 1)
      console.log(tx.receipt.gasUsed)
    }
    expect(42).to.equal(42)
  })
})
