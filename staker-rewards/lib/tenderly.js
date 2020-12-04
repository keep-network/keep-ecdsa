import request from "request"

export default class Tenderly {
  constructor(web3, options) {
    this.web3 = web3
    this.options = options
  }

  static initialize(web3, apiAccessKey) {
    const options = {
      headers: {
        "Content-Type": "application/json",
        "X-Access-Key": apiAccessKey,
      },
    }

    return new Tenderly(web3, options)
  }

  async getFunctionCalls(contractAddress, functionSignature) {
    console.debug(
      `Looking for calls to contract [${contractAddress}] function [${functionSignature}]`
    )

    const functionSelector = this.web3.eth.abi.encodeFunctionSignature(
      functionSignature
    )

    return new Promise((resolve, reject) => {
      request.get(
        `https://api.tenderly.co/api/v1/account/thesis/project/keep/transactions?contractId[]=eth:1:${contractAddress}&functionSelector=${functionSelector}`,
        this.options,
        (err, res, body) => {
          if (err) {
            return reject(err)
          }

          return resolve(JSON.parse(body))
        }
      )
    })
  }
}
