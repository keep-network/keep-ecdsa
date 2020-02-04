const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory')

// Following artifacts are expected to be copied over from previous keep-core 
// migrations.
const Registry = artifacts.require('Registry')
const TokenStaking = artifacts.require('TokenStaking')

module.exports = async function () {
    const accounts = await web3.eth.getAccounts()
    const operators = [accounts[1], accounts[2], accounts[3]]

    let ecdsaKeepFactory
    let registry
    let tokenStaking

    try {
        ecdsaKeepFactory = await ECDSAKeepFactory.deployed()
        registry = await Registry.deployed()
        tokenStaking = await TokenStaking.deployed()
    } catch (err) {
        console.error('failed to get deployed contracts', err)
        process.exit(1)
    }

    try {
        const operatorContract = ecdsaKeepFactory.address

        const authorizeOperator = async (operator) => {
            const authorizer = operator

            try {
                await tokenStaking.authorizeOperatorContract(operator, operatorContract, { from: authorizer })
            } catch (err) {
                console.error(err)
                process.exit(1)
            }
            console.log(`authorized operator [${operator}] for factory [${operatorContract}]`)
        }

        await registry.approveOperatorContract(ecdsaKeepFactory.address)
        console.log(`approved operator contract [${ecdsaKeepFactory.address}] in registry`)

        for (let i = 0; i < operators.length; i++) {
            await authorizeOperator(operators[i])

            const res = await tokenStaking.eligibleStake(operators[i], operatorContract)
            console.log("res", res.toString())
        }
    } catch (err) {
        console.error('failed to initialize operators staking', err)
        process.exit(1)
    }

    process.exit(0)
}
