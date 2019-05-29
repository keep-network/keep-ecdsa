/**
 * Use this file to configure your truffle project. 
 * 
 * More information about configuration can be found at:
 * truffleframework.com/docs/advanced/configuration
 *
 */

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
    local: {
      host: "127.0.0.1",
      port: 8545,
      network_id: "*",
      websockets: false,
    },
  },

  // Configure your compilers
  compilers: {
    solc: {
      version: "0.5.4",    // Fetch exact version from solc-bin (default: truffle's version)
    }
  }
}
