
module.exports = async function () {
    const accounts = await web3.eth.getAccounts()
    console.log(`${accounts[0]}`)
    process.exit(1)
};