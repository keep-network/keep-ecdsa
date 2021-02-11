module.exports = async function () {
  const accounts = await web3.eth.getAccounts()
  // arbitrary chosen account for default application
  if (accounts.length < 7) {
    // Assign the last account address as application account
    console.log(`${accounts[accounts.length-1]}`)
  } else {
    console.log(`${accounts[6]}`)
  }
  process.exit(0)
}
