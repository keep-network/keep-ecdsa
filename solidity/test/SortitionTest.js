const { contract, web3 } = require("@openzeppelin/test-environment")

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
    console.log(pool)
    expect(42).to.equal(42)
  })
})
