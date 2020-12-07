import BN from "bn.js"

import testCache from "../data/test-cache.json"
import transactionsCache from "../data/test-transactions.json"

export const createMockContext = () => ({
  cache: {
    getKeeps: (status) =>
      testCache.keeps.filter((keep) => !status || keep.status.name === status),
    getTransactionFunctionCalls: (to, method) =>
      transactionsCache.transactions.filter(
        (tx) => tx.to.toLowerCase() === to.toLowerCase() && tx.method === method
      ),
  },
  web3: {
    utils: {
      toBN: (value) => new BN(value),
      fromWei: (value) =>
        new BN(value).div(new BN("1000000000000000000")).toString(),
    },
  },
  contracts: {},
})
