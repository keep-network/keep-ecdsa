pragma solidity ^0.5.10;

contract SortitionPoolStub {
    address payable[] operators;

    function isOperatorRegistered(address operator) public view returns (bool) {
        for (uint256 i = 0; i < operators.length; i++) {
            if (operators[i] == operator) {
                return true;
            }
        }
        return false;
    }

    function insertOperator(address payable operator, uint256 weight) public {
        weight;

        operators.push(operator);
    }

    function getOperators() public view returns (address payable[] memory) {
        return operators;
    }
}
