const BN = web3.utils.BN

/**
 *  Gets a list of ETH balances from a list of addresses. 
 *  @param members A list of addresses 
 *  @return The list of balances in BN notation
 */
async function getETHBalancesFromList(members) {
  async function getBalance(address) {
    let balance = await web3.eth.getBalance(address)
    return new BN(balance)
  }
  return await Promise.all(members.map(address => getBalance(address)))
}

/**
 *  Gets a map of ETH balances from a list of addresses.
 *  @param members A list of addresses
 *  @return The map of balances in BN notation
 */
async function getETHBalancesMap(members) {
  async function getBalance(address) {
    let balance = await web3.eth.getBalance(address)
    return new BN(balance)
  }

  let map = {}
  for (let i = 0; i < members.length; i++) {
    const member = members[i]
    map[member] = await getBalance(member)
  }
  return map
}

/**
 *  Gets a list of ERC20 balances given a token and a list of addresses. 
 *  @param members A list of addresses 
 *  @param token ERC20 token instance
 *  @return The list of balances in BN notation
 */
async function getERC20BalancesFromList(members, token) {
  async function getBalance(address) {
    let balance = await token.balanceOf(address)
    return new BN(balance)
  }
  return await Promise.all(members.map(address => getBalance(address)))
}

/**
 *  Adds a value to every element in a list
 *  @param list A list of values 
 *  @param increment The amount to add to each element
 *  @return The new list in BN notation
 */
function addToBalances(list, increment) {
  return list.map(element => element.add(new BN(increment)));
}
/**
 *  Adds a value to every entry value in a map
 *  @param map A map of values
 *  @param increment The amount to add to each element
 *  @return The new map in BN notation
 */
function addToBalancesMap(map, increment) {
  for (let key in map) {
    map[key] = map[key].add(new BN(increment))
  }
  return map
}

export {
  getETHBalancesFromList,
  getETHBalancesMap,
  getERC20BalancesFromList,
  addToBalances,
  addToBalancesMap
};
