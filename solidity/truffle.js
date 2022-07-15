/**
 * Use this file to configure your truffle project.
 *
 * More information about configuration can be found at:
 * truffleframework.com/docs/advanced/configuration
 *
 */

require("@babel/register")
require("@babel/polyfill")
const HDWalletProvider = require("@truffle/hdwallet-provider")
const Kit = require("@celo/contractkit")

module.exports = {
  /**
   * Networks define how you connect to your ethereum client and let you set the
   * defaults web3 uses to send transactions. You can ask a truffle command to
   * use a specific network from the command line, e.g
   *
   * $ truffle test --network <network-name>
   */

  networks: {
    // Useful for testing. The `development` name is special - truffle uses it by default
    // if it's defined here and no other network is specified at the command line.
    // You should run a client (like ganache-cli, geth or parity) in a separate terminal
    // tab if you use this network and you must also set the `host`, `port` and `network_id`
    // options below to some value.
    //
    local: {
      host: "localhost", // Localhost (default: none)
      port: 8546, // Standard Ethereum port (default: none)
      network_id: "*", // Any network (default: none)
      websockets: true, //  Enable EventEmitter interface for web3 (default: false)
    },
    keep_dev: {
      provider: function () {
        return new HDWalletProvider({
          privateKeys: [process.env.CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY],
          providerOrUrl: "http://localhost:8545",
        })
      },
      gas: 6721975,
      network_id: 1101,
    },
    ropsten: {
      provider: function () {
        return new HDWalletProvider({
          privateKeys: [process.env.CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY],
          providerOrUrl: process.env.CHAIN_API_URL,
        })
      },
      gas: 8000000,
      network_id: 3,
      skipDryRun: true,
      networkCheckTimeout: 120000,
      timeoutBlocks: 200, // # of blocks before a deployment times out  (minimum/default: 50)
    },
    alfajores: {
      provider: function () {
        const kit = Kit.newKit(process.env.CHAIN_API_URL)
        kit.addAccount(process.env.CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY)
        return kit.web3.currentProvider
      },
      network_id: 44787,
    },
    mainnet: {
      provider: function () {
        return new HDWalletProvider({
          privateKeys: [process.env.ETH_ACCOUNT_PRIVATE_KEY],
          providerOrUrl: process.env.ETH_HOSTNAME,
        })
      },
      network_id: 1,
    },
  },
  // Configure your compilers
  compilers: {
    solc: {
      version: "0.5.17", // Fetch exact version from solc-bin (default: truffle's version)
    },
  },

  plugins: ["truffle-plugin-verify"],

  api_keys: {
    etherscan: process.env.ETHERSCAN_API_KEY,
  },
}
