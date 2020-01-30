pragma solidity ^0.5.10;

contract SortitionPoolStub {

    address payable[] operators;

    function insertOperator(address payable operator, uint256 weight) public {
        weight;

        operators.push(operator);
    }

    function getOperators() public view returns (address payable[] memory) {
        return operators;
    }
}
