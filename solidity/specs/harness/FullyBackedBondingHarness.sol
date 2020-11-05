
pragma solidity 0.5.17;

import "../../contracts/fully-backed/FullyBackedBonding.sol";

contract FullyBackedBondingHarness is FullyBackedBonding
{


    constructor(KeepRegistry _keepRegistry, uint256 _initializationPeriod)
    public
    FullyBackedBonding(_keepRegistry, _initializationPeriod)
    {

    }

    function balanceOf(address account) public view returns (uint256) {
        return account.balance;
    }

    function init_state() public {
    }

    function getDelegatedAuthority(address a) public view returns (address) {
        return delegatedAuthority[a];
    }

}