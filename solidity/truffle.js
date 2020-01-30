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
    // Useful for testing. The `development` name is special - truffle uses it by default
    // if it's defined here and no other network is specified at the command line.
    // You should run a client (like ganache-cli, geth or parity) in a separate terminal
    // tab if you use this network and you must also set the `host`, `port` and `network_id`
    // options below to some value.
    //
    development: {
      host: "127.0.0.1",     // Localhost (default: none)
      port: 8545,            // Standard Ethereum port (default: none)
      network_id: "*",       // Any network (default: none)
      websockets: false,      // Enable EventEmitter interface for web3 (default: false)
    },
    keep_dev: {
      host: "localhost",
      port: 8545,
      network_id: "*",
      from: "0x0F0977c4161a371B5E5eE6a8F43Eb798cD1Ae1DB",
    },
    keep_test: {
      host: "localhost",
      port: 8545,
      network_id: "*",
      from: "0x0F0977c4161a371B5E5eE6a8F43Eb798cD1Ae1DB",
    },
    ropsten: {
      provider: function() {
        return new HDWalletProvider(process.env.CONTRACT_OWNER_ETH_ACCOUNT_PASSWORD, "https://ropsten.infura.io/v3/59fb36a36fa4474b890c13dd30038be5")
      },
      gas: 6721975,
      network_id: 3
    }
  },

  // Configure your compilers
  compilers: {
    solc: {
      version: "0.5.4",    // Fetch exact version from solc-bin (default: truffle's version)
    }
  }
}
