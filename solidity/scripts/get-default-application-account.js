module.exports = async function () {
  const accounts = await web3.eth.getAccounts()
  // In case we are on test network, ex. Ethereum Ropsten or Celo Alfajores,
  // then we operatate on one account only. Account is specified in truffle.js
  // For local network development, the default application account is accounts[6]
  if (accounts.length == 1) {
    console.log(`${accounts[0]}`)
  } else {
    console.log(`${accounts[6]}`)
  }
  process.exit(0)
}
