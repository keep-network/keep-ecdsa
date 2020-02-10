const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory')

// Following artifacts are expected to be copied over from previous keep-core 
// migrations.
const Registry = artifacts.require('Registry')
const TokenStaking = artifacts.require('TokenStaking')
const KeepBonding = artifacts.require('KeepBonding')

module.exports = async function () {
    const bondingValue = 100;

    const accounts = await web3.eth.getAccounts()
    const operators = [accounts[1], accounts[2], accounts[3]]

    let ecdsaKeepFactory
    let registry
    let tokenStaking
    let keepBonding

    try {
        ecdsaKeepFactory = await ECDSAKeepFactory.deployed()
        registry = await Registry.deployed()
        tokenStaking = await TokenStaking.deployed()
        keepBonding = await KeepBonding.deployed()
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

        const depositUnbondedValue = async (operator) => {
            try {
                await keepBonding.deposit(operator, { value: bondingValue })
                console.log(`deposited bonding value for operator [${operator}]`)



            } catch (err) {
                console.error(err)
                process.exit(1)
            }
        }

        for (let i = 0; i < operators.length; i++) {
            await authorizeOperator(operators[i])

            const res = await tokenStaking.eligibleStake(operators[i], operatorContract)
            console.log("res", res.toString())

            await depositUnbondedValue(operators[i])
            // TODO: Check available bonding value for factory.
        }
    } catch (err) {
        console.error('failed to initialize operators', err)
        process.exit(1)
    }

    process.exit(0)
}
