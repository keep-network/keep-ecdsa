const KeepBonding = artifacts.require("KeepBonding")

module.exports = async function () {
  try {
    const keepBondingAddress = process.env.KEEP_BONDING_ADDRESS
    if (!keepBondingAddress) {
      throw new Error("KEEP_BONDING_ADDRESS env not set")
    }

    const operatorPrivateKey = process.env.OPERATOR_PRIVATE_KEY
    if (!operatorPrivateKey) {
      throw new Error("OPERATOR_PRIVATE_KEY env not set")
    }

    const keepBonding = await KeepBonding.at(keepBondingAddress)

    const operatorAccount = web3.eth.accounts.privateKeyToAccount(
      operatorPrivateKey
    )

    console.log(`withdrawing bonds for operator ${operatorAccount.address}`)

    const unbondedValue = await keepBonding.unbondedValue(
      operatorAccount.address
    )
    if (unbondedValue.eq(web3.utils.toBN(0))) {
      console.log(`operator ${operatorAccount.address} has no unbonded value`)
      process.exit(0)
    }

    const tx = await keepBonding.withdraw.request(
      unbondedValue,
      operatorAccount.address,
      {
        from: operatorAccount.address,
      }
    )
    const signedTx = (await operatorAccount.signTransaction(tx)).rawTransaction
    const result = await web3.eth.sendSignedTransaction(signedTx)

    console.log(
      `bonds for operator ${operatorAccount.address} have been withdrawn ` +
        `within transaction ${result.transactionHash}`
    )
  } catch (e) {
    console.error(e)
    process.exit(1)
  }

  process.exit(0)
}
