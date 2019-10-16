const BN = require('bn.js')

/**
 *  gets a list of ETH balances from a list of addresses. 
 *  @param members List of addreses 
 *  @return list of balances in BN notation
 */
async function getEthBalancesFromList(members) {
  async function getBalance(address){
    let balance =  await web3.eth.getBalance(address)
    return new BN(balance)
  }
  return await Promise.all(members.map(address => getBalance(address)))
}

/**
 *  gets a list of ERC20 balances given a token and a list of addresses. 
 *  @param members List of addreses 
 *  @param tokens ERC20 token instance
 *  @return list of balances in BN notation
 */
async function getERC20BalancesFromList(members, token) {
  async function getBalance(address){
    let balance =  await token.balanceOf(address)
    return new BN(balance)
  }
  return await Promise.all(members.map(address => getBalance(address)))
}

/**
 *  subtracts a value from every element in a list
 *  @param list List of values 
 *  @param decrement amount to subtract from each element
 *  @return new list in BN notation
 */
function subtractBalancesFromList(list, decrement) {
  return list.map(element => element.sub(new BN(decrement)));
}

export {getEthBalancesFromList, getERC20BalancesFromList, subtractBalancesFromList};