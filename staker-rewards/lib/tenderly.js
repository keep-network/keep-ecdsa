import request from "request"

export default class Tenderly {
  constructor(web3, options) {
    this.web3 = web3
    this.options = options
  }

  static initialize(web3, projectUrl, accessToken) {
    const options = {
      baseUrl:
        projectUrl ||
        "https://api.tenderly.co/api/v1/account/thesis/project/keep",
      headers: {
        "Content-Type": "application/json",
        "X-Access-Key": accessToken,
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

    const uri = `/transactions?contractId[]=eth:1:${contractAddress}&functionSelector=${functionSelector}`

    return new Promise((resolve, reject) => {
      request.get(uri, this.options, (err, res, body) => {
        if (err) {
          return reject(err)
        }

        const bodyJSON = JSON.parse(body)

        if (bodyJSON.error) {
          return reject(bodyJSON.error)
        }

        return resolve(bodyJSON)
      })
    })
  }
}
