const BondedECDSAKeepFactory = artifacts.require('BondedECDSAKeepFactory')
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
        let bondedECDSAKeepFactory
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
            bondedECDSAKeepFactory = await BondedECDSAKeepFactory.deployed()
            tokenStaking = await TokenStaking.at(TokenStakingAddress)
            keepBonding = await KeepBonding.deployed()

            operatorContract = bondedECDSAKeepFactory.address
        } catch (err) {
            console.error('failed to get deployed contracts', err)
            process.exit(1)
        }

        try {
            await bondedECDSAKeepFactory.createSortitionPool(application)
            console.log(`created sortition pool for application: [${application}]`)

            sortitionPoolAddress = await bondedECDSAKeepFactory.getSortitionPool(application)
        } catch (err) {
            console.error('failed to create sortition pool', err)
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
