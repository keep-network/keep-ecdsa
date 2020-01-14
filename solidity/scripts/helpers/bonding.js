const {
    fromWei,
    toWei,
    toBN
} = require("web3").utils

/** 
 * @typedef {import('web3-eth-contract')} Contract
 * @see {@link https://web3js.readthedocs.io/en/v1.2.4/web3-eth-contract.html#web3-eth-contract|web3.eth.Contract}
*/

/**
 * Deposits specific value of ether to the KeepBonding contract as value available
 * for future bonding for an operator. If there is already some value available
 * in the contract it will fill it up to desired bonding value. If currently available
 * value is greater or equal required bonding value it won't send more funds.
 * @param {Contract} keepBondingContract KeepBonding contract as web3 contract
 * @param {string} fromAccount account from which to send a deposit
 * @param {string} operator operator address to deposit bonding value to
 * @param {string} bondingValueEther required bonding value
 */
exports.depositBondingValue = async (
    keepBondingContract,
    fromAccount,
    operator,
    bondingValueEther,
) => {
    console.log(`deposit bonding value for operator [${operator}]`)

    const requiredBondingValue = toBN(toWei(bondingValueEther, "ether"))

    console.debug(`check if available bonding value for [${operator}] matches required [${fromWei(requiredBondingValue)}] ETH`)
    const currentBondingValue = toBN(await keepBondingContract.methods.availableBondingValue(operator).call())

    if (requiredBondingValue.cmp(currentBondingValue) <= 0) {
        console.log(`operator already has required bonding value [${fromWei(currentBondingValue)}] ETH`)
        return
    }

    const transferValue = requiredBondingValue.sub(currentBondingValue)

    console.debug(`depositing [${fromWei(transferValue)}] ETH as bonding value for operator [${operator}]`);
    await keepBondingContract.methods.deposit(operator).send({ from: fromAccount, value: transferValue })

    console.log(`deposited bonding value for operator [${operator}]`)
}
