/**
 * Use this file to configure your truffle project. 
 * 
 * More information about configuration can be found at:
 * truffleframework.com/docs/advanced/configuration
 *
 */

require('@babel/register');
require('@babel/polyfill');

module.exports = {
  /**
   * Networks define how you connect to your ethereum client and let you set the
   * defaults web3 uses to send transactions. You can ask a truffle command to 
   * use a specific network from the command line, e.g
   *
   * $ truffle test --network <network-name>
   */

  networks: {
    development: {
      host: "127.0.0.1",
      port: 8545,
      network_id: "*",
      websockets: true, // Enable EventEmitter interface for web3
    },
    geth: {
      host: "127.0.0.1",
      port: 8546,
      network_id: "*",
      websockets: true,
      // Gas is set to a value lower as gas limit set in geth's `--targetgaslimit` 
      // option and in chain's genesis block (hex: `0x47B760`).
      // In case of exceeded block limit errors try lowering it.
      gas: "4000000",
      gasPrice: "1000000000", // same as on configured in geth (`web3.eth.gasPrice`)
    },
    keep_dev: {
      host: "localhost",
      port: 8545,
      network_id: "*",
      from: "0x0F0977c4161a371B5E5eE6a8F43Eb798cD1Ae1DB",
    },
  },

  compilers: {
    solc: {
      version: "0.5.12",
    }
  }
}
