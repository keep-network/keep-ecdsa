const HDWalletProvider = require("@truffle/hdwallet-provider")

// HDWalletProvider requires us to provide a valid mnemonic. Since we are just
// reading, we provide a dummy mnemonic that has no ether on it.
const dummyMnemonic =
    "6892a90dab700bab8cee21cef939461f41f48b91c271120aa8b10cd3d9dd86dc"

module.exports = {
    networks: {
        ropsten: {
            provider: function () {
                return new HDWalletProvider(dummyMnemonic, process.env.ETH_HOSTNAME)
            },
            network_id: 3,
        },
        mainnet: {
            provider: function () {
                return new HDWalletProvider(dummyMnemonic, process.env.ETH_HOSTNAME)
            },
            network_id: 1,
        },
    },
}