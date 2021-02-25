const { web3 } = require("@openzeppelin/test-environment")

const BN = web3.utils.BN

/**
 *  Gets a list of ETH balances from a list of addresses.
 *  @param {list} members A list of addresses
 *  @return {list} The list of balances in BN notation
 */
async function getETHBalancesFromList(members) {
  async function getBalance(address) {
    const balance = await web3.eth.getBalance(address)
    return new BN(balance)
  }
  return await Promise.all(members.map((address) => getBalance(address)))
}

/**
 *  Gets a map of ETH balances from a list of addresses.
 *  @param {list} members A list of addresses
 *  @return {map} The map of balances in BN notation
 */
async function getETHBalancesMap(members) {
  async function getBalance(address) {
    const balance = await web3.eth.getBalance(address)
    return new BN(balance)
  }

  const map = {}
  for (let i = 0; i < members.length; i++) {
    const member = members[i]
    map[member] = await getBalance(member)
  }
  return map
}

/**
 *  Gets a list of ERC20 balances given a token and a list of addresses.
 *  @param {list} members A list of addresses
 *  @param {Token} token ERC20 token instance
 *  @return {list} The list of balances in BN notation
 */
async function getERC20BalancesFromList(members, token) {
  async function getBalance(address) {
    const balance = await token.balanceOf(address)
    return new BN(balance)
  }
  return await Promise.all(members.map((address) => getBalance(address)))
}

/**
 *  Adds a value to every element in a list
 *  @param {list} list A list of values
 *  @param {number} increment The amount to add to each element
 *  @return {list} The new list in BN notation
 */
function addToBalances(list, increment) {
  return list.map((element) => element.add(new BN(increment)))
}
/**
 *  Adds a value to every entry value in a map
 *  @param {map} map A map of values
 *  @param {number} increment The amount to add to each element
 *  @return {list} The new map in BN notation
 */
function addToBalancesMap(map, increment) {
  // eslint-disable-next-line guard-for-in
  for (const key in map) {
    map[key] = map[key].add(new BN(increment))
  }
  return map
}

module.exports.getETHBalancesFromList = getETHBalancesFromList
module.exports.getETHBalancesMap = getETHBalancesMap
module.exports.getERC20BalancesFromList = getERC20BalancesFromList
module.exports.addToBalances = addToBalances
module.exports.addToBalancesMap = addToBalancesMap
