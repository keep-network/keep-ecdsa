import { EthereumHelpers } from "@keep-network/tbtc.js"

export class Contract {
    constructor(artifact, web3) {
        this.artifact = artifact
        this.web3 = web3
    }

    async deployed() {
        const { getDeployedContract } = EthereumHelpers

        const networkId = await this.web3.eth.net.getId()

        return getDeployedContract(this.artifact, this.web3, networkId)
    }

    async at(address) {
        const { buildContract } = EthereumHelpers

        return buildContract(this.web3, this.artifact.abi, address)
    }
}

export const getDeploymentBlockNumber = async function (artifact, web3) {
    const networkId = await web3.eth.net.getId()

    const transactionHash = artifact.networks[networkId].transactionHash

    const transaction = await web3.eth.getTransaction(transactionHash)

    return transaction.blockNumber
}

export const callWithRetry = async function(
    contractMethod,
    params,
    totalAttempts = 3
) {
    return EthereumHelpers.callWithRetry(contractMethod, params, totalAttempts)
}