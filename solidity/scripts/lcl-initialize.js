const BondedECDSAKeepFactory = artifacts.require("BondedECDSAKeepFactory")
const KeepBonding = artifacts.require("KeepBonding")

const TokenStaking = artifacts.require(
  "@keep-network/keep-core/build/truffle/TokenStaking"
)

const {
  TokenStakingAddress,
  TBTCSystemAddress,
} = require("../migrations/external-contracts")

module.exports = async function () {
  try {
    // Assuming BTC/ETH rate = 50 to cover a keep bond of 1 BTC we need to have
    // 50 ETH / 3 members = 16,67 ETH of unbonded value for each member.
    // Here we set the bonding value to bigger value so members can handle
    // multiple keeps.
    const bondingValue = web3.utils.toWei("50", "ether")

    const accounts = await web3.eth.getAccounts()
    const operators = [accounts[1], accounts[2], accounts[3], accounts[4]]
    const application = TBTCSystemAddress

    let sortitionPoolAddress
    let bondedECDSAKeepFactory
    let tokenStaking
    let keepBonding
    let operatorContract

    const authorizeOperator = async (operator) => {
      try {
        await tokenStaking.authorizeOperatorContract(
          operator,
          operatorContract,
          { from: operator }
        )

        await keepBonding.authorizeSortitionPoolContract(
          operator,
          sortitionPoolAddress,
          { from: operator }
        ) // this function should be called by authorizer but it's currently set to operator in demo.js
      } catch (err) {
        console.error(err)
        process.exit(1)
      }
      console.log(
        `authorized operator [${operator}] for factory [${operatorContract}]`
      )
    }

    const depositUnbondedValue = async (operator) => {
      try {
        await keepBonding.deposit(operator, { value: bondingValue })
        console.log(
          `deposited ${web3.utils.fromWei(
            bondingValue
          )} ETH bonding value for operator [${operator}]`
        )
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
      console.error("failed to get deployed contracts", err)
      process.exit(1)
    }

    try {
      console.log(`getting sortition pool for application [${application}]`)

      sortitionPoolAddress = await bondedECDSAKeepFactory.getSortitionPool(
        application
      )
    } catch (err) {
      console.error("failed to get sortition pool", err)
    }

    // FIXME: Workaround for https://github.com/ethereum/go-ethereum/pull/21083 - instead of failing,
    // contract call returns address 0x0. Once we are able to update to the most recent abigen
    // (see https://github.com/keep-network/keep-common/issues/45), the additional case for 0x0 address
    // should be removed.
    if (
      !sortitionPoolAddress ||
      sortitionPoolAddress === "0x0000000000000000000000000000000000000000"
    ) {
      try {
        console.error(
          `no sortition pool for application: [${application}]; creating a new one`
        )

        await bondedECDSAKeepFactory.createSortitionPool(application)

        console.log(`created sortition pool for application: [${application}]`)

        sortitionPoolAddress = await bondedECDSAKeepFactory.getSortitionPool(
          application
        )
      } catch (err) {
        console.error("failed to create sortition pool", err)
        process.exit(1)
      }
    }

    console.log(
      `sortition pool for application [${application}] is [${sortitionPoolAddress}]`
    )

    try {
      for (let i = 0; i < operators.length; i++) {
        await authorizeOperator(operators[i])
        await depositUnbondedValue(operators[i])
      }
    } catch (err) {
      console.error("failed to initialize operators", err)
      process.exit(1)
    }
  } catch (err) {
    console.error(err)
    process.exit(1)
  }

  process.exit(0)
}
