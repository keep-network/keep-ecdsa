const KeepBonding = artifacts.require('KeepBonding')

module.exports = async function () {
    const bondingValue = 100;

    const accounts = await web3.eth.getAccounts()
    const operators = [accounts[1], accounts[2], accounts[3]]

    let keepBonding

    try {
        keepBonding = await KeepBonding.deployed()
    } catch (err) {
        console.error('failed to get deployed contracts', err)
        process.exit(1)
    }

    try {
        const bondingDepositForOperator = async (operator) => {
            try {
                await keepBonding.deposit(operator, { value: bondingValue })
            } catch (err) {
                console.error(err)
                process.exit(1)
            }
            console.log(`deposited bonding value for operator [${operator}]`)

            // TODO: Check available bonding value for factory.
        }

        for (let i = 0; i < operators.length; i++) {
            await bondingDepositForOperator(operators[i])
        }
    } catch (err) {
        console.error('failed to initialize operators bonding', err)
        process.exit(1)
    }

    process.exit(0)
}
