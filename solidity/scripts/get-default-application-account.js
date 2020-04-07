module.exports = async function () {
  const accounts = await web3.eth.getAccounts()
  console.log(`${accounts[6]}`)
  process.exit(0)
}
