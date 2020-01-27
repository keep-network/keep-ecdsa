// Reads external contracts artifacts provided as `json` files. It is expected
// that artifacts are placed in `artifacts` directory.

const TruffleContract = require('@truffle/contract')

const SortitionPoolFactory = TruffleContract(require("./contracts/SortitionPoolFactory.json"))

module.exports = {
  SortitionPoolFactory,
}
