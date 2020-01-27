// Reads external contracts artifacts provided as `json` files. It is expected
// that artifacts are placed in `artifacts` directory.

const TruffleContract = require('@truffle/contract')

export function SortitionPoolFactory() {
  return TruffleContract(require("./contracts/SortitionPoolFactory.json"))
}
