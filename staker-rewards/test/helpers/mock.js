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
  contracts: {},
})
