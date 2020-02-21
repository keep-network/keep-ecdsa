const BondedECDSAKeepFactory = artifacts.require('BondedECDSAKeepFactory')
const KeepBonding = artifacts.require('KeepBonding')

const TokenStaking = artifacts.require('@keep-network/keep-core/build/truffle/TokenStaking')
const KeepToken = artifacts.require('@keep-network/keep-core/build/truffle/KeepToken') // check if it works after removing build directory

let { TBTCSystemAddress } = require('../migrations/external-contracts')

// THIS IS A TRUFFLE SCRIPT JUST FOR DEVELOPMENT TESTING IT CANNOT BE MERGED TO MASTER WITHOUT
// MAJOR REFACTORING. 

module.exports = async function () {
    try {
        // Assuming BTC/ETH rate = 50 to cover a keep bond of 1 BTC we need to have
        // 50 ETH / 3 members = 16,67 ETH of unbonded value for each member.
        // Here we set the bonding value to bigger value so members can handle
        // multiple keeps.
        const bondingValue = web3.utils.toWei("50", "ether")

        const contractOwnerAddress = "0x923c5dbf353e99394a21aa7b67f3327ca111c67d"
        var authorizer = "0x923c5dbf353e99394a21aa7b67f3327ca111c67d"

        // TODO: REPLACE WITH ACCOUNTS READ FROM KUBE CONFIG
        const operators = [
            "0x0dac09e4bc823e6d7e49eee06a4744e8bc84d6cf",
            "0xa6ccc1824e9cdc3fb5d760ee4c4a772b6484b16f",
            "0x1599738c902e34dea8d9293efb4ea8412f0bc7fa"
        ]
        const application = TBTCSystemAddress

        let sortitionPoolAddress
        let bondedECDSAKeepFactory
        let tokenStakingContract
        let keepBonding
        let keepTokenContract
        let operatorContract


        const authorizeOperator = async (operator) => {
            try {
                await tokenStakingContract.authorizeOperatorContract(operator, operatorContract, { from: authorizer })

                await keepBonding.authorizeSortitionPoolContract(operator, sortitionPoolAddress, { from: authorizer }) // this function should be called by authorizer but it's currently set to operator in demo.js
            } catch (err) {
                console.error(err)
                process.exit(1)
            }
            console.log(`authorized operator [${operator}] for factory [${operatorContract}]`)
        }

        const depositUnbondedValue = async (operator) => {
            try {
                await keepBonding.deposit(operator, { value: bondingValue })
                console.log(`deposited ${web3.utils.fromWei(bondingValue)} ETH bonding value for operator [${operator}]`)
            } catch (err) {
                console.error(err)
                process.exit(1)
            }
        }

        try {
            bondedECDSAKeepFactory = await BondedECDSAKeepFactory.at("0x7A163bf0221f50B1eE2CF4e27615160f1934324F")
            tokenStakingContract = await TokenStaking.at("0x68cf5aE297663De7B9E46ab34DC04325216e973E")
            keepBonding = await KeepBonding.at("0x0dbebE67b2D41d27E572c7B36e95a94F8AfB43C2")
            keepTokenContract = await KeepToken.at("0x9AEeDcACe04D80d42C46c88dDbda094970e7456b")

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
                await stakeOperator(operators[i], contractOwnerAddress, authorizer)
                await authorizeOperator(operators[i])
                await depositUnbondedValue(operators[i])
            }
        } catch (err) {
            console.error('failed to initialize operators', err)
            process.exit(1)
        }


        async function stakeOperator(operatorAddress, contractOwnerAddress, authorizer) {

            let magpie = contractOwnerAddress;
            let staked = await isStaked(operatorAddress);

            /*
            We need to stake only in cases where an operator account is not already staked.  If the account
            is staked, or the client type is relay-requester we need to exit staking, albeit for different
            reasons.  In the case where the account is already staked, additional staking will fail.
            Clients of type relay-requester don't need to be staked to submit a request, they're acting more
            as a consumer of the network, rather than an operator.
            */
            if (process.env.KEEP_CLIENT_TYPE === 'relay-requester') {
                console.log('Subtype relay-requester set. No staking needed, exiting!');
                return;
            } else if (staked === true) {
                console.log('Operator account already staked, exiting!');
                return;
            } else {
                console.log(`Staking 2000000 KEEP tokens on operator account ${operatorAddress}`);
            }

            let delegation = '0x' + Buffer.concat([
                Buffer.from(magpie.substr(2), 'hex'),
                Buffer.from(operatorAddress.substr(2), 'hex'),
                Buffer.from(authorizer.substr(2), 'hex')
            ]).toString('hex');

            await keepTokenContract.approveAndCall(
                tokenStakingContract.address,
                formatAmount(20000000, 18),
                delegation, { from: contractOwnerAddress })

            console.log(`Staked!`);
        };


        async function isStaked(operatorAddress) {
            console.log('Checking if operator address is staked:');
            let stakedAmount = await tokenStakingContract.balanceOf(operatorAddress);
            return stakedAmount != 0;
        }

        function formatAmount(amount, decimals) {
            return '0x' + web3.utils.toBN(amount).mul(web3.utils.toBN(10).pow(web3.utils.toBN(decimals))).toString('hex');
        };

    } catch (err) {
        console.error(err)
        process.exit(1)
    }

    process.exit(0)
}
