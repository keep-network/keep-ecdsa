const { contract, web3, accounts } = require("@openzeppelin/test-environment")
const { createSnapshot, restoreSnapshot } = require("./helpers/snapshot")
const { time, expectRevert } = require("@openzeppelin/test-helpers")
const SortitionPool = contract.fromArtifact("TestSortitionPool")
const StackLib = contract.fromArtifact("StackLib")

const BN = web3.utils.BN
const chai = require("chai")
chai.use(require("bn-chai")(BN))
const expect = chai.expect

describe("blah", () => {
  let pool
  beforeEach(async () => {
    stackLib = await StackLib.new()
    await SortitionPool.detectNetwork()
    SortitionPool.link(stackLib)
    pool = await SortitionPool.new()
  })

  it("bar", async () => {
    console.log(pool)
    expect(42).to.equal(42)
  })
})
