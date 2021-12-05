require("dotenv").config()

require("@nomiclabs/hardhat-etherscan")
require("@nomiclabs/hardhat-waffle")
require("hardhat-gas-reporter")
require("solidity-coverage")

const ECDSARewardsDistributor = require("@keep-network/keep-ecdsa/artifacts/ECDSARewardsDistributor.json")

const ECDSARewardsDistributorDeployBlock = 11432833
const RewardsMerkleObjects = require("../distributor/output-merkle-objects.json")

// const multicallAggregate = require("@makerdao/multicall").aggregate;
const multicallAggregate = require("@daramir/multicall").aggregateCallData

const { utils, BigNumber } = require("ethers")
const { isAddress, getAddress, formatUnits, parseUnits } = utils

Object.filter = (obj, predicate) =>
  Object.keys(obj)
    .filter((key) => predicate(key, obj[key]))
    .reduce((res, key) => Object.assign(res, { [key]: obj[key] }), {})

async function addr(ethers, addr) {
  if (isAddress(addr)) {
    return getAddress(addr)
  }
  const accounts = await ethers.provider.listAccounts()
  if (accounts[addr] !== undefined) {
    return accounts[addr]
  }
  throw `Could not normalize address: ${addr}`
}

/**
 *
 * @param {*} ethers : Ethers plugin for Hardhad
 * @param {string} rewardsContractAddy : Address for the ECDSARewardsDistributor
 * @param {string} operatorAddress : Address of the operator that has earned the rewards
 * @returns
 */
async function findClaimedRewards(
  ethers,
  rewardsContractAddy,
  operatorAddress
) {
  const rewardsContract = new ethers.Contract(
    rewardsContractAddy,
    ECDSARewardsDistributor.abi,
    ethers.provider
  )

  const rewardsContractDeployBlock = ECDSARewardsDistributor.networks[1]
    ? (
        await ethers.provider.getTransaction(
          ECDSARewardsDistributor.networks[1].transactionHash
        )
      ).blockNumber
    : ECDSARewardsDistributorDeployBlock

  const filterClaimedByOperator = rewardsContract.filters.RewardsClaimed(
    null,
    null,
    operatorAddress
  )

  const logsResponse = await rewardsContract.queryFilter(
    filterClaimedByOperator,
    rewardsContractDeployBlock,
    "latest"
  )
  return logsResponse.map((ev) => ev.args.merkleRoot)
}

// This is a sample Hardhat task. To learn how to create your own go to
// https://hardhat.org/guides/create-task.html
task("accounts", "Prints the list of accounts", async (taskArgs, hre) => {
  const accounts = await hre.ethers.getSigners()

  for (const account of accounts) {
    console.log(account.address)
  }
})

task("balance", "Prints an account's balance")
  .addPositionalParam("account", "The account's address or hardhat index")
  .addOptionalParam(
    "tokenaddress",
    "Additional ERC-20 Token address to check balance of. Single address or comma separated list of addresses"
  )
  .setAction(async (taskArgs, { ethers }) => {
    const balance = await ethers.provider.getBalance(
      await addr(ethers, taskArgs.account)
    )
    console.log(formatUnits(balance, "ether"), "ETH")

    if (taskArgs.tokenaddress != null) {
      const tokenaddressSpl = taskArgs.tokenaddress.split(",")
      await Promise.all(
        tokenaddressSpl.map(async (tknElem) => {
          const resolvedAddress = await addr(ethers, tknElem)
          const tokenSymbol = `ERC-20 (${tknElem})`
          await initContractGetBalanceAndPrint(
            ethers,
            taskArgs.account,
            resolvedAddress,
            tokenSymbol
          )
        })
      )
    }
  })

