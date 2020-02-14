const ECDSAKeepFactory = artifacts.require('ECDSAKeepFactory')
const KeepBonding = artifacts.require('KeepBonding')

const TokenStaking = artifacts.require('@keep-network/keep-core/build/truffle/TokenStaking')

let { TokenStakingAddress, TBTCSystemAddress } = require('../migrations/external-contracts')

module.exports = async function () {
    try {
        const bondingValue = 1000

        const accounts = await web3.eth.getAccounts()
        const owner = accounts[0]
        const operators = [accounts[1], accounts[2], accounts[3]]
        const application = TBTCSystemAddress

        let sortitionPoolAddress
        let ecdsaKeepFactory
        let tokenStaking
        let keepBonding
        let operatorContract

        const authorizeOperator = async (operator) => {
            try {
                await tokenStaking.authorizeOperatorContract(operator, operatorContract, { from: operator })

                await keepBonding.authorizeSortitionPoolContract(operator, sortitionPoolAddress, { from: operator }) // this function should be called by authorizer but it's currently set to operator in demo.js
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

        try {
            ecdsaKeepFactory = await ECDSAKeepFactory.deployed()
            tokenStaking = await TokenStaking.at(TokenStakingAddress)
            keepBonding = await KeepBonding.deployed()

            operatorContract = ecdsaKeepFactory.address
            sortitionPoolAddress = await ecdsaKeepFactory.getSortitionPool(application)
        } catch (err) {
            console.error('failed to get deployed contracts', err)
            process.exit(1)
        }

        try {
            for (let i = 0; i < operators.length; i++) {
                await authorizeOperator(operators[i])
                await depositUnbondedValue(operators[i])
            }
        } catch (err) {
            console.error('failed to initialize operators', err)
            process.exit(1)
        }

    } catch (err) {
        console.error(err)
        process.exit(1)
    }

    process.exit(0)
}
