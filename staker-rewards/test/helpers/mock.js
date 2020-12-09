import testCache from "../data/test-cache.json"
import transactionsCache from "../data/test-transactions.json"
import { SANCTIONED_APPLICATION_ADDRESS } from "../../lib/context.js"

export const createMockContext = () => ({
  cache: {
    getKeeps: (status) =>
      testCache.keeps.filter((keep) => !status || keep.status.name === status),
    getTransactionFunctionCalls: (to, method) =>
      transactionsCache.transactions.filter(
        (tx) => tx.to.toLowerCase() === to.toLowerCase() && tx.method === method
      ),
  },
  contracts: { sanctionedApplicationAddress: SANCTIONED_APPLICATION_ADDRESS },
})