task(
  "preview-ecdsa-rewards",
  "Check the token quantity that would be claimed by the script"
)
  .addParam(
    "op",
    "Operator address for which you will be claiming the rewards."
  )
  .addOptionalParam(
    "distaddr",
    "Address for the ECDSARewardsDistributor contract"
  )
  .setAction(async (taskArgs, { ethers }) => {
    const distributorAddress = await addr(
      ethers,
      taskArgs.distaddr ?? "0x5b9e48f8818962699fe38f5989b130cee691bbb3" // ECDSARewardsDistributor: https://etherscan.io/address/0x5b9e48f8818962699fe38f5989b130cee691bbb3
    )
    const operatorAddress = await addr(ethers, taskArgs.op)
    const ethersProvider = await ethers.provider

    const multicallConfig = {
      // rpcUrl: config.networks[config.defaultNetwork].url,
      // ethersSigner: await ethers.provider.getSigner(from),
      multicallAddress: "0x5ba1e12693dc8f9c48aad8770482f4739beed696", // Uniswap Multicall v2: etherscan.io/address/0x5ba1e12693dc8f9c48aad8770482f4739beed696
    }

    const claimedMerkleRoots = await findClaimedRewards(
      ethers,
      distributorAddress,
      operatorAddress
    )

    // unclaimedMerkleObjects as per below, contains claims from other operators
    let unclaimedMerkleObjects = Object.filter(
      RewardsMerkleObjects,
      (mrklRoot, mrklObj) => {
        const unclaimedPeriod = !claimedMerkleRoots.includes(mrklRoot)
        const periodHasOperator = mrklObj.claims[operatorAddress] != null
        return unclaimedPeriod && periodHasOperator
      }
    )
    // select only the claims for the target operator
    Object.keys(unclaimedMerkleObjects).forEach((mrklRoot) => {
      unclaimedMerkleObjects[mrklRoot] = Object.filter(
        unclaimedMerkleObjects[mrklRoot].claims,
        (claimKey, claimObj) => claimKey == operatorAddress
      )
    })

    const bigNumberSum = (previousValue, currentValue) =>
      previousValue.add(currentValue)
    let unclaimedAmounts = Object.keys(unclaimedMerkleObjects).map(
      (mrklRootKey) => {
        return unclaimedMerkleObjects[mrklRootKey][operatorAddress].amount
      }
    )
    // NOTE: usually last period is in git but not on chain
    unclaimedAmounts = unclaimedAmounts.slice(0, unclaimedAmounts.length)
    const claimAmountSum = unclaimedAmounts.reduce(
      bigNumberSum,
      BigNumber.from(0)
    )

    console.log(
      `Amount of KEEP to claim: ${formatUnits(claimAmountSum, "ether")}`
    )
    console.log(`----------------------------------------`)
    try {
      const feeData = await ethersProvider.getFeeData()
      console.log("Fee data from RPC provider:")
      console.log("maxFeePerGas:", formatUnits(feeData.maxFeePerGas, "gwei"))
      console.log("maxPriorityFeePerGas:", formatUnits(feeData.maxPriorityFeePerGas, "gwei"))
      console.log("gasPrice:", formatUnits(feeData.gasPrice, "gwei"))
      console.log(`----------------------------------------`)
      const estTotalCost = BigNumber.from(800000).mul(feeData.gasPrice)
      console.log(`Max tx cost estimate: ${formatUnits(estTotalCost, "ether")} Ξ`)
    } catch (err) {
      console.error(err.message)
    }
  })

