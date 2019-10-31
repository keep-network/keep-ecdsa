const BN = web3.utils.BN

/**
 *  Gets a list of ETH balances from a list of addresses. 
 *  @param members A list of addresses 
 *  @return The list of balances in BN notation
 */
async function getETHBalancesFromList(members) {
  async function getBalance(address){
    let balance =  await web3.eth.getBalance(address)
    return new BN(balance)
  }
  return await Promise.all(members.map(address => getBalance(address)))
}

/**
 *  Gets a list of ERC20 balances given a token and a list of addresses. 
 *  @param members A list of addresses 
 *  @param token ERC20 token instance
 *  @return The list of balances in BN notation
 */
async function getERC20BalancesFromList(members, token) {
  async function getBalance(address){
    let balance =  await token.balanceOf(address)
    return new BN(balance)
  }
  return await Promise.all(members.map(address => getBalance(address)))
}

/**
 *  Subtracts a value from every element in a list
 *  @param list A list of values 
 *  @param decrement The amount to subtract from each element
 *  @return The new list in BN notation
 */
function subtractBalancesFromList(list, decrement) {
  return list.map(element => element.sub(new BN(decrement)));
}

/**
 *  Adds a value to every element in a list
 *  @param list A list of values 
 *  @param increment The amount to add to each element
 *  @return The new list in BN notation
 */
function addBalancesToList(list, increment) {
  return list.map(element => element.add(new BN(increment)));
}


export {
  getETHBalancesFromList,
  getERC20BalancesFromList,
  subtractBalancesFromList,
  addBalancesToList
};