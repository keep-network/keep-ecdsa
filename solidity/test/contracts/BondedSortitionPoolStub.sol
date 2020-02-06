pragma solidity ^0.5.10;

import "@keep-network/sortition-pools/contracts/api/IStaking.sol";
import "@keep-network/sortition-pools/contracts/api/IBonding.sol";

contract BondedSortitionPoolStub {
    address payable[] operators;

    constructor(IStaking, IBonding, uint256, uint256, address) public {}

    function isOperatorInPool(address operator) public view returns (bool) {
        for (uint256 i = 0; i < operators.length; i++) {
            if (operators[i] == operator) {
                return true;
            }
        }
        return false;
    }

    function joinPool(address payable operator) public {
        operators.push(operator);
    }

    function getOperators() public view returns (address payable[] memory) {
        return operators;
    }
}