task("claim-ecdsa-rewards", "Calls target contract multiple times")
  .addParam(
    "ww",
    "Worker wallet that will be submitting the tx. Address or account index"
  )
  .addParam(
    "op",
    "Operator address for which you will be claiming the rewards."
  )
  .addOptionalParam(
    "distaddr",
    "Address for the ECDSARewardsDistributor contract"
  )
  .addOptionalParam(
    "priorityfee",
    "The tx priority fee (EIP-1559 miner tip) in gwei. e.g. 1.4567"
  )
  .addOptionalParam(
    "maxgasfee",
    "The max tx gas fee (EIP-1559 base=tip) in gwei. e.g. 123"
  )
  .setAction(async (taskArgs, { ethers }) => {
    const distributorAddress = await addr(
      ethers,
      taskArgs.distaddr ?? "0x5b9e48f8818962699fe38f5989b130cee691bbb3" // ECDSARewardsDistributor: https://etherscan.io/address/0x5b9e48f8818962699fe38f5989b130cee691bbb3
    )
    const from = await addr(ethers, taskArgs.ww)
    const operatorAddress = await addr(ethers, taskArgs.op)
    console.log(`Normalized from address: ${from}`)
    const ethersSigner = await ethers.provider.getSigner(from)

    const multicallConfig = {
      // rpcUrl: config.networks[config.defaultNetwork].url,
      // ethersSigner: await ethers.provider.getSigner(from),
      multicallAddress: "0x5ba1e12693dc8f9c48aad8770482f4739beed696", // Uniswap Multicall v2: etherscan.io/address/0x5ba1e12693dc8f9c48aad8770482f4739beed696
    }
    let gasConfig = {}
    if (taskArgs.priorityfee)
      gasConfig.maxPriorityFeePerGas = parseUnits(
        taskArgs.priorityfee,
        "gwei"
      )
    if (taskArgs.maxgasfee)
      gasConfig.maxFeePerGas = parseUnits(taskArgs.maxgasfee, "gwei")

    const claimedMerkleRoots = await findClaimedRewards(
      ethers,
      distributorAddress,
      operatorAddress
    )

    // unclaimedMerkleObjects as per below, contains claims from other operators
    let unclaimedMerkleObjects = Object.filter(
      RewardsMerkleObjects,
      (mrklRoot, mrklObj) => {
        const unclaimedPeriod = !claimedMerkleRoots.includes(mrklRoot)
        const periodHasOperator = mrklObj.claims[operatorAddress] != null
        return unclaimedPeriod && periodHasOperator
      }
    )
    Object.keys(unclaimedMerkleObjects).forEach((mrklRoot) => {
      unclaimedMerkleObjects[mrklRoot] = Object.filter(
        unclaimedMerkleObjects[mrklRoot].claims,
        (claimKey, claimObj) => claimKey == operatorAddress
      )
    })

    // claim(bytes32 merkleRoot, uint256 index, address operator, uint256 amount, bytes32[] merkleProof)
    const calls = Object.keys(unclaimedMerkleObjects).map((mrklRootKey) => {
      return {
        target: distributorAddress,
        call: [
          "claim(bytes32,uint256,address,uint256,bytes32[])",
          mrklRootKey,
          unclaimedMerkleObjects[mrklRootKey][operatorAddress].index,
          operatorAddress,
          unclaimedMerkleObjects[mrklRootKey][operatorAddress].amount,
          unclaimedMerkleObjects[mrklRootKey][operatorAddress].proof,
        ],
      }
    })
    // NOTE: usually last period is in git but not on chain
    const slicedCalls = calls.slice(0, calls.length)

    const mcData = multicallAggregate(slicedCalls, multicallConfig)

    const txReq = {
      to: multicallConfig.multicallAddress,
      data: mcData,
      ...gasConfig,
    }
    let gasEstimate = null
    let feeData = null
    try {
      gasEstimate = await ethersSigner.estimateGas(txReq)
      console.log("Gas estimate for tx is:", gasEstimate.toString())
      feeData = await ethersSigner.getFeeData()
      console.log("Fee data from RPC provider:", feeData)
    } catch (err) {
      if (err.code == "UNPREDICTABLE_GAS_LIMIT")
        console.warn("UNPREDICTABLE_GAS_LIMIT")
      else console.error(err.message)
    }
    if (gasEstimate) {
      const gasPriceToEstimateWith = gasConfig.maxFeePerGas ?? feeData.maxFeePerGas;
      const estTotalCost = BigNumber.from(gasEstimate).mul(gasPriceToEstimateWith)
      console.log(`Max tx cost estimate: ${formatUnits(estTotalCost, "ether")} Ξ`)
      const signedTx = await ethersSigner.sendTransaction(txReq)
      // console.log("Sending transaction:", txReq)
      // console.log("Sending transaction:", signedTx)
      console.log("Waiting for tx to be mined ...")
      return signedTx.wait()
    }
  })

//
// Select the network you want to deploy/interact with here:
//
const defaultNetwork = "mainnet"

// You need to export an object to set up your config
// Go to https://hardhat.org/config/ to learn more
/**
 * @type import('hardhat/config').HardhatUserConfig
 */
module.exports = {
  defaultNetwork,

  solidity: "0.8.4",
  networks: {
    localhost: {
      url: "http://localhost:8545",
      /*
        notice no mnemonic here? it will just use account 0 of the hardhat node to deploy
        (you can put in a mnemonic here to set the deployer locally)
      */
    },
    ropsten: {
      url: process.env.ROPSTEN_URL || "",
      accounts:
        process.env.PRIVATE_KEY !== undefined ? [process.env.PRIVATE_KEY] : [],
    },
    mainnet: {
      url: `${process.env.RPC_PROVIDER_URL}`, //<---- YOUR INFURA ID! (or it won't work)
      accounts: [
        process.env.PRIVATE_KEY_W0,
      ].filter((value) => value != null),
      // accounts: {
      //   mnemonic: "abc",
      // },
    },
  },
  gasReporter: {
    enabled: process.env.REPORT_GAS !== undefined,
    currency: "USD",
  },
  etherscan: {
    apiKey: process.env.ETHERSCAN_API_KEY,
  },
}
